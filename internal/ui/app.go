package ui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"pdfsplitter/internal/pdf"
	"pdfsplitter/internal/platform"
)

type palette struct {
	Page         color.NRGBA
	Surface      color.NRGBA
	SurfaceMuted color.NRGBA
	SurfaceAlt   color.NRGBA
	Primary      color.NRGBA
	PrimarySoft  color.NRGBA
	Accent       color.NRGBA
	Text         color.NRGBA
	TextMuted    color.NRGBA
	Success      color.NRGBA
	SuccessSoft  color.NRGBA
	Error        color.NRGBA
	ErrorSoft    color.NRGBA
	Border       color.NRGBA
	Hover        color.NRGBA
	Pressed      color.NRGBA
}

type statusTone int

const (
	statusInfo statusTone = iota
	statusSuccess
	statusError
)

type inspectTaskResult struct {
	Result pdf.InspectResult
	Err    error
}

type splitTaskResult struct {
	Result pdf.SplitResult
	Err    error
}

type previewMode int

const (
	previewList previewMode = iota
	previewGrid
)

type themeMode int

const (
	themeModeSystem themeMode = iota
	themeModeLight
	themeModeDark
)

type thumbnailState struct {
	Op      paint.ImageOp
	Loaded  bool
	Loading bool
	Failed  bool
}

type thumbnailTaskResult struct {
	InputPath string
	StartPage int
	EndPage   int
	Images    map[int]paint.ImageOp
	Err       error
}

type App struct {
	window  *app.Window
	theme   *material.Theme
	colors  palette
	service *pdf.Service
	thumbs  *pdf.Thumbnailer
	ops     op.Ops

	controlList widget.List
	pageList    widget.List
	decorations widget.Decorations

	dropZoneBtn     widget.Clickable
	analyzeBtn      widget.Clickable
	spanToggleBtn   widget.Clickable
	spanMenuBtn     widget.Clickable
	applySpanBtn    widget.Clickable
	listModeBtn     widget.Clickable
	gridModeBtn     widget.Clickable
	splitBtn        widget.Clickable
	openOutputBtn   widget.Clickable
	closePreviewBtn widget.Clickable
	themeSystemBtn  widget.Clickable
	themeLightBtn   widget.Clickable
	themeDarkBtn    widget.Clickable

	pathEditor widget.Editor
	spanEditor widget.Editor

	inputPath          string
	pageCount          int
	pages              []pdf.PageMeta
	pageOpenBtns       []widget.Clickable
	separatorBtns      []widget.Clickable
	spanOptionBtns     []widget.Clickable
	boundaries         map[int]bool
	thumbnails         map[int]*thumbnailState
	thumbnailErrors    map[int]string
	outputDirectory    string
	outputPreview      []string
	completedOutputDir string
	previewMode        previewMode
	spanPresetEnabled  bool
	spanDropdownOpen   bool
	spanOptions        []int
	selectedPage       int
	themeMode          themeMode

	busy             bool
	thumbLoading     bool
	thumbLoadedPages int
	thumbTotalPages  int
	activeTask       string

	statusTitle  string
	statusDetail string
	statusTone   statusTone

	inspectCh chan inspectTaskResult
	thumbCh   chan thumbnailTaskResult
	splitCh   chan splitTaskResult
}

func New(window *app.Window) *App {
	th := material.NewTheme()
	th.Shaper = text.NewShaper()
	thumbs, _ := pdf.NewThumbnailer()

	a := &App{
		window:            window,
		theme:             th,
		colors:            newPalette(platform.ColorSchemeLight),
		service:           pdf.NewService(),
		thumbs:            thumbs,
		inspectCh:         make(chan inspectTaskResult, 1),
		thumbCh:           make(chan thumbnailTaskResult, 4),
		splitCh:           make(chan splitTaskResult, 1),
		boundaries:        map[int]bool{},
		thumbnails:        map[int]*thumbnailState{},
		thumbnailErrors:   map[int]string{},
		previewMode:       previewList,
		spanPresetEnabled: true,
		spanOptions:       []int{1, 2, 3, 5, 10, 20, 50, 100},
		themeMode:         themeModeSystem,
	}
	a.spanOptionBtns = make([]widget.Clickable, len(a.spanOptions))

	a.controlList.Axis = layout.Vertical
	a.pageList.Axis = layout.Vertical
	a.pathEditor.SingleLine = true
	a.spanEditor.SingleLine = true
	a.spanEditor.SetText("1")
	a.applyThemeMode(themeModeSystem)
	a.setStatus(statusInfo, "파일을 선택해 주세요", "PDF 파일을 선택하면 분석 후 페이지 사이 구분선을 직접 클릭해 분할할 수 있습니다.")

	return a
}

func (a *App) Run() error {
	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.ConfigEvent:
			a.decorations.Maximized = e.Config.Mode == app.Maximized
			if a.themeMode == themeModeSystem {
				a.applyThemeMode(themeModeSystem)
			}
		case app.FrameEvent:
			gtx := app.NewContext(&a.ops, e)
			a.handleAsyncResults()
			a.handleEditorEvents(gtx)
			a.handleActions(gtx)
			a.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) handleAsyncResults() {
	for {
		select {
		case res := <-a.inspectCh:
			if res.Err != nil {
				a.busy = false
				a.thumbLoading = false
				a.activeTask = ""
				a.pageCount = 0
				a.pages = nil
				a.pageOpenBtns = nil
				a.separatorBtns = nil
				a.boundaries = map[int]bool{}
				a.thumbnails = map[int]*thumbnailState{}
				a.thumbnailErrors = map[int]string{}
				a.outputPreview = nil
				a.completedOutputDir = ""
				a.selectedPage = 0
				a.setStatus(statusError, "분석 실패", res.Err.Error())
				continue
			}

			a.pageCount = res.Result.PageCount
			a.pages = res.Result.Pages
			a.outputDirectory = res.Result.OutputDir
			a.pageOpenBtns = make([]widget.Clickable, a.pageCount)
			a.separatorBtns = make([]widget.Clickable, max(0, a.pageCount-1))
			a.boundaries = map[int]bool{}
			a.thumbnails = map[int]*thumbnailState{}
			a.thumbnailErrors = map[int]string{}
			a.selectedPage = 0
			if a.spanPresetEnabled {
				a.applySpanPreset()
			} else {
				a.updateOutputPlan()
			}
			a.startThumbnailRender()
		case res := <-a.thumbCh:
			if res.InputPath != a.inputPath {
				continue
			}
			if res.Err != nil {
				a.busy = false
				a.thumbLoading = false
				a.activeTask = ""
				for page := res.StartPage; page <= res.EndPage; page++ {
					state := a.ensureThumbnailState(page)
					state.Loading = false
					state.Failed = true
					a.thumbnailErrors[page] = res.Err.Error()
				}
				a.setStatus(statusError, "미리보기 생성 실패", res.Err.Error())
				continue
			}
			for page := res.StartPage; page <= res.EndPage; page++ {
				state := a.ensureThumbnailState(page)
				state.Loading = false
				if op, ok := res.Images[page]; ok {
					state.Op = op
					state.Loaded = true
					state.Failed = false
					delete(a.thumbnailErrors, page)
				}
			}
			a.thumbLoadedPages += res.EndPage - res.StartPage + 1
			if a.thumbLoadedPages >= a.thumbTotalPages {
				a.busy = false
				a.thumbLoading = false
				a.activeTask = ""
				a.setStatus(statusSuccess, "분석 완료", fmt.Sprintf("총 %d페이지 미리보기를 준비했습니다. 구분선을 조정한 뒤 분할을 실행해 주세요.", a.pageCount))
			} else {
				a.setStatus(statusInfo, "미리보기 생성 중", fmt.Sprintf("%d / %d 페이지 썸네일을 준비했습니다.", a.thumbLoadedPages, a.thumbTotalPages))
			}
		case res := <-a.splitCh:
			a.busy = false
			a.activeTask = ""
			if res.Err != nil {
				a.setStatus(statusError, "분할 실패", res.Err.Error())
				continue
			}

			a.outputDirectory = res.Result.OutputDir
			a.outputPreview = shortNames(res.Result.Files, 3)
			a.completedOutputDir = res.Result.OutputDir
			a.setStatus(statusSuccess, "분할 완료", fmt.Sprintf("%d개 파일을 원본 폴더에 생성했습니다. 첫 파일: %s. 저장 위치 열기 버튼으로 바로 확인할 수 있습니다.", len(res.Result.Files), filepath.Base(res.Result.Files[0])))
		default:
			return
		}
	}
}

func (a *App) handleEditorEvents(gtx layout.Context) {
	for {
		event, ok := a.spanEditor.Update(gtx)
		if !ok {
			break
		}
		if _, changed := event.(widget.ChangeEvent); changed {
			cleaned := digitsOnly(a.spanEditor.Text())
			if cleaned != a.spanEditor.Text() {
				a.spanEditor.SetText(cleaned)
			}
			if a.pageCount > 0 {
				a.setStatus(statusInfo, "빠른 분할 단위 변경됨", "정수 단위를 입력하거나 선택한 뒤 자동 구분선 적용을 누르세요.")
			}
		}
	}

	for {
		_, ok := a.pathEditor.Update(gtx)
		if !ok {
			break
		}
	}
}

func (a *App) handleActions(gtx layout.Context) {
	if actions := a.decorations.Update(gtx); actions != 0 {
		a.window.Perform(actions)
	}

	if a.dropZoneBtn.Clicked(gtx) && !a.busy {
		path, err := platform.PickPDF()
		if err != nil {
			a.setStatus(statusError, "파일 선택 실패", err.Error())
			return
		}
		if path != "" {
			a.acceptInputFile(path)
		}
	}

	if a.analyzeBtn.Clicked(gtx) && !a.busy {
		path := a.currentInputPath()
		if path == "" {
			a.setStatus(statusError, "입력 필요", "먼저 PDF 파일을 선택해 주세요.")
			return
		}
		a.acceptInputFile(path)
	}

	if a.applySpanBtn.Clicked(gtx) && !a.busy {
		if !a.spanPresetEnabled {
			a.setStatus(statusInfo, "빠른 분할 꺼짐", "토글을 켜면 페이지 단위로 자동 구분선을 배치할 수 있습니다.")
			return
		}
		if a.pageCount == 0 {
			a.setStatus(statusError, "분석 필요", "먼저 PDF 분석을 완료해 주세요.")
			return
		}
		a.applySpanPreset()
		a.setStatus(statusInfo, "구분선 재배치 완료", fmt.Sprintf("현재 설정으로 %d개 파일이 생성됩니다.", a.segmentCount()))
	}

	if a.spanToggleBtn.Clicked(gtx) && !a.busy {
		a.spanPresetEnabled = !a.spanPresetEnabled
		if a.spanPresetEnabled {
			if a.pageCount > 0 {
				a.applySpanPreset()
				a.setStatus(statusInfo, "빠른 분할 켜짐", fmt.Sprintf("%d페이지 단위 자동 배치를 다시 적용했습니다.", max(1, a.currentSpanValue())))
			} else {
				a.setStatus(statusInfo, "빠른 분할 켜짐", "분석 후 빠른 분할 단위가 자동 적용됩니다.")
			}
		} else {
			a.boundaries = map[int]bool{}
			a.updateOutputPlan()
			a.setStatus(statusInfo, "빠른 분할 꺼짐", "자동 분할을 끄면서 현재 구분선도 모두 초기화했습니다.")
		}
	}

	if a.spanMenuBtn.Clicked(gtx) && !a.busy {
		a.spanDropdownOpen = !a.spanDropdownOpen
	}
	for i := range a.spanOptionBtns {
		if a.spanOptionBtns[i].Clicked(gtx) && !a.busy {
			a.spanEditor.SetText(strconv.Itoa(a.spanOptions[i]))
			a.spanDropdownOpen = false
			if a.pageCount > 0 {
				a.setStatus(statusInfo, "빠른 분할 단위 선택됨", fmt.Sprintf("%d페이지 단위가 선택되었습니다. 적용 버튼으로 자동 구분선을 다시 배치할 수 있습니다.", a.spanOptions[i]))
			}
		}
	}

	if a.splitBtn.Clicked(gtx) && !a.busy {
		a.startSplit()
	}
	if a.openOutputBtn.Clicked(gtx) && !a.busy {
		if strings.TrimSpace(a.completedOutputDir) == "" {
			a.setStatus(statusError, "출력 위치 없음", "분할이 완료된 뒤 저장 위치를 열 수 있습니다.")
			return
		}
		if err := platform.RevealDirectory(a.completedOutputDir); err != nil {
			a.setStatus(statusError, "출력 폴더 열기 실패", err.Error())
			return
		}
		a.setStatus(statusInfo, "출력 폴더 열기", "저장된 파일 위치를 탐색기에서 열었습니다.")
	}

	if a.listModeBtn.Clicked(gtx) && !a.thumbLoading {
		a.previewMode = previewList
	}
	if a.gridModeBtn.Clicked(gtx) && !a.thumbLoading {
		a.previewMode = previewGrid
	}
	if a.themeSystemBtn.Clicked(gtx) {
		a.applyThemeMode(themeModeSystem)
	}
	if a.themeLightBtn.Clicked(gtx) {
		a.applyThemeMode(themeModeLight)
	}
	if a.themeDarkBtn.Clicked(gtx) {
		a.applyThemeMode(themeModeDark)
	}
	if a.closePreviewBtn.Clicked(gtx) {
		a.selectedPage = 0
	}

	for i := range a.pageOpenBtns {
		for a.pageOpenBtns[i].Clicked(gtx) {
			a.selectedPage = i + 1
		}
	}

	for i := range a.separatorBtns {
		for a.separatorBtns[i].Clicked(gtx) {
			a.toggleBoundary(i + 1)
		}
	}
}

func (a *App) acceptInputFile(path string) {
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		a.setStatus(statusError, "잘못된 파일", "PDF 파일만 선택할 수 있습니다.")
		return
	}

	a.inputPath = path
	a.pathEditor.SetText(path)
	a.pageCount = 0
	a.pages = nil
	a.pageOpenBtns = nil
	a.separatorBtns = nil
	a.boundaries = map[int]bool{}
	a.outputDirectory = filepath.Dir(path)
	a.outputPreview = nil
	a.completedOutputDir = ""
	a.thumbnails = map[int]*thumbnailState{}
	a.thumbnailErrors = map[int]string{}
	a.selectedPage = 0
	a.setStatus(statusInfo, "파일 선택됨", "선택한 PDF를 분석하고 있습니다.")
	a.startInspect(path)
}

func (a *App) startInspect(path string) {
	a.busy = true
	a.thumbLoading = false
	a.thumbLoadedPages = 0
	a.thumbTotalPages = 0
	a.activeTask = "PDF 분석 중"
	a.setStatus(statusInfo, "분석 중", "페이지 수와 페이지 비율을 계산하고 있습니다.")
	a.window.Invalidate()

	go func() {
		result, err := a.service.Inspect(path)
		a.inspectCh <- inspectTaskResult{Result: result, Err: err}
		a.window.Invalidate()
	}()
}

func (a *App) startSplit() {
	if strings.TrimSpace(a.inputPath) == "" {
		a.setStatus(statusError, "입력 필요", "먼저 PDF 파일을 선택해 주세요.")
		return
	}
	if a.pageCount == 0 {
		a.setStatus(statusError, "분석 필요", "PDF 분석이 아직 완료되지 않았습니다.")
		return
	}
	if a.thumbLoading {
		a.setStatus(statusError, "미리보기 생성 중", "페이지 미리보기가 아직 준비되지 않았습니다. 잠시 후 다시 시도해 주세요.")
		return
	}

	boundaries := a.sortedBoundaries()
	a.busy = true
	a.activeTask = "PDF 분할 중"
	a.completedOutputDir = ""
	a.setStatus(statusInfo, "분할 실행 중", "선택한 구분선을 기준으로 원본 폴더에 차번 파일명을 생성하고 있습니다.")
	a.window.Invalidate()

	go func() {
		result, err := a.service.Split(pdf.SplitRequest{
			InputPath:    a.inputPath,
			Boundaries:   boundaries,
			RequireSplit: a.pageCount > 1,
		})
		a.splitCh <- splitTaskResult{Result: result, Err: err}
		a.window.Invalidate()
	}()
}

func (a *App) applySpanPreset() {
	span, err := a.parseSpan()
	if err != nil {
		a.setStatus(statusError, "입력 오류", err.Error())
		return
	}

	a.boundaries = map[int]bool{}
	for _, boundary := range pdf.BuildBoundariesBySpan(a.pageCount, span) {
		a.boundaries[boundary] = true
	}
	a.updateOutputPlan()
}

func (a *App) parseSpan() (int, error) {
	value := strings.TrimSpace(a.spanEditor.Text())
	if value == "" {
		return 0, fmt.Errorf("빠른 분할 단위는 숫자로 입력해 주세요")
	}
	span, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("빠른 분할 단위는 숫자로 입력해 주세요")
	}
	if span < 1 {
		return 0, fmt.Errorf("빠른 분할 단위는 1 이상이어야 합니다")
	}
	return span, nil
}

func (a *App) currentSpanValue() int {
	span, err := a.parseSpan()
	if err != nil {
		return 0
	}
	return span
}

func (a *App) toggleBoundary(page int) {
	if a.boundaries[page] {
		delete(a.boundaries, page)
	} else {
		a.boundaries[page] = true
	}
	a.updateOutputPlan()
	a.setStatus(statusInfo, "구분선 변경됨", fmt.Sprintf("현재 %d개 파일로 분할됩니다.", a.segmentCount()))
}

func (a *App) updateOutputPlan() {
	if a.pageCount == 0 || strings.TrimSpace(a.inputPath) == "" {
		a.outputPreview = nil
		return
	}

	plan, err := a.service.PlanOutput(a.inputPath, a.segmentCount())
	if err != nil {
		a.outputPreview = nil
		return
	}

	a.outputDirectory = plan.Directory
	a.outputPreview = shortNames(plan.Files, 3)
}

func (a *App) sortedBoundaries() []int {
	boundaries := make([]int, 0, len(a.boundaries))
	for page := range a.boundaries {
		boundaries = append(boundaries, page)
	}
	sort.Ints(boundaries)
	return boundaries
}

func (a *App) segmentCount() int {
	if a.pageCount == 0 {
		return 0
	}
	return len(a.boundaries) + 1
}

func (a *App) splitSummary() string {
	if a.pageCount == 0 {
		return "-"
	}

	ranges, err := pdf.BuildRanges(a.pageCount, a.sortedBoundaries())
	if err != nil {
		return "-"
	}

	parts := make([]string, 0, min(len(ranges), 4))
	for i, pageRange := range ranges {
		if i == 4 {
			parts = append(parts, "...")
			break
		}
		parts = append(parts, fmt.Sprintf("%d-%d", pageRange.From, pageRange.Thru))
	}

	return strings.Join(parts, ", ")
}

func (a *App) setStatus(tone statusTone, title, detail string) {
	a.statusTone = tone
	a.statusTitle = title
	a.statusDetail = detail
}

func (a *App) applyThemeMode(mode themeMode) {
	a.themeMode = mode

	scheme := platform.ColorSchemeLight
	switch mode {
	case themeModeSystem:
		scheme = platform.DetectSystemColorScheme()
	case themeModeDark:
		scheme = platform.ColorSchemeDark
	}

	a.colors = newPalette(scheme)
	a.theme.Palette = material.Palette{
		Bg:         a.colors.Surface,
		Fg:         a.colors.Text,
		ContrastBg: a.colors.Primary,
		ContrastFg: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
	}
	a.window.Invalidate()
}

func (a *App) currentInputPath() string {
	path := strings.TrimSpace(a.pathEditor.Text())
	if path != "" {
		return path
	}
	return strings.TrimSpace(a.inputPath)
}

func (a *App) ensureThumbnailState(page int) *thumbnailState {
	state, ok := a.thumbnails[page]
	if !ok {
		state = &thumbnailState{}
		a.thumbnails[page] = state
	}
	return state
}

func (a *App) startThumbnailRender() {
	if a.pageCount == 0 || strings.TrimSpace(a.inputPath) == "" {
		a.busy = false
		a.thumbLoading = false
		a.activeTask = ""
		return
	}
	if a.thumbs == nil {
		a.busy = false
		a.thumbLoading = false
		a.activeTask = ""
		a.setStatus(statusError, "미리보기 생성 실패", "PDF 썸네일 렌더러를 찾을 수 없습니다.")
		return
	}

	const batchSize = 6
	a.busy = true
	a.thumbLoading = true
	a.thumbLoadedPages = 0
	a.thumbTotalPages = a.pageCount
	a.activeTask = "썸네일 생성 중"
	a.setStatus(statusInfo, "미리보기 생성 중", fmt.Sprintf("%d페이지 썸네일을 준비하고 있습니다.", a.pageCount))
	a.window.Invalidate()

	inputPath := a.inputPath
	for page := 1; page <= a.pageCount; page++ {
		state := a.ensureThumbnailState(page)
		state.Loading = true
		state.Loaded = false
		state.Failed = false
	}

	go func() {
		for start := 1; start <= a.pageCount; start += batchSize {
			end := min(a.pageCount, start+batchSize-1)
			images, err := a.thumbs.LoadBatch(inputPath, start, end, 30)
			ops := make(map[int]paint.ImageOp, len(images))
			for page, img := range images {
				ops[page] = paint.NewImageOp(img)
			}
			a.thumbCh <- thumbnailTaskResult{
				InputPath: inputPath,
				StartPage: start,
				EndPage:   end,
				Images:    ops,
				Err:       err,
			}
			a.window.Invalidate()
			if err != nil {
				return
			}
		}
	}()
}

func (a *App) fileButtonLabel() string {
	if a.busy && a.activeTask == "PDF 분석 중" {
		return "분석 중..."
	}
	if a.thumbLoading {
		return "미리보기 생성 중..."
	}
	if a.inputPath == "" {
		return "PDF 파일 선택"
	}
	return "다른 PDF 선택"
}

func (a *App) applyButtonLabel() string {
	if a.thumbLoading {
		return "미리보기 생성 중"
	}
	if a.busy {
		return "대기 중"
	}
	if !a.spanPresetEnabled {
		return "기능 꺼짐"
	}
	return "자동 구분선 적용"
}

func (a *App) analyzeButtonLabel() string {
	if a.busy && a.activeTask == "PDF 분석 중" {
		return "분석 중..."
	}
	if a.thumbLoading {
		return "미리보기 생성 중..."
	}
	return "입력 경로 분석"
}

func (a *App) splitButtonLabel() string {
	if a.busy && a.activeTask == "PDF 분할 중" {
		return "분할 중..."
	}
	if a.thumbLoading {
		return "미리보기 생성 중..."
	}
	return "PDF 분할 실행"
}

func (a *App) spanDropdownLabel() string {
	if span := a.currentSpanValue(); span > 0 {
		return fmt.Sprintf("선택 %d", span)
	}
	return "빠른 선택"
}

func (a *App) selectedPageMeta() (pdf.PageMeta, bool) {
	if a.selectedPage < 1 || a.selectedPage > len(a.pages) {
		return pdf.PageMeta{}, false
	}
	return a.pages[a.selectedPage-1], true
}

func shortNames(paths []string, limit int) []string {
	if len(paths) <= limit {
		names := make([]string, 0, len(paths))
		for _, path := range paths {
			names = append(names, filepath.Base(path))
		}
		return names
	}

	names := make([]string, 0, limit+1)
	for i := 0; i < limit; i++ {
		names = append(names, filepath.Base(paths[i]))
	}
	names = append(names, "...")
	return names
}

func newPalette(scheme platform.ColorScheme) palette {
	if scheme == platform.ColorSchemeDark {
		return palette{
			Page:         color.NRGBA{R: 0x0B, G: 0x12, B: 0x20, A: 0xFF},
			Surface:      color.NRGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xFF},
			SurfaceMuted: color.NRGBA{R: 0x17, G: 0x21, B: 0x32, A: 0xFF},
			SurfaceAlt:   color.NRGBA{R: 0x22, G: 0x30, B: 0x46, A: 0xFF},
			Primary:      color.NRGBA{R: 0x5B, G: 0x8F, B: 0xB9, A: 0xFF},
			PrimarySoft:  color.NRGBA{R: 0x1E, G: 0x31, B: 0x45, A: 0xFF},
			Accent:       color.NRGBA{R: 0x7D, G: 0xB3, B: 0xD1, A: 0xFF},
			Text:         color.NRGBA{R: 0xE5, G: 0xEC, B: 0xF6, A: 0xFF},
			TextMuted:    color.NRGBA{R: 0x97, G: 0xA3, B: 0xB6, A: 0xFF},
			Success:      color.NRGBA{R: 0x4D, G: 0xAA, B: 0x84, A: 0xFF},
			SuccessSoft:  color.NRGBA{R: 0x12, G: 0x28, B: 0x1F, A: 0xFF},
			Error:        color.NRGBA{R: 0xE4, G: 0x7B, B: 0x72, A: 0xFF},
			ErrorSoft:    color.NRGBA{R: 0x33, G: 0x16, B: 0x18, A: 0xFF},
			Border:       color.NRGBA{R: 0x32, G: 0x41, B: 0x55, A: 0xFF},
			Hover:        color.NRGBA{R: 0x2A, G: 0x3A, B: 0x50, A: 0xFF},
			Pressed:      color.NRGBA{R: 0x3B, G: 0x51, B: 0x6C, A: 0xFF},
		}
	}

	return palette{
		Page:         color.NRGBA{R: 0xF7, G: 0xF8, B: 0xFA, A: 0xFF},
		Surface:      color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		SurfaceMuted: color.NRGBA{R: 0xF2, G: 0xF4, B: 0xF7, A: 0xFF},
		SurfaceAlt:   color.NRGBA{R: 0xE6, G: 0xEB, B: 0xF2, A: 0xFF},
		Primary:      color.NRGBA{R: 0x1D, G: 0x35, B: 0x57, A: 0xFF},
		PrimarySoft:  color.NRGBA{R: 0xE7, G: 0xED, B: 0xF5, A: 0xFF},
		Accent:       color.NRGBA{R: 0x2C, G: 0x7D, B: 0xA0, A: 0xFF},
		Text:         color.NRGBA{R: 0x10, G: 0x17, B: 0x28, A: 0xFF},
		TextMuted:    color.NRGBA{R: 0x66, G: 0x70, B: 0x85, A: 0xFF},
		Success:      color.NRGBA{R: 0x1F, G: 0x6F, B: 0x50, A: 0xFF},
		SuccessSoft:  color.NRGBA{R: 0xE7, G: 0xF5, B: 0xEE, A: 0xFF},
		Error:        color.NRGBA{R: 0xB4, G: 0x23, B: 0x18, A: 0xFF},
		ErrorSoft:    color.NRGBA{R: 0xFE, G: 0xEC, B: 0xEB, A: 0xFF},
		Border:       color.NRGBA{R: 0xD0, G: 0xD5, B: 0xDD, A: 0xFF},
		Hover:        color.NRGBA{R: 0xE1, G: 0xE8, B: 0xF0, A: 0xFF},
		Pressed:      color.NRGBA{R: 0xCD, G: 0xD8, B: 0xE6, A: 0xFF},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func digitsOnly(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
