//go:build windows

package platform

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func PickPDF() (string, error) {
	return runHiddenPowerShell(
		`[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false);
Add-Type -AssemblyName System.Windows.Forms;
$dialog = New-Object System.Windows.Forms.OpenFileDialog;
$dialog.Filter = 'PDF files (*.pdf)|*.pdf';
$dialog.Multiselect = $false;
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
	[Console]::WriteLine($dialog.FileName)
}`,
	)
}

func PickDirectory() (string, error) {
	return runHiddenPowerShell(
		`[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false);
Add-Type -AssemblyName System.Windows.Forms;
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog;
$dialog.ShowNewFolderButton = $true;
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
	[Console]::WriteLine($dialog.SelectedPath)
}`,
	)
}

func RevealDirectory(path string) error {
	dir, err := validateExistingDirectory(path)
	if err != nil {
		return err
	}

	cmd := exec.Command("explorer.exe", dir)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("탐색기 실행 실패: %w", err)
	}
	if cmd.Process != nil {
		_ = cmd.Process.Release()
	}

	return nil
}

func runHiddenPowerShell(script string) (string, error) {
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-STA",
		"-WindowStyle", "Hidden",
		"-Command", script,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		combined := strings.TrimSpace(stderr.String())
		if combined == "" {
			combined = err.Error()
		}
		if isUserCancelled(combined) {
			return "", nil
		}
		return "", fmt.Errorf("파일 선택창 실행 실패: %s", combined)
	}

	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return "", nil
	}

	return filepath.Clean(value), nil
}

func isUserCancelled(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "user canceled") ||
		strings.Contains(lower, "cancelled") ||
		strings.Contains(lower, "canceled")
}
