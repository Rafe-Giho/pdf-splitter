//go:build !windows

package pdf

import "os/exec"

func configureHiddenCommand(cmd *exec.Cmd) {}
