//go:build windows

package pdf

import (
	"os/exec"
	"syscall"
)

func configureHiddenCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
