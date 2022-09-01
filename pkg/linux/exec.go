package linux

import "os/exec"

func Execute(input string) string {
	cmd := exec.Command("/bin/bash", "-c", input)
	stdout, _ := cmd.Output()

	return string(stdout)
}
