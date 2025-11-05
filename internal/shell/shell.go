package shell

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, errb.String())
	}
	return out.String(), nil
}
