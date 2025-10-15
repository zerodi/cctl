package kubex

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zerodi/cctl/internal/executil"
)

// Client wraps kubectl interactions for namespaces scoped by Cluster API scripts.
type Client struct {
	namespace  string
	kubeconfig string
}

// New returns a kubectl-backed client.
func New(kubeconfig, namespace string) *Client {
	return &Client{kubeconfig: kubeconfig, namespace: namespace}
}

// WaitForSecret replicates wait_for_secret script behaviour.
func (c *Client) WaitForSecret(ctx context.Context, namePattern, clusterName string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	var mainRE *regexp.Regexp
	if namePattern != "" {
		re, err := regexp.Compile(namePattern)
		if err != nil {
			return "", fmt.Errorf("compile secret regex %q: %w", namePattern, err)
		}
		mainRE = re
	}
	fallbackRE := regexp.MustCompile(`(kubeconfig|talosconfig)`)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if sec, err := c.getSecretIfExists(ctx, namePattern); err != nil {
			return "", err
		} else if sec != "" {
			return sec, nil
		}

		names, err := c.listSecrets(ctx)
		if err != nil {
			return "", err
		}
		for _, sec := range names {
			if mainRE != nil && mainRE.MatchString(sec.Name) && sec.ClusterLabel == clusterName {
				return sec.Name, nil
			}
			if fallbackRE.MatchString(sec.Name) && sec.ClusterLabel == clusterName {
				return sec.Name, nil
			}
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("timed out waiting for secret like %q in namespace %s", namePattern, c.namespace)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}

// ExtractSecretToFile mirrors extract_secret_to_file.
func (c *Client) ExtractSecretToFile(ctx context.Context, secretName, preferredKey, outPath string) error {
	if secretName == "" {
		return errors.New("secret name is required")
	}
	secret, err := c.getSecret(ctx, secretName)
	if err != nil {
		return err
	}
	if len(secret.Data) == 0 {
		return fmt.Errorf("secret %s has no data", secretName)
	}

	var key string
	if preferredKey != "" {
		if _, ok := secret.Data[preferredKey]; ok {
			key = preferredKey
		}
	}
	if key == "" {
		if _, ok := secret.Data["value"]; ok {
			key = "value"
		}
	}
	if key == "" {
		var keys []string
		for k := range secret.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		key = keys[0]
	}

	raw, err := base64.StdEncoding.DecodeString(secret.Data[key])
	if err != nil {
		return fmt.Errorf("decode secret %s key %s: %w", secretName, key, err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("ensure output dir: %w", err)
	}
	if err := os.WriteFile(outPath, raw, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

// DiscoverControlPlaneIPs replicates discover_cp_ips.
func (c *Client) DiscoverControlPlaneIPs(ctx context.Context) ([]string, error) {
	args := []string{"get", "nodes", "-l", "node-role.kubernetes.io/control-plane", "-o", "json"}
	stdout, stderr, err := c.runCapture(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get nodes: %v: %s", err, strings.TrimSpace(stderr))
	}

	var resp struct {
		Items []struct {
			Status struct {
				Addresses []struct {
					Type    string `json:"type"`
					Address string `json:"address"`
				} `json:"addresses"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("parse kubectl nodes response: %w", err)
	}

	var ips []string
	seen := make(map[string]struct{})
	for _, item := range resp.Items {
		for _, addr := range item.Status.Addresses {
			if addr.Type == "InternalIP" && addr.Address != "" {
				if _, ok := seen[addr.Address]; !ok {
					seen[addr.Address] = struct{}{}
					ips = append(ips, addr.Address)
				}
			}
		}
	}
	return ips, nil
}

// RunKubectl streams stdout/stderr.
func (c *Client) RunKubectl(ctx context.Context, args ...string) error {
	env := c.env()
	if c.namespace != "" {
		args = append([]string{"-n", c.namespace}, args...)
	}
	return executil.RunStreaming(ctx, env, "kubectl", args...)
}

func (c *Client) getSecretIfExists(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	args := []string{"get", "secret", name, "-o", "json"}
	if c.namespace != "" {
		args = append([]string{"-n", c.namespace}, args...)
	}
	stdout, stderr, err := c.runCapture(ctx, args...)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			text := strings.ToLower(stderr)
			if strings.Contains(text, "notfound") || strings.Contains(text, "not found") {
				return "", nil
			}
		}
		return "", fmt.Errorf("kubectl get secret %s: %v: %s", name, err, strings.TrimSpace(stderr))
	}
	var resp struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return "", fmt.Errorf("parse secret %s: %w", name, err)
	}
	return resp.Metadata.Name, nil
}

type secretInfo struct {
	Name         string
	ClusterLabel string
}

func (c *Client) listSecrets(ctx context.Context) ([]secretInfo, error) {
	args := []string{"get", "secrets", "-o", "json"}
	if c.namespace != "" {
		args = append([]string{"-n", c.namespace}, args...)
	}
	stdout, stderr, err := c.runCapture(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get secrets: %v: %s", err, strings.TrimSpace(stderr))
	}

	var resp struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("parse secrets list: %w", err)
	}

	result := make([]secretInfo, 0, len(resp.Items))
	for _, item := range resp.Items {
		label := ""
		if item.Metadata.Labels != nil {
			label = item.Metadata.Labels["cluster.x-k8s.io/cluster-name"]
		}
		result = append(result, secretInfo{Name: item.Metadata.Name, ClusterLabel: label})
	}
	return result, nil
}

func (c *Client) getSecret(ctx context.Context, name string) (*secret, error) {
	args := []string{"get", "secret", name, "-o", "json"}
	if c.namespace != "" {
		args = append([]string{"-n", c.namespace}, args...)
	}
	stdout, stderr, err := c.runCapture(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get secret %s: %v: %s", name, err, strings.TrimSpace(stderr))
	}
	var resp secret
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("parse secret %s: %w", name, err)
	}
	return &resp, nil
}

type secret struct {
	Data map[string]string `json:"data"`
}

func (c *Client) runCapture(ctx context.Context, args ...string) (string, string, error) {
	env := c.env()
	return executil.RunCapture(ctx, env, "kubectl", args...)
}

func (c *Client) env() []string {
	if c.kubeconfig == "" {
		return nil
	}
	env := os.Environ()
	return append(env, "KUBECONFIG="+c.kubeconfig)
}
