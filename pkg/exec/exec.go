package exec

import (
	"os"
	"os/exec"
)

// RunCommand executes a given command
func RunCommand(command string, args []string) (string, error) {
	out, err := exec.Command(
		command, args...,
	).CombinedOutput()
	return string(out), err
}

// RunInteractiveCommand runs the given command interactively, with stdin/stdout/stderr connected
func RunInteractiveCommand(command string, args []string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
