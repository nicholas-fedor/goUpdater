// Package exec provides interfaces and implementations for command execution.
// It centralizes command execution functionality for use by other internal packages.
package exec

import (
	"context"
	"fmt"
	"os/exec"
)

// LookPath implements CommandExecutor.LookPath.
func (o OSCommandExecutor) LookPath(file string) (string, error) {
	path, err := exec.LookPath(file)
	if err != nil {
		return "", fmt.Errorf("failed to find executable %q: %w", file, err)
	}

	return path, nil
}

// CommandContext implements CommandExecutor.CommandContext.
func (o OSCommandExecutor) CommandContext(ctx context.Context, name string, arg ...string) Command {
	return &execCmdWrapper{Cmd: exec.CommandContext(ctx, name, arg...)}
}
