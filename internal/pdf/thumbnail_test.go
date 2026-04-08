package pdf

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneThumbnailCacheRemovesExpiredAndOverflowDirs(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	makeDir := func(name string, modTime time.Time) string {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
		return path
	}

	newest := makeDir("newest", now.Add(-1*time.Hour))
	second := makeDir("second", now.Add(-2*time.Hour))
	makeDir("expired", now.Add(-10*24*time.Hour))
	third := makeDir("third", now.Add(-3*time.Hour))

	if err := pruneThumbnailCache(dir, now, 7*24*time.Hour, 2); err != nil {
		t.Fatalf("pruneThumbnailCache() error = %v", err)
	}

	if _, err := os.Stat(newest); err != nil {
		t.Fatalf("expected newest cache dir to remain: %v", err)
	}
	if _, err := os.Stat(second); err != nil {
		t.Fatalf("expected second cache dir to remain: %v", err)
	}
	if _, err := os.Stat(third); !os.IsNotExist(err) {
		t.Fatalf("expected overflow cache dir to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "expired")); !os.IsNotExist(err) {
		t.Fatalf("expected expired cache dir to be removed, got err=%v", err)
	}
}
