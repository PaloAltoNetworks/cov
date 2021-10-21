package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetDiffFiles will return the added and changed files given a target branch
func GetDiffFiles(branch string) (files []string, err error) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=AMCR", branch)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(stderr.Bytes()))
	}

	return strings.Split(string(stdout.Bytes()), "\n"), nil

}
