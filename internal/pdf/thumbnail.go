package pdf

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

const (
	preferredPDFBoxJarName = "pdfbox-app-3.0.7.jar"
	thumbnailCacheMaxAge   = 7 * 24 * time.Hour
	thumbnailCacheMaxDirs  = 24
	thumbnailRenderTimeout = 90 * time.Second
)

type Thumbnailer struct {
	javaPath string
	jarPath  string
	cacheDir string
}

func NewThumbnailer() (*Thumbnailer, error) {
	javaPath, err := exec.LookPath("java")
	if err != nil {
		return nil, fmt.Errorf("Java를 찾을 수 없습니다: %w", err)
	}

	jarPath, err := findPDFBoxJar()
	if err != nil {
		return nil, err
	}

	cacheRoot := resolveThumbnailCacheRoot()
	cacheDir := filepath.Join(cacheRoot, "pdfsplitter", "thumbnails")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("썸네일 캐시 폴더 생성 실패: %w", err)
	}

	t := &Thumbnailer{
		javaPath: javaPath,
		jarPath:  jarPath,
		cacheDir: cacheDir,
	}
	_ = pruneThumbnailCache(cacheDir, time.Now(), thumbnailCacheMaxAge, thumbnailCacheMaxDirs)

	return t, nil
}

func (t *Thumbnailer) LoadBatch(inputPath string, startPage, endPage, dpi int) (map[int]image.Image, error) {
	cleanPath, err := validateInputPath(inputPath)
	if err != nil {
		return nil, err
	}
	if startPage < 1 || endPage < startPage {
		return nil, fmt.Errorf("잘못된 썸네일 범위입니다: %d-%d", startPage, endPage)
	}
	if dpi < 1 {
		dpi = 30
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("PDF 파일 확인 실패: %w", err)
	}

	key := fmt.Sprintf("%x", sha1.Sum([]byte(cleanPath+"|"+strconv.FormatInt(info.Size(), 10)+"|"+info.ModTime().UTC().Format(time.RFC3339Nano)+"|"+strconv.Itoa(dpi))))
	dir := filepath.Join(t.cacheDir, key)
	prefix := filepath.Join(dir, "page")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("썸네일 캐시 폴더 생성 실패: %w", err)
	}

	missing := false
	for page := startPage; page <= endPage; page++ {
		if _, err := os.Stat(filepath.Join(dir, fmt.Sprintf("page-%d.png", page))); err != nil {
			missing = true
			break
		}
	}

	if missing {
		args := []string{
			"-Xms64m",
			"-Xmx512m",
			"-Djava.awt.headless=true",
			"-jar", t.jarPath,
			"render",
			"-i=" + cleanPath,
			"-startPage=" + strconv.Itoa(startPage),
			"-endPage=" + strconv.Itoa(endPage),
			"-format=png",
			"-prefix=" + prefix,
			"-dpi=" + strconv.Itoa(dpi),
			"-subsampling",
		}
		ctx, cancel := context.WithTimeout(context.Background(), thumbnailRenderTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, t.javaPath, args...)
		configureHiddenCommand(cmd)
		if out, err := cmd.CombinedOutput(); err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("페이지 썸네일 생성 시간이 너무 오래 걸려 중단했습니다")
			}
			return nil, fmt.Errorf("페이지 썸네일 생성 실패: %v: %s", err, string(out))
		}
	}

	images := make(map[int]image.Image, endPage-startPage+1)
	for page := startPage; page <= endPage; page++ {
		path := filepath.Join(dir, fmt.Sprintf("page-%d.png", page))
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("썸네일 파일 열기 실패: %w", err)
		}
		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("썸네일 디코드 실패: %w", err)
		}
		images[page] = img
	}

	return images, nil
}

func findPDFBoxJar() (string, error) {
	candidates := []string{}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, preferredPDFBoxJarName),
			filepath.Join(exeDir, "tools", "pdfbox", preferredPDFBoxJarName),
			filepath.Join(exeDir, "..", "tools", "pdfbox", preferredPDFBoxJarName),
			filepath.Join(exeDir, "..", "..", "tools", "pdfbox", preferredPDFBoxJarName),
			filepath.Join(exeDir, "pdfbox-app-3.0.3.jar"),
			filepath.Join(exeDir, "tools", "pdfbox", "pdfbox-app-3.0.3.jar"),
			filepath.Join(exeDir, "..", "tools", "pdfbox", "pdfbox-app-3.0.3.jar"),
			filepath.Join(exeDir, "..", "..", "tools", "pdfbox", "pdfbox-app-3.0.3.jar"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, preferredPDFBoxJarName),
			filepath.Join(wd, "tools", "pdfbox", preferredPDFBoxJarName),
			filepath.Join(wd, "pdfbox-app-3.0.3.jar"),
			filepath.Join(wd, "tools", "pdfbox", "pdfbox-app-3.0.3.jar"),
		)
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("PDF 썸네일 렌더러를 찾을 수 없습니다: %s", preferredPDFBoxJarName)
}

func resolveThumbnailCacheRoot() string {
	candidates := []string{}
	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "cache"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, ".cache"))
	}
	if tmp := os.TempDir(); tmp != "" {
		candidates = append(candidates, filepath.Join(tmp, "pdfsplitter-cache"))
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if err := os.MkdirAll(candidate, 0o755); err == nil {
			return candidate
		}
	}

	return ".cache"
}

func pruneThumbnailCache(cacheDir string, now time.Time, maxAge time.Duration, maxDirs int) error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	type cacheEntry struct {
		path    string
		modTime time.Time
	}

	kept := make([]cacheEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(cacheDir, entry.Name())
		if maxAge > 0 && now.Sub(info.ModTime()) > maxAge {
			_ = os.RemoveAll(path)
			continue
		}
		kept = append(kept, cacheEntry{path: path, modTime: info.ModTime()})
	}

	if maxDirs <= 0 || len(kept) <= maxDirs {
		return nil
	}

	sort.Slice(kept, func(i, j int) bool {
		return kept[i].modTime.After(kept[j].modTime)
	})
	for _, entry := range kept[maxDirs:] {
		_ = os.RemoveAll(entry.path)
	}

	return nil
}
