package exec

import (
	"context"
	"os/exec"
)

// Command defines the interface for executing commands.
type Command interface {
	Output() ([]byte, error)
	Path() string
	Args() []string
}

// CommandExecutor defines the interface for executing commands.
type CommandExecutor interface {
	// LookPath searches for an executable named file in the directories
	// named by the PATH environment variable.
	LookPath(file string) (string, error)

	// CommandContext is like Command but includes a context.
	CommandContext(ctx context.Context, name string, arg ...string) Command
}

// execCmdWrapper wraps *exec.Cmd to implement Command interface.
type execCmdWrapper struct {
	*exec.Cmd
}

// Path returns the path of the command.
func (e *execCmdWrapper) Path() string {
	return e.Cmd.Path
}

// Args returns the arguments of the command.
func (e *execCmdWrapper) Args() []string {
	return e.Cmd.Args
}

// OSCommandExecutor is an implementation of CommandExecutor using os/exec.
type OSCommandExecutor struct{}
