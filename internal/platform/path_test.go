package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateExistingDirectoryAcceptsDirectory(t *testing.T) {
	dir := t.TempDir()

	got, err := validateExistingDirectory(dir)
	if err != nil {
		t.Fatalf("validateExistingDirectory() error = %v", err)
	}

	want, _ := filepath.Abs(dir)
	if got != want {
		t.Fatalf("validateExistingDirectory() = %q, want %q", got, want)
	}
}

func TestValidateExistingDirectoryRejectsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := validateExistingDirectory(path); err == nil {
		t.Fatal("validateExistingDirectory() expected file rejection")
	}
}
