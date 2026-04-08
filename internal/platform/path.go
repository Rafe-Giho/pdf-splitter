package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateExistingDirectory(path string) (string, error) {
	cleanPath := filepath.Clean(strings.Trim(strings.TrimSpace(path), `"`))
	if cleanPath == "" || cleanPath == "." {
		return "", fmt.Errorf("열 위치가 비어 있습니다")
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("경로 해석 실패: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("출력 폴더를 찾을 수 없습니다")
		}
		return "", fmt.Errorf("출력 폴더 확인 실패: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("출력 폴더가 아니라 파일이 지정되었습니다")
	}

	return absPath, nil
}
