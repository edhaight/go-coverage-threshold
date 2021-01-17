package cover

import (
	"os/exec"
)

// Run executes `go test -cover ./...` and returns the raw result.
func Run(profile bool, packages ...string) ([]byte, error) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}
	coverProfile := "-cover"
	if profile {
		coverProfile = "-coverprofile=cover.out"
	}

	cmdArgs := []string{
		"test",
		coverProfile,
	}

	if len(packages) == 0 {
		cmdArgs = append(cmdArgs, "./...")
	} else {
		cmdArgs = append(cmdArgs, packages...)
	}
	// gas lint warns about possible injection here, but we're fine
	cmd := exec.Command(goExe, cmdArgs...) // nolint: gas,gosec

	output, err := cmd.CombinedOutput()
	return output, err
}
