// Package exec provides interfaces and implementations for command execution.
// It centralizes command execution functionality for use by other internal packages.
package exec

import (
	"context"
	"fmt"
	"os/exec"
)

// Command defines the interface for executing commands.
type Command interface {
	Output() ([]byte, error)
	Path() string
	Args() []string
}

// execCmdWrapper wraps *exec.Cmd to implement Command interface.
type execCmdWrapper struct {
	*exec.Cmd
}

// CommandExecutor defines the interface for executing commands.
type CommandExecutor interface {
	// LookPath searches for an executable named file in the directories
	// named by the PATH environment variable.
	LookPath(file string) (string, error)

	// CommandContext is like Command but includes a context.
	CommandContext(ctx context.Context, name string, arg ...string) Command
}

// OSCommandExecutor is an implementation of CommandExecutor using os/exec.
type OSCommandExecutor struct{}

// Path returns the path of the command.
func (e *execCmdWrapper) Path() string {
	return e.Cmd.Path
}

// Args returns the arguments of the command.
func (e *execCmdWrapper) Args() []string {
	return e.Cmd.Args
}

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
