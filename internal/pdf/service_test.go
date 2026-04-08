package pdf

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	pdfTypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestBuildBoundariesBySpan(t *testing.T) {
	got := BuildBoundariesBySpan(10, 3)
	want := []int{3, 6, 9}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildBoundariesBySpan() = %v, want %v", got, want)
	}
}

func TestBuildRanges(t *testing.T) {
	got, err := BuildRanges(8, []int{2, 5})
	if err != nil {
		t.Fatalf("BuildRanges() error = %v", err)
	}

	want := []PageRange{
		{From: 1, Thru: 2},
		{From: 3, Thru: 5},
		{From: 6, Thru: 8},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildRanges() = %v, want %v", got, want)
	}
}

func TestBuildOutputNames(t *testing.T) {
	got := buildOutputNames("sample", 7, 3)
	want := []string{"sample_007.pdf", "sample_008.pdf", "sample_009.pdf"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildOutputNames() = %v, want %v", got, want)
	}
}

func TestSuggestedOutputDir(t *testing.T) {
	input := filepath.Join("C:\\work", "sample.pdf")
	want := filepath.Join("C:\\work")

	if got := SuggestedOutputDir(input); got != want {
		t.Fatalf("SuggestedOutputDir() = %q, want %q", got, want)
	}
}

func TestValidateInputPathAcceptsPDF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7\n1 0 obj\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := validateInputPath(path)
	if err != nil {
		t.Fatalf("validateInputPath() error = %v", err)
	}

	want, _ := filepath.Abs(path)
	if got != want {
		t.Fatalf("validateInputPath() = %q, want %q", got, want)
	}
}

func TestValidateInputPathRejectsInvalidSignature(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake.pdf")
	if err := os.WriteFile(path, []byte("not a pdf"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := validateInputPath(path)
	if err == nil || !strings.Contains(err.Error(), "PDF 헤더") {
		t.Fatalf("validateInputPath() error = %v, want PDF header error", err)
	}
}

func TestValidateInputPathRejectsTooLargeFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.pdf")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if err := file.Truncate(maxInputPDFSizeBytes + 1); err != nil {
		t.Fatal(err)
	}

	_, err = validateInputPath(path)
	if err == nil || !strings.Contains(err.Error(), "너무 큽니다") {
		t.Fatalf("validateInputPath() error = %v, want size limit error", err)
	}
}

func TestRejectUnsafeObjectGraphRejectsDangerousAction(t *testing.T) {
	obj := pdfTypes.Dict{
		"S": pdfTypes.Name("JavaScript"),
	}

	err := rejectUnsafeObjectGraph(nil, obj, map[string]struct{}{})
	if err == nil || !strings.Contains(err.Error(), "JavaScript 액션") {
		t.Fatalf("rejectUnsafeObjectGraph() error = %v, want JavaScript action rejection", err)
	}
}

func TestRejectUnsafeObjectGraphRejectsDangerousKey(t *testing.T) {
	obj := pdfTypes.Dict{
		"OpenAction": pdfTypes.Dict{},
	}

	err := rejectUnsafeObjectGraph(nil, obj, map[string]struct{}{})
	if err == nil || !strings.Contains(err.Error(), "자동 실행 액션") {
		t.Fatalf("rejectUnsafeObjectGraph() error = %v, want OpenAction rejection", err)
	}
}

func TestCommitStagedFilesRollsBackOnConflict(t *testing.T) {
	dir := t.TempDir()
	tempFile := filepath.Join(dir, "part.tmp.pdf")
	finalFile := filepath.Join(dir, "result.pdf")

	if err := os.WriteFile(tempFile, []byte("temp"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(finalFile, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := commitStagedFiles([]stagedOutputFile{{TempPath: tempFile, FinalPath: finalFile}})
	if err == nil || !strings.Contains(err.Error(), "이미 존재") {
		t.Fatalf("commitStagedFiles() error = %v, want conflict error", err)
	}
	if _, statErr := os.Stat(tempFile); statErr != nil {
		t.Fatalf("temp file should remain until cleanup: %v", statErr)
	}
}
