package dts

import (
	"os/exec"
)

func ExecCommand(args []string) ([]byte, error) {
	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}
	return exec.Command(command, args...).Output()
}
