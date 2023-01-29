package linux

import (
	"os/exec"

	"github.com/limanmys/render-engine/internal/constants"
)

// This function executes specified command in shell
func Execute(input string) string {
	cmd := exec.Command(constants.EXEC_RUNNER, "-c", input)
	stdout, _ := cmd.Output()

	return string(stdout)
}
