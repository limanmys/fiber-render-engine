package linux

import (
	"fmt"
	"os/exec"

	"github.com/limanmys/render-engine/internal/constants"
)

// This function executes specified command in shell
func Execute(input string) (string, error) {
	cmd := exec.Command(constants.EXEC_RUNNER, "-c", input)
	stdout, err := cmd.Output()
	if err != nil {
		return string(stdout), fmt.Errorf("command execution failed: %w", err)
	}

	return string(stdout), nil
}
