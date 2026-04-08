//go:build !windows

package platform

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func PickPDF() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return runAndClean(
			"osascript",
			"-e",
			`POSIX path of (choose file with prompt "PDF 파일 선택" of type {"pdf"})`,
		)
	default:
		return "", fmt.Errorf("%s에서는 아직 파일 선택 대화상자를 지원하지 않습니다", runtime.GOOS)
	}
}

func PickDirectory() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return runAndClean(
			"osascript",
			"-e",
			`POSIX path of (choose folder with prompt "출력 폴더 선택")`,
		)
	default:
		return "", fmt.Errorf("%s에서는 아직 폴더 선택 대화상자를 지원하지 않습니다", runtime.GOOS)
	}
}

func RevealDirectory(path string) error {
	dir, err := validateExistingDirectory(path)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("open", dir)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("Finder 실행 실패: %w", err)
		}
		if cmd.Process != nil {
			_ = cmd.Process.Release()
		}
		return nil
	default:
		return fmt.Errorf("%s에서는 아직 출력 폴더 열기를 지원하지 않습니다", runtime.GOOS)
	}
}

func runAndClean(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
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
		return "", fmt.Errorf("대화상자 실행 실패: %s", combined)
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
