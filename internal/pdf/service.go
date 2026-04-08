package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	pdfModel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfTypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type Service struct{}

type PageMeta struct {
	Number int
	Width  float64
	Height float64
	Ratio  float32
}

type PageRange struct {
	From int
	Thru int
}

type OutputPlan struct {
	Directory string
	Files     []string
}

type InspectResult struct {
	PageCount int
	OutputDir string
	Pages     []PageMeta
}

type SplitRequest struct {
	InputPath    string
	Boundaries   []int
	RequireSplit bool
}

type SplitResult struct {
	PageCount int
	OutputDir string
	Files     []string
	Ranges    []PageRange
}

const (
	maxInputPDFSizeBytes = 256 << 20
	maxInputPDFPages     = 2000
	pdfHeaderProbeBytes  = 1024
)

var dangerousActionNames = map[string]string{
	"GoToE":            "임베디드 문서 이동 액션",
	"GoToR":            "외부 문서 이동 액션",
	"ImportData":       "데이터 가져오기 액션",
	"JavaScript":       "JavaScript 액션",
	"Launch":           "외부 실행 액션",
	"Movie":            "미디어 실행 액션",
	"Rendition":        "리치 미디어 액션",
	"RichMediaExecute": "리치 미디어 실행 액션",
	"Sound":            "사운드 실행 액션",
	"SubmitForm":       "폼 제출 액션",
}

var dangerousDictKeys = map[string]string{
	"AA":               "추가 액션",
	"Collection":       "포트폴리오 컬렉션",
	"EmbeddedFiles":    "임베디드 파일",
	"JavaScript":       "JavaScript name tree",
	"OpenAction":       "자동 실행 액션",
	"RichMediaContent": "리치 미디어",
	"XFA":              "XFA 폼",
}

var dangerousSubtypeNames = map[string]string{
	"FileAttachment": "첨부파일 주석",
	"Movie":          "동영상 주석",
	"RichMedia":      "리치 미디어 주석",
	"Screen":         "스크린 주석",
	"Sound":          "사운드 주석",
}

type stagedOutputFile struct {
	TempPath  string
	FinalPath string
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Inspect(path string) (InspectResult, error) {
	cleanPath, err := validateInputPath(path)
	if err != nil {
		return InspectResult{}, err
	}

	ctx, err := preflightProcessablePDF(cleanPath)
	if err != nil {
		return InspectResult{}, err
	}

	return InspectResult{
		PageCount: ctx.PageCount,
		OutputDir: filepath.Dir(cleanPath),
		Pages:     defaultPages(ctx.PageCount),
	}, nil
}

func (s *Service) PlanOutput(inputPath string, fileCount int) (OutputPlan, error) {
	cleanPath, err := validateInputPath(inputPath)
	if err != nil {
		return OutputPlan{}, err
	}
	if fileCount < 1 {
		return OutputPlan{Directory: filepath.Dir(cleanPath)}, nil
	}

	dir := filepath.Dir(cleanPath)
	base := strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath))
	start, err := nextSequenceNumber(dir, base)
	if err != nil {
		return OutputPlan{}, err
	}

	return OutputPlan{
		Directory: dir,
		Files:     buildOutputNames(base, start, fileCount),
	}, nil
}

func (s *Service) Split(req SplitRequest) (SplitResult, error) {
	cleanPath, err := validateInputPath(req.InputPath)
	if err != nil {
		return SplitResult{}, err
	}

	ctx, err := preflightProcessablePDF(cleanPath)
	if err != nil {
		return SplitResult{}, err
	}
	pageCount := ctx.PageCount

	ranges, err := BuildRanges(pageCount, req.Boundaries)
	if err != nil {
		return SplitResult{}, err
	}
	if req.RequireSplit && len(ranges) < 2 {
		return SplitResult{}, errors.New("구분선을 하나 이상 선택해 주세요")
	}

	dir := filepath.Dir(cleanPath)
	base := strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath))
	start, err := nextSequenceNumber(dir, base)
	if err != nil {
		return SplitResult{}, err
	}
	names := buildOutputNames(base, start, len(ranges))

	staged := make([]stagedOutputFile, 0, len(ranges))
	for i, pageRange := range ranges {
		ctxNew, err := pdfcpu.ExtractPages(ctx, api.PagesForPageRange(pageRange.From, pageRange.Thru), false)
		if err != nil {
			cleanupStagedFiles(staged)
			return SplitResult{}, fmt.Errorf("%d-%d 페이지 추출 실패: %w", pageRange.From, pageRange.Thru, err)
		}

		outPath := filepath.Join(dir, names[i])
		tmpPath := filepath.Join(dir, fmt.Sprintf(".pdfsplitter-%d-%03d.tmp.pdf", os.Getpid(), i+1))
		if err := api.WriteContextFile(ctxNew, tmpPath); err != nil {
			cleanupStagedFiles(staged)
			_ = os.Remove(tmpPath)
			return SplitResult{}, fmt.Errorf("분할 파일 저장 실패: %w", err)
		}
		staged = append(staged, stagedOutputFile{TempPath: tmpPath, FinalPath: outPath})
	}

	if err := commitStagedFiles(staged); err != nil {
		cleanupStagedFiles(staged)
		return SplitResult{}, err
	}

	files := stagedFinalPaths(staged)
	return SplitResult{
		PageCount: pageCount,
		OutputDir: dir,
		Files:     files,
		Ranges:    ranges,
	}, nil
}

func BuildRanges(pageCount int, boundaries []int) ([]PageRange, error) {
	if pageCount < 1 {
		return nil, errors.New("페이지 수가 유효하지 않습니다")
	}

	cleanBoundaries, err := normalizeBoundaries(pageCount, boundaries)
	if err != nil {
		return nil, err
	}

	ranges := make([]PageRange, 0, len(cleanBoundaries)+1)
	from := 1
	for _, boundary := range cleanBoundaries {
		ranges = append(ranges, PageRange{From: from, Thru: boundary})
		from = boundary + 1
	}
	ranges = append(ranges, PageRange{From: from, Thru: pageCount})

	return ranges, nil
}

func BuildBoundariesBySpan(pageCount, span int) []int {
	if pageCount < 2 || span < 1 {
		return nil
	}

	boundaries := make([]int, 0, pageCount/span)
	for page := span; page < pageCount; page += span {
		boundaries = append(boundaries, page)
	}

	return boundaries
}

func defaultPages(pageCount int) []PageMeta {
	pages := make([]PageMeta, 0, pageCount)
	for i := 1; i <= pageCount; i++ {
		pages = append(pages, PageMeta{
			Number: i,
			Width:  210,
			Height: 297,
			Ratio:  297.0 / 210.0,
		})
	}
	return pages
}

func pagesFromDims(dims []pdfTypes.Dim) []PageMeta {
	pages := make([]PageMeta, 0, len(dims))
	for i, dim := range dims {
		ratio := float32(1.414)
		if dim.Width > 0 && dim.Height > 0 {
			ratio = float32(dim.Height / dim.Width)
		}
		pages = append(pages, PageMeta{
			Number: i + 1,
			Width:  dim.Width,
			Height: dim.Height,
			Ratio:  ratio,
		})
	}
	return pages
}

func normalizeBoundaries(pageCount int, boundaries []int) ([]int, error) {
	if len(boundaries) == 0 {
		return nil, nil
	}

	unique := map[int]struct{}{}
	clean := make([]int, 0, len(boundaries))
	for _, boundary := range boundaries {
		if boundary < 1 || boundary >= pageCount {
			return nil, fmt.Errorf("잘못된 구분선이 포함되어 있습니다: %d", boundary)
		}
		if _, exists := unique[boundary]; exists {
			continue
		}
		unique[boundary] = struct{}{}
		clean = append(clean, boundary)
	}

	sort.Ints(clean)
	return clean, nil
}

func buildOutputNames(base string, start, count int) []string {
	if count < 1 {
		return nil
	}

	width := 3
	if digits := len(strconv.Itoa(start + count - 1)); digits > width {
		width = digits
	}

	files := make([]string, 0, count)
	for i := 0; i < count; i++ {
		files = append(files, fmt.Sprintf("%s_%0*d.pdf", base, width, start+i))
	}
	return files
}

func nextSequenceNumber(dir, base string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("출력 폴더 조회 실패: %w", err)
	}

	pattern := regexp.MustCompile(`^` + regexp.QuoteMeta(base) + `_(\d+)\.pdf$`)
	maxNumber := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := pattern.FindStringSubmatch(entry.Name())
		if len(match) != 2 {
			continue
		}
		number, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		if number > maxNumber {
			maxNumber = number
		}
	}

	return maxNumber + 1, nil
}

func EstimateFileCount(pageCount, span int) int {
	return len(BuildBoundariesBySpan(pageCount, span)) + boolToInt(pageCount > 0)
}

func SuggestedOutputDir(inputPath string) string {
	return filepath.Dir(inputPath)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func validateInputPath(path string) (string, error) {
	cleanPath := filepath.Clean(strings.Trim(strings.TrimSpace(path), `"`))
	if cleanPath == "" || cleanPath == "." {
		return "", errors.New("PDF 파일 경로를 입력해 주세요")
	}
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("PDF 파일 경로 해석 실패: %w", err)
	}
	if !strings.EqualFold(filepath.Ext(cleanPath), ".pdf") {
		return "", errors.New("PDF 파일만 선택할 수 있습니다")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New("선택한 PDF 파일을 찾을 수 없습니다")
		}
		return "", fmt.Errorf("PDF 파일 확인 실패: %w", err)
	}
	if info.IsDir() {
		return "", errors.New("PDF 파일 경로에 폴더가 선택되었습니다")
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("정상적인 PDF 파일만 선택할 수 있습니다")
	}
	if info.Size() <= 0 {
		return "", errors.New("비어 있는 PDF 파일은 처리할 수 없습니다")
	}
	if info.Size() > maxInputPDFSizeBytes {
		return "", fmt.Errorf("PDF 파일이 너무 큽니다. %dMB 이하 파일만 처리할 수 있습니다", maxInputPDFSizeBytes>>20)
	}
	if err := validatePDFSignature(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

func preflightProcessablePDF(path string) (*pdfModel.Context, error) {
	if err := api.ValidateFile(path, nil); err != nil {
		if isEncryptedPDFError(err) {
			return nil, errors.New("암호화된 PDF는 처리할 수 없습니다")
		}
		return nil, fmt.Errorf("유효하지 않거나 손상된 PDF입니다: %w", err)
	}

	ctx, err := api.ReadContextFile(path)
	if err != nil {
		if isEncryptedPDFError(err) {
			return nil, errors.New("암호화된 PDF는 처리할 수 없습니다")
		}
		return nil, fmt.Errorf("PDF 분석 실패: %w", err)
	}
	if ctx.PageCount < 1 {
		return nil, errors.New("페이지가 없는 PDF는 처리할 수 없습니다")
	}
	if ctx.PageCount > maxInputPDFPages {
		return nil, fmt.Errorf("PDF 페이지 수가 너무 많습니다. %d페이지 이하 파일만 처리할 수 있습니다", maxInputPDFPages)
	}
	if err := enforcePDFSecurityPolicy(ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func validatePDFSignature(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("PDF 파일 열기 실패: %w", err)
	}
	defer file.Close()

	buf := make([]byte, pdfHeaderProbeBytes)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return fmt.Errorf("PDF 헤더 확인 실패: %w", err)
	}
	if !bytes.Contains(buf[:n], []byte("%PDF-")) {
		return errors.New("PDF 헤더가 올바르지 않습니다. 실제 PDF 파일만 선택해 주세요")
	}

	return nil
}

func enforcePDFSecurityPolicy(ctx *pdfModel.Context) error {
	if ctx.Encrypt != nil {
		return errors.New("암호화된 PDF는 처리할 수 없습니다")
	}

	attachments, err := ctx.ListAttachments()
	if err != nil {
		return fmt.Errorf("PDF 첨부파일 검사 실패: %w", err)
	}
	if len(attachments) > 0 {
		return errors.New("첨부파일이 포함된 PDF는 처리할 수 없습니다")
	}

	catalog, err := ctx.Catalog()
	if err != nil {
		return fmt.Errorf("PDF 카탈로그 검사 실패: %w", err)
	}

	if err := rejectUnsafeObjectGraph(ctx.XRefTable, catalog, map[string]struct{}{}); err != nil {
		return err
	}

	return nil
}

func rejectUnsafeObjectGraph(xRefTable *pdfModel.XRefTable, obj pdfTypes.Object, visited map[string]struct{}) error {
	switch v := obj.(type) {
	case nil:
		return nil
	case pdfTypes.IndirectRef:
		if xRefTable == nil {
			return nil
		}
		key := indirectRefKey(v)
		if _, ok := visited[key]; ok {
			return nil
		}
		visited[key] = struct{}{}

		resolved, err := xRefTable.Dereference(v)
		if err != nil {
			return fmt.Errorf("PDF 참조 해석 실패: %w", err)
		}
		return rejectUnsafeObjectGraph(xRefTable, resolved, visited)
	case pdfTypes.Dict:
		return rejectUnsafeDict(xRefTable, v, visited)
	case pdfTypes.StreamDict:
		return rejectUnsafeDict(xRefTable, v.Dict, visited)
	case pdfTypes.Array:
		for _, item := range v {
			if err := rejectUnsafeObjectGraph(xRefTable, item, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

func rejectUnsafeDict(xRefTable *pdfModel.XRefTable, d pdfTypes.Dict, visited map[string]struct{}) error {
	if d == nil {
		return nil
	}

	if actionName := d.NameEntry("S"); actionName != nil {
		if reason, ok := dangerousActionNames[*actionName]; ok {
			return fmt.Errorf("%s이(가) 포함된 PDF는 처리할 수 없습니다", reason)
		}
	}
	if subtype := d.NameEntry("Subtype"); subtype != nil {
		if reason, ok := dangerousSubtypeNames[*subtype]; ok {
			return fmt.Errorf("%s이(가) 포함된 PDF는 처리할 수 없습니다", reason)
		}
	}

	for key, value := range d {
		if reason, ok := dangerousDictKeys[key]; ok {
			return fmt.Errorf("%s이(가) 포함된 PDF는 처리할 수 없습니다", reason)
		}
		if err := rejectUnsafeObjectGraph(xRefTable, value, visited); err != nil {
			return err
		}
	}

	return nil
}

func indirectRefKey(ir pdfTypes.IndirectRef) string {
	return fmt.Sprintf("%d:%d", ir.ObjectNumber.Value(), ir.GenerationNumber.Value())
}

func isEncryptedPDFError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "encrypted") ||
		strings.Contains(msg, "correct password") ||
		strings.Contains(msg, "owner password") ||
		strings.Contains(msg, "user password")
}

func commitStagedFiles(files []stagedOutputFile) error {
	for _, file := range files {
		if _, err := os.Stat(file.FinalPath); err == nil {
			return fmt.Errorf("출력 파일이 이미 존재합니다: %s", filepath.Base(file.FinalPath))
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("출력 파일 확인 실패: %w", err)
		}
	}

	moved := make([]string, 0, len(files))
	for i, file := range files {
		if err := os.Rename(file.TempPath, file.FinalPath); err != nil {
			for _, finalPath := range moved {
				_ = os.Remove(finalPath)
			}
			for _, pending := range files[i:] {
				_ = os.Remove(pending.TempPath)
			}
			return fmt.Errorf("분할 파일 확정 실패: %w", err)
		}
		moved = append(moved, file.FinalPath)
	}

	return nil
}

func cleanupStagedFiles(files []stagedOutputFile) {
	for _, file := range files {
		_ = os.Remove(file.TempPath)
	}
}

func stagedFinalPaths(files []stagedOutputFile) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.FinalPath)
	}
	return paths
}
