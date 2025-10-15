package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	defaultTimeout            = 60 * time.Second
	defaultISOStorage         = "local"
	defaultSchematicFile      = ".schematic_id"
	defaultTalosSchematicPath = "talos-factory-schematic.yaml"
	defaultTemplateJSONPath   = "template.json"
)

// Config describes the minimum information required to talk to the Proxmox and Talos APIs.
type Config struct {
	URL                string        // Proxmox host without scheme
	TokenID            string        // PVEAPIToken token ID
	Secret             string        // PVEAPIToken secret
	Node               string        // Proxmox node name
	ISOStorage         string        // Proxmox storage target for ISO uploads
	SchematicFile      string        // Path to cached schematic id file
	TalosSchematicPath string        // Talos factory schematic YAML path
	TemplateJSONPath   string        // VM template payload
	SkipTLSVerify      bool          // Whether to skip TLS verification for Proxmox API calls
	HTTPClient         *http.Client  // Optional custom HTTP client for Proxmox
	FactoryClient      *http.Client  // Optional custom HTTP client for Talos factory
	Timeout            time.Duration // Optional override for HTTP timeouts
}

// Client contains helpers for interacting with Proxmox and Talos factory APIs.
type Client struct {
	baseURL            string
	tokenID            string
	secret             string
	node               string
	isoStorage         string
	schematicFile      string
	talosSchematicPath string
	templateJSONPath   string
	proxmoxHTTP        *http.Client
	factoryHTTP        *http.Client
}

// New validates the config and constructs a new Client.
func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("proxmox URL is required")
	}
	if cfg.TokenID == "" {
		return nil, errors.New("proxmox token ID is required")
	}
	if cfg.Secret == "" {
		return nil, errors.New("proxmox token secret is required")
	}
	if cfg.Node == "" {
		return nil, errors.New("proxmox node is required")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	proxmoxClient := cfg.HTTPClient
	if proxmoxClient == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.SkipTLSVerify}, //nolint:gosec // CLI tool mirrors curl -k behaviour
		}
		proxmoxClient = &http.Client{Timeout: timeout, Transport: tr}
	}

	factoryClient := cfg.FactoryClient
	if factoryClient == nil {
		factoryClient = &http.Client{Timeout: timeout}
	}

	schematicFile := cfg.SchematicFile
	if schematicFile == "" {
		schematicFile = defaultSchematicFile
	}

	isoStorage := cfg.ISOStorage
	if isoStorage == "" {
		isoStorage = defaultISOStorage
	}

	talosSchematic := cfg.TalosSchematicPath
	if talosSchematic == "" {
		talosSchematic = defaultTalosSchematicPath
	}

	templateJSON := cfg.TemplateJSONPath
	if templateJSON == "" {
		templateJSON = defaultTemplateJSONPath
	}

	return &Client{
		baseURL:            fmt.Sprintf("https://%s:8006/api2/json", cfg.URL),
		tokenID:            cfg.TokenID,
		secret:             cfg.Secret,
		node:               cfg.Node,
		isoStorage:         isoStorage,
		schematicFile:      schematicFile,
		talosSchematicPath: talosSchematic,
		templateJSONPath:   templateJSON,
		proxmoxHTTP:        proxmoxClient,
		factoryHTTP:        factoryClient,
	}, nil
}

// RefreshSchematic uploads the Talos factory schematic YAML and caches the returned ID locally.
func (c *Client) RefreshSchematic(ctx context.Context) (string, error) {
	data, err := os.ReadFile(c.talosSchematicPath)
	if err != nil {
		return "", fmt.Errorf("read schematic yaml: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://factory.talos.dev/schematics", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("build talos request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-yaml")

	resp, err := c.factoryHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("talos factory request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("talos factory returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode talos factory response: %w", err)
	}
	if payload.ID == "" {
		return "", errors.New("talos factory response missing id")
	}

	if err := os.WriteFile(c.schematicFile, []byte(payload.ID), 0o644); err != nil {
		return "", fmt.Errorf("write schematic cache: %w", err)
	}

	log.Info().Str("schematicID", payload.ID).Str("path", c.schematicFile).Msg("Cached Talos schematic ID")
	return payload.ID, nil
}

// ShowSchematic reads the cached schematic ID from disk.
func (c *Client) ShowSchematic() (string, error) {
	id, err := c.readSchematicID()
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", errors.New("no cached schematic ID; run refresh first")
	}
	return id, nil
}

// ClearSchematic removes the cached schematic ID file.
func (c *Client) ClearSchematic() error {
	if err := os.Remove(c.schematicFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove schematic cache: %w", err)
	}
	log.Info().Str("path", c.schematicFile).Msg("Cleared cached schematic ID")
	return nil
}

// EnsureSchematic returns the cached schematic ID, refreshing it if missing.
func (c *Client) EnsureSchematic(ctx context.Context) (string, error) {
	if id, err := c.readSchematicID(); err != nil {
		return "", err
	} else if id != "" {
		return id, nil
	}
	return c.RefreshSchematic(ctx)
}

// GetTalosImage downloads the Talos ISO for the provided version and uploads it to Proxmox.
func (c *Client) GetTalosImage(ctx context.Context, version string) error {
	if version == "" {
		return errors.New("talos version is required")
	}

	id, err := c.EnsureSchematic(ctx)
	if err != nil {
		return fmt.Errorf("ensure schematic: %w", err)
	}

	isoName := fmt.Sprintf("talos-%s-nocloud-amd64.iso", version)
	url := fmt.Sprintf("https://factory.talos.dev/image/%s/v%s/nocloud-amd64.iso", id, version)
	log.Info().
		Str("schematicID", id).
		Str("version", version).
		Str("url", url).
		Msg("Downloading Talos ISO")

	if err := c.downloadToFile(ctx, url, isoName); err != nil {
		return fmt.Errorf("download talos iso: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(isoName); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			log.Warn().Err(removeErr).Str("file", isoName).Msg("Failed to remove temporary ISO")
		}
	}()

	log.Info().
		Str("node", c.node).
		Str("storage", c.isoStorage).
		Str("file", isoName).
		Msg("Uploading ISO to Proxmox")

	if err := c.uploadISO(ctx, isoName); err != nil {
		return fmt.Errorf("upload iso to proxmox: %w", err)
	}

	log.Info().Str("version", version).Msg("Talos ISO uploaded successfully")
	return nil
}

// CreateTemplate ensures no VM with the same VMID exists, creates a new VM from JSON payload, and converts it to a template.
func (c *Client) CreateTemplate(ctx context.Context) error {
	payload, err := os.ReadFile(c.templateJSONPath)
	if err != nil {
		return fmt.Errorf("read template json: %w", err)
	}

	var meta struct {
		VMID json.Number `json:"vmid"`
		Name string      `json:"name"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return fmt.Errorf("parse template json: %w", err)
	}
	if meta.VMID == "" {
		return errors.New("template json missing vmid")
	}
	if meta.Name == "" {
		return errors.New("template json missing name")
	}
	vmid, err := meta.VMID.Int64()
	if err != nil {
		return fmt.Errorf("vmid is not a number: %w", err)
	}

	exists, err := c.vmExists(ctx, vmid)
	if err != nil {
		return fmt.Errorf("check vm exists: %w", err)
	}
	if exists {
		log.Warn().Int64("vmid", vmid).Msg("Existing VM found; deleting before template creation")
		if err := c.deleteVM(ctx, vmid); err != nil {
			return fmt.Errorf("delete vm %d: %w", vmid, err)
		}
		if err := c.waitForVMDeletion(ctx, vmid, 90*time.Second); err != nil {
			return fmt.Errorf("wait for vm deletion: %w", err)
		}
	}

	log.Info().
		Int64("vmid", vmid).
		Str("name", meta.Name).
		Msg("Creating Proxmox VM from template descriptor")

	if err := c.createVM(ctx, payload); err != nil {
		return fmt.Errorf("create vm: %w", err)
	}

	time.Sleep(2 * time.Second)

	if err := c.convertToTemplate(ctx, vmid); err != nil {
		return fmt.Errorf("convert vm %d to template: %w", vmid, err)
	}

	log.Info().Int64("vmid", vmid).Str("name", meta.Name).Msg("Template created successfully")
	return nil
}

func (c *Client) readSchematicID() (string, error) {
	data, err := os.ReadFile(c.schematicFile)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read schematic cache: %w", err)
	}
	id := strings.TrimSpace(string(data))
	return id, nil
}

func (c *Client) downloadToFile(ctx context.Context, url, filename string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}

	resp, err := c.factoryHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("download returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create iso file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write iso file: %w", err)
	}
	return nil
}

func (c *Client) uploadISO(ctx context.Context, isoName string) error {
	file, err := os.Open(isoName)
	if err != nil {
		return fmt.Errorf("open iso: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("content", "iso"); err != nil {
		return fmt.Errorf("write multipart field: %w", err)
	}

	part, err := writer.CreateFormFile("filename", filepath.Base(isoName))
	if err != nil {
		return fmt.Errorf("create multipart file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy iso into multipart: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	path := fmt.Sprintf("/nodes/%s/storage/%s/upload", c.node, c.isoStorage)
	req, err := c.newProxmoxRequest(ctx, http.MethodPost, path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.proxmoxHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("upload iso request failed: %w", err)
	}
	defer resp.Body.Close()

	return checkProxmoxResponse(resp)
}

func (c *Client) vmExists(ctx context.Context, vmid int64) (bool, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", c.node, vmid)
	req, err := c.newProxmoxRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.proxmoxHTTP.Do(req)
	if err != nil {
		return false, fmt.Errorf("check vm request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return false, fmt.Errorf("vm existence check returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
}

func (c *Client) deleteVM(ctx context.Context, vmid int64) error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d", c.node, vmid)
	req, err := c.newProxmoxRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	resp, err := c.proxmoxHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("delete vm request failed: %w", err)
	}
	defer resp.Body.Close()

	return checkProxmoxResponse(resp)
}

func (c *Client) waitForVMDeletion(ctx context.Context, vmid int64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for vm %d deletion", vmid)
		}

		exists, err := c.vmExists(ctx, vmid)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for deletion cancelled: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

func (c *Client) createVM(ctx context.Context, payload []byte) error {
	path := fmt.Sprintf("/nodes/%s/qemu", c.node)
	req, err := c.newProxmoxRequest(ctx, http.MethodPost, path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.proxmoxHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("create vm request failed: %w", err)
	}
	defer resp.Body.Close()

	return checkProxmoxResponse(resp)
}

func (c *Client) convertToTemplate(ctx context.Context, vmid int64) error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/template", c.node, vmid)
	req, err := c.newProxmoxRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	resp, err := c.proxmoxHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("convert to template request failed: %w", err)
	}
	defer resp.Body.Close()

	return checkProxmoxResponse(resp)
}

func (c *Client) newProxmoxRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build proxmox request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.secret))
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func checkProxmoxResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		io.Copy(io.Discard, resp.Body) // allow connection reuse
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("proxmox API %s %s returned %s: %s",
		resp.Request.Method,
		resp.Request.URL.Path,
		resp.Status,
		strings.TrimSpace(string(body)))
}
