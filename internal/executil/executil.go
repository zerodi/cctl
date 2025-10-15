package executil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EnsureCommands checks that all provided binaries are available on PATH.
func EnsureCommands(names ...string) error {
	var missing []string
	for _, name := range names {
		if name == "" {
			continue
		}
		if _, err := exec.LookPath(name); err != nil {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required commands not found: %s", strings.Join(missing, ", "))
	}
	return nil
}

// RunStreaming executes a command, wiring stdout/stderr to the current process.
func RunStreaming(ctx context.Context, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(env) > 0 {
		cmd.Env = env
	}
	return cmd.Run()
}

// RunCapture executes a command and returns stdout/stderr.
func RunCapture(ctx context.Context, env []string, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if len(env) > 0 {
		cmd.Env = env
	}
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
