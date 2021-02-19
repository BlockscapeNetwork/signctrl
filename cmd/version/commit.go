package version

import (
	"os/exec"
	"strings"
)

// CommitHash returns the commit hash of the current version.
func CommitHash() string {
	// TODO: This only works in SignCTRLs github repository directory.
	output, _ := exec.Command("go", "env", "GOPATH").Output()
	gopath := strings.TrimSuffix(string(output), "\n")

	commitHash, err := exec.Command("git", "rev-list", "-1", "HEAD", "--path", gopath, "/src/github.com/BlockscapeNetwork/signctrl").Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSuffix(string(commitHash), "\n")
}
