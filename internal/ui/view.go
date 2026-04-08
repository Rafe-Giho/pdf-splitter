package ui

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strconv"
	"strings"

	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"pdfsplitter/internal/pdf"
)

func (a *App) layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, a.colors.Page)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.layoutWindowBar),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(a.layoutHeader),
					layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
					layout.Flexed(1, a.layoutBody),
				)
			})
		}),
	)
}

func (a *App) layoutWindowBar(gtx layout.Context) layout.Dimensions {
	return a.box(gtx, a.colors.Primary, layout.Inset{
		Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(12), Right: unit.Dp(8),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return a.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(a.titleMark),
						layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body1(a.theme, "PDF Splitter")
							label.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
							return label.Layout(gtx)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.themeModeButton(gtx, &a.themeSystemBtn, "시스템", a.themeMode == themeModeSystem)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.themeModeButton(gtx, &a.themeLightBtn, "라이트", a.themeMode == themeModeLight)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.themeModeButton(gtx, &a.themeDarkBtn, "다크", a.themeMode == themeModeDark)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.windowControlButton(gtx, a.decorations.Clickable(system.ActionMinimize), "_", false)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "[ ]"
						if a.decorations.Maximized {
							label = "[_]"
						}
						return a.windowControlButton(gtx, a.decorations.Clickable(system.ActionMaximize), label, false)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.windowControlButton(gtx, a.decorations.Clickable(system.ActionClose), "X", true)
					}),
				)
			}),
		)
	})
}

func (a *App) titleMark(gtx layout.Context) layout.Dimensions {
	size := gtx.Dp(unit.Dp(22))
	rect := image.Rect(0, 0, size, size)
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x30}, clip.UniformRRect(rect, gtx.Dp(unit.Dp(7))).Op(gtx.Ops))
	page := image.Rect(gtx.Dp(unit.Dp(5)), gtx.Dp(unit.Dp(4)), gtx.Dp(unit.Dp(17)), gtx.Dp(unit.Dp(18)))
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, clip.UniformRRect(page, gtx.Dp(unit.Dp(4))).Op(gtx.Ops))
	split := image.Rect(gtx.Dp(unit.Dp(10)), gtx.Dp(unit.Dp(7)), gtx.Dp(unit.Dp(12)), gtx.Dp(unit.Dp(16)))
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0x53, G: 0xD2, B: 0xF4, A: 0xFF}, clip.UniformRRect(split, gtx.Dp(unit.Dp(1))).Op(gtx.Ops))
	return layout.Dimensions{Size: image.Pt(size, size)}
}

func (a *App) layoutHeader(gtx layout.Context) layout.Dimensions {
	return a.box(gtx, a.colors.Surface, layout.Inset{
		Top: unit.Dp(18), Bottom: unit.Dp(18), Left: unit.Dp(20), Right: unit.Dp(20),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H5(a.theme, "PDF 분할기")
						title.Color = a.colors.Text
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Label(a.theme, unit.Sp(14), "파일을 선택하고 분석한 뒤, 페이지 사이 막대를 눌러 원하는 위치에서 분할합니다.")
						label.Color = a.colors.TextMuted
						return label.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.metricBox(gtx, "총 페이지", valueOrDash(a.pageCount))
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.metricBox(gtx, "분할 파일", valueOrDash(a.segmentCount()))
					}),
				)
			}),
		)
	})
}

func (a *App) layoutBody(gtx layout.Context) layout.Dimensions {
	stacked := gtx.Constraints.Max.X < gtx.Dp(980)
	if stacked {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(0.42, a.layoutControlPanel),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			layout.Flexed(0.58, a.layoutPreviewPanel),
		)
	}

	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(360))
			return a.layoutControlPanel(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
		layout.Flexed(1, a.layoutPreviewPanel),
	)
}

func (a *App) layoutControlPanel(gtx layout.Context) layout.Dimensions {
	return a.box(gtx, a.colors.Surface, layout.UniformInset(0), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return material.List(a.theme, &a.controlList).Layout(gtx, 4, func(gtx layout.Context, index int) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							switch index {
							case 0:
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return a.sectionTitle(gtx, "입력 파일", "PDF 선택")
									}),
									layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
									layout.Rigid(a.layoutFilePicker),
								)
							case 1:
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return a.sectionTitle(gtx, "빠른 분할 단위", "자동 구분선 배치")
									}),
									layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
									layout.Rigid(a.layoutSpanControls),
								)
							case 2:
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return a.sectionTitle(gtx, "출력", "원본과 같은 폴더에 차번 저장")
									}),
									layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
									layout.Rigid(a.layoutOutputInfo),
								)
							default:
								return a.layoutStatusCard(gtx)
							}
						})
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.primaryButton(gtx, &a.splitBtn, a.splitButtonLabel())
				}),
			)
		})
	})
}

func (a *App) layoutFilePicker(gtx layout.Context) layout.Dimensions {
	name := "-"
	if a.inputPath != "" {
		name = filepath.Base(a.inputPath)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.secondaryButton(gtx, &a.dropZoneBtn, a.fileButtonLabel())
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.secondaryButton(gtx, &a.analyzeBtn, a.analyzeButtonLabel())
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.inputBox(gtx, &a.pathEditor, `PDF 경로를 붙여넣거나 "PDF 파일 선택"을 누르세요`)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
				Top: unit.Dp(14), Bottom: unit.Dp(14), Left: unit.Dp(14), Right: unit.Dp(14),
			}, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Label(a.theme, unit.Sp(12), "선택 파일")
						label.Color = a.colors.TextMuted
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(a.theme, name)
						label.Color = a.colors.Text
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						hint := material.Label(a.theme, unit.Sp(13), "탐색기 파일 선택창만 사용합니다. 별도 콘솔 창은 띄우지 않습니다.")
						hint.Color = a.colors.TextMuted
						return hint.Layout(gtx)
					}),
				)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			full := "-"
			if a.inputPath != "" {
				full = a.inputPath
			}
			return a.infoRow(gtx, "선택 경로", full)
		}),
	)
}

func (a *App) layoutSpanControls(gtx layout.Context) layout.Dimensions {
	toggleLabel := "빠른 분할 꺼짐"
	if a.spanPresetEnabled {
		toggleLabel = "빠른 분할 켜짐"
	}

	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.modeButton(gtx, &a.spanToggleBtn, toggleLabel, a.spanPresetEnabled)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.inputBox(gtx, &a.spanEditor, "숫자만 입력")
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.secondaryButton(gtx, &a.spanMenuBtn, a.spanDropdownLabel())
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.secondaryButton(gtx, &a.applySpanBtn, a.applyButtonLabel())
				}),
			)
		}),
	}
	if a.spanDropdownOpen {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(a.layoutSpanOptions),
		)
	}
	children = append(children,
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			hint := "꺼져 있으면 수동 구분선만 유지합니다."
			if a.spanPresetEnabled {
				hint = "정수만 입력하거나 드롭다운에서 선택해 자동 구분선을 배치할 수 있습니다."
			}
			label := material.Label(a.theme, unit.Sp(12), hint)
			label.Color = a.colors.TextMuted
			return label.Layout(gtx)
		}),
	)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (a *App) layoutSpanOptions(gtx layout.Context) layout.Dimensions {
	cols := 4
	rows := (len(a.spanOptions) + cols - 1) / cols

	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(12), Left: unit.Dp(12), Right: unit.Dp(12),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(12), "추천 단위")
				label.Color = a.colors.TextMuted
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, makeSpanOptionRows(a, gtx, rows, cols)...)
			}),
		)
	})
}

func makeSpanOptionRows(a *App, gtx layout.Context, rows, cols int) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, rows*2-1)
	for row := 0; row < rows; row++ {
		if row > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout))
		}
		rowIndex := row
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			items := make([]layout.FlexChild, 0, cols*2-1)
			for col := 0; col < cols; col++ {
				index := rowIndex*cols + col
				if col > 0 {
					items = append(items, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout))
				}
				if index >= len(a.spanOptions) {
					items = append(items, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					}))
					continue
				}
				option := a.spanOptions[index]
				button := &a.spanOptionBtns[index]
				items = append(items, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.modeButton(gtx, button, strconv.Itoa(option), a.currentSpanValue() == option)
				}))
			}
			return layout.Flex{}.Layout(gtx, items...)
		}))
	}
	return children
}

func (a *App) layoutOutputInfo(gtx layout.Context) layout.Dimensions {
	dir := "-"
	if a.outputDirectory != "" {
		dir = a.outputDirectory
	}

	names := "-"
	if len(a.outputPreview) > 0 {
		names = strings.Join(a.outputPreview, ", ")
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.infoRow(gtx, "출력 폴더", dir)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.infoRow(gtx, "파일명 예시", names)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.infoRow(gtx, "현재 구간", a.splitSummary())
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if strings.TrimSpace(a.completedOutputDir) == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.secondaryButton(gtx, &a.openOutputBtn, "저장 위치 열기")
			})
		}),
	)
}

func (a *App) layoutStatusCard(gtx layout.Context) layout.Dimensions {
	bg := a.colors.SurfaceMuted
	if a.statusTone == statusError {
		bg = a.colors.ErrorSoft
	}
	if a.statusTone == statusSuccess {
		bg = a.colors.SuccessSoft
	}

	return a.box(gtx, bg, layout.Inset{
		Top: unit.Dp(14), Bottom: unit.Dp(14), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(12), a.statusTitle)
				label.Color = a.statusColor()
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(13), a.statusDetail)
				label.Color = a.colors.Text
				return label.Layout(gtx)
			}),
		)
	})
}

func (a *App) layoutPreviewPanel(gtx layout.Context) layout.Dimensions {
	return a.box(gtx, a.colors.Surface, layout.Inset{
		Top: unit.Dp(18), Bottom: unit.Dp(18), Left: unit.Dp(18), Right: unit.Dp(18),
	}, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return a.sectionTitle(gtx, "페이지 구분선", "페이지 사이 막대를 눌러 분할 위치를 켜거나 끕니다.")
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return a.modeButton(gtx, &a.listModeBtn, "리스트", a.previewMode == previewList)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return a.modeButton(gtx, &a.gridModeBtn, "격자", a.previewMode == previewGrid)
							}),
						)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		}
		if a.selectedPage > 0 && !a.thumbLoading {
			children = append(children,
				layout.Rigid(a.layoutSelectedPagePanel),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			)
		}
		children = append(children, layout.Flexed(1, a.layoutPageList))
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

func (a *App) layoutPageList(gtx layout.Context) layout.Dimensions {
	if a.thumbLoading {
		return a.loadingBox(gtx)
	}
	if a.busy && len(a.pages) == 0 {
		return a.emptyBox(gtx, "PDF를 분석하고 있습니다.", "페이지 수를 계산하는 동안 잠시 기다려 주세요.")
	}
	if len(a.pages) == 0 {
		return a.emptyBox(gtx, "아직 분석된 PDF가 없습니다.", "왼쪽에서 파일을 선택한 뒤 분석을 완료해 주세요.")
	}
	if a.previewMode == previewGrid {
		return a.layoutPageGrid(gtx)
	}

	items := len(a.pages)*2 - 1
	return material.List(a.theme, &a.pageList).Layout(gtx, items, func(gtx layout.Context, index int) layout.Dimensions {
		if index%2 == 0 {
			page := a.pages[index/2]
			return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.pageRow(gtx, page)
			})
		}

		page := index/2 + 1
		return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return a.separatorRow(gtx, page, a.boundaries[page], &a.separatorBtns[page-1])
		})
	})
}

func (a *App) pageRow(gtx layout.Context, page pdf.PageMeta) layout.Dimensions {
	bg := a.colors.SurfaceMuted
	if a.selectedPage == page.Number {
		bg = a.colors.PrimarySoft
	}
	return a.box(gtx, bg, layout.Inset{
		Top: unit.Dp(14), Bottom: unit.Dp(14), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.pagePreviewButton(gtx, page, unit.Dp(72), unit.Dp(96), unit.Dp(132))
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(a.theme, fmt.Sprintf("페이지 %d", page.Number))
						label.Color = a.colors.Text
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						orientation := "세로형"
						if page.Ratio < 1 {
							orientation = "가로형"
						}
						label := material.Label(a.theme, unit.Sp(13), fmt.Sprintf("%s / 비율 %.2f", orientation, page.Ratio))
						label.Color = a.colors.TextMuted
						return label.Layout(gtx)
					}),
				)
			}),
		)
	})
}

func (a *App) layoutPageGrid(gtx layout.Context) layout.Dimensions {
	cols := 2
	if gtx.Constraints.Max.X >= gtx.Dp(900) {
		cols = 4
	} else if gtx.Constraints.Max.X >= gtx.Dp(640) {
		cols = 3
	}

	rows := (len(a.pages) + cols - 1) / cols
	return material.List(a.theme, &a.pageList).Layout(gtx, rows, func(gtx layout.Context, row int) layout.Dimensions {
		return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := make([]layout.FlexChild, 0, cols*2-1)
			for col := 0; col < cols; col++ {
				index := row*cols + col
				if col > 0 {
					children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout))
				}
				if index >= len(a.pages) {
					children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					}))
					continue
				}
				page := a.pages[index]
				children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.pageGridCard(gtx, page)
				}))
			}
			return layout.Flex{}.Layout(gtx, children...)
		})
	})
}

func (a *App) pageGridCard(gtx layout.Context, page pdf.PageMeta) layout.Dimensions {
	bg := a.colors.SurfaceMuted
	if a.selectedPage == page.Number {
		bg = a.colors.PrimarySoft
	}
	return a.box(gtx, bg, layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(12), Left: unit.Dp(12), Right: unit.Dp(12),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.pagePreviewButton(gtx, page, unit.Dp(120), unit.Dp(124), unit.Dp(156))
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(a.theme, fmt.Sprintf("페이지 %d", page.Number))
				label.Color = a.colors.Text
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if page.Number >= a.pageCount {
					label := material.Label(a.theme, unit.Sp(12), "마지막 페이지")
					label.Color = a.colors.TextMuted
					return label.Layout(gtx)
				}
				return a.modeButton(gtx, &a.separatorBtns[page.Number-1], a.boundaryButtonLabel(page.Number), a.boundaries[page.Number])
			}),
		)
	})
}

func (a *App) layoutSelectedPagePanel(gtx layout.Context) layout.Dimensions {
	page, ok := a.selectedPageMeta()
	if !ok {
		return layout.Dimensions{}
	}

	stacked := gtx.Constraints.Max.X < gtx.Dp(760)
	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(16), Bottom: unit.Dp(16), Left: unit.Dp(16), Right: unit.Dp(16),
	}, func(gtx layout.Context) layout.Dimensions {
		header := []layout.FlexChild{
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H6(a.theme, fmt.Sprintf("페이지 %d 확대 보기", page.Number))
						title.Color = a.colors.Text
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						orientation := "세로형"
						if page.Ratio < 1 {
							orientation = "가로형"
						}
						label := material.Label(a.theme, unit.Sp(13), fmt.Sprintf("%s / 비율 %.2f", orientation, page.Ratio))
						label.Color = a.colors.TextMuted
						return label.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.secondaryButton(gtx, &a.closePreviewBtn, "닫기")
			}),
		}

		body := func(gtx layout.Context) layout.Dimensions {
			if stacked {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return a.pagePreviewSized(gtx, page, unit.Dp(220), unit.Dp(260), unit.Dp(360))
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.infoRow(gtx, "선택 페이지", fmt.Sprintf("%d", page.Number))
					}),
				)
			}

			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.pagePreviewSized(gtx, page, unit.Dp(220), unit.Dp(260), unit.Dp(360))
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(18)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.infoRow(gtx, "선택 페이지", fmt.Sprintf("%d", page.Number))
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Label(a.theme, unit.Sp(13), "아래 목록에서 다른 페이지를 누르면 여기서 바로 바뀝니다.")
							label.Color = a.colors.TextMuted
							return label.Layout(gtx)
						}),
					)
				}),
			)
		}

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx, header...)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			layout.Rigid(body),
		)
	})
}

func (a *App) separatorRow(gtx layout.Context, page int, active bool, clickable *widget.Clickable) layout.Dimensions {
	bg := a.colors.SurfaceMuted
	textColor := a.colors.TextMuted
	labelText := fmt.Sprintf("%d페이지 뒤에 분할선 추가", page)
	if active {
		bg = a.colors.PrimarySoft
		textColor = a.colors.Primary
		labelText = fmt.Sprintf("%d페이지 뒤에서 분할됨", page)
	}
	if clickable.Hovered() {
		bg = a.colors.SurfaceAlt
	}
	if clickable.Pressed() {
		bg = a.colors.Pressed
	}
	if active && clickable.Hovered() {
		bg = a.colors.Hover
	}

	return a.clickBox(gtx, clickable, bg, layout.Inset{
		Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(a.theme, unit.Sp(13), labelText)
		label.Color = textColor
		return label.Layout(gtx)
	})
}

func (a *App) pagePreview(gtx layout.Context, page pdf.PageMeta) layout.Dimensions {
	return a.pagePreviewSized(gtx, page, unit.Dp(72), unit.Dp(96), unit.Dp(132))
}

func (a *App) pagePreviewButton(gtx layout.Context, page pdf.PageMeta, widthDp, minHeightDp, maxHeightDp unit.Dp) layout.Dimensions {
	bg := a.colors.Page
	if a.selectedPage == page.Number {
		bg = a.colors.PrimarySoft
	}
	return a.clickBox(gtx, &a.pageOpenBtns[page.Number-1], bg, layout.UniformInset(0), func(gtx layout.Context) layout.Dimensions {
		return a.pagePreviewSized(gtx, page, widthDp, minHeightDp, maxHeightDp)
	})
}

func (a *App) pagePreviewSized(gtx layout.Context, page pdf.PageMeta, widthDp, minHeightDp, maxHeightDp unit.Dp) layout.Dimensions {
	width := gtx.Dp(widthDp)
	height := int(float32(width) * page.Ratio)
	if height < gtx.Dp(minHeightDp) {
		height = gtx.Dp(minHeightDp)
	}
	if height > gtx.Dp(maxHeightDp) {
		height = gtx.Dp(maxHeightDp)
	}

	size := image.Pt(width, height)
	state := a.ensureThumbnailState(page.Number)

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rect := image.Rectangle{Max: size}
			paint.FillShape(gtx.Ops, a.colors.Page, clip.Rect(rect).Op())
			inner := image.Rect(2, 2, width-2, height-2)
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, clip.Rect(inner).Op())
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(size)
			switch {
			case state.Loaded:
				imageWidget := widget.Image{
					Src:      state.Op,
					Fit:      widget.Contain,
					Position: layout.Center,
					Scale:    1.0 / gtx.Metric.PxPerDp,
				}
				return layout.UniformInset(unit.Dp(3)).Layout(gtx, imageWidget.Layout)
			case state.Failed:
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Label(a.theme, unit.Sp(11), "썸네일 실패")
					label.Color = a.colors.TextMuted
					return label.Layout(gtx)
				})
			default:
				textValue := "미리보기 생성 중"
				if !state.Loading {
					textValue = "대기 중"
				}
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Label(a.theme, unit.Sp(11), textValue)
					label.Color = a.colors.TextMuted
					return label.Layout(gtx)
				})
			}
		}),
	)
}

func (a *App) loadingBox(gtx layout.Context) layout.Dimensions {
	progressText := "준비 중"
	if a.thumbTotalPages > 0 {
		progressText = fmt.Sprintf("%d / %d 페이지", a.thumbLoadedPages, a.thumbTotalPages)
	}

	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(28), Bottom: unit.Dp(28), Left: unit.Dp(28), Right: unit.Dp(28),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(520))
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(a.theme, "미리보기 굽는 중")
					title.Color = a.colors.Text
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Label(a.theme, unit.Sp(14), "파일을 한 번에 정리해서 보여주고 있습니다. 첫 로딩만 잠시 기다려 주세요.")
					label.Color = a.colors.TextMuted
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.progressBar(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(a.theme, progressText)
					label.Color = a.colors.Primary
					return label.Layout(gtx)
				}),
			)
		})
	})
}

func (a *App) progressBar(gtx layout.Context) layout.Dimensions {
	width := min(gtx.Constraints.Max.X, gtx.Dp(360))
	if width <= 0 {
		width = gtx.Dp(360)
	}
	height := gtx.Dp(unit.Dp(10))
	size := image.Pt(width, height)

	base := image.Rectangle{Max: size}
	paint.FillShape(gtx.Ops, a.colors.SurfaceAlt, clip.Rect(base).Op())

	fillWidth := width / 8
	if a.thumbTotalPages > 0 {
		fillWidth = width * a.thumbLoadedPages / a.thumbTotalPages
	}
	if fillWidth < width/10 {
		fillWidth = width / 10
	}
	if fillWidth > width {
		fillWidth = width
	}
	fill := image.Rect(0, 0, fillWidth, height)
	paint.FillShape(gtx.Ops, a.colors.Primary, clip.Rect(fill).Op())

	return layout.Dimensions{Size: size}
}

func (a *App) emptyBox(gtx layout.Context, titleText, detail string) layout.Dimensions {
	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H6(a.theme, titleText)
				title.Color = a.colors.Text
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(14), detail)
				label.Color = a.colors.TextMuted
				return label.Layout(gtx)
			}),
		)
	})
}

func (a *App) sectionTitle(gtx layout.Context, titleText, subtitle string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := material.H6(a.theme, titleText)
			title.Color = a.colors.Text
			return title.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(a.theme, unit.Sp(13), subtitle)
			label.Color = a.colors.TextMuted
			return label.Layout(gtx)
		}),
	)
}

func (a *App) metricBox(gtx layout.Context, name, value string) layout.Dimensions {
	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(12), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(12), name)
				label.Color = a.colors.TextMuted
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(a.theme, value)
				label.Color = a.colors.Text
				return label.Layout(gtx)
			}),
		)
	})
}

func (a *App) infoRow(gtx layout.Context, name, value string) layout.Dimensions {
	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(12), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(12), name)
				label.Color = a.colors.TextMuted
				return label.Layout(gtx)
			}),
			layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(13), value)
				label.Color = a.colors.Text
				return label.Layout(gtx)
			}),
		)
	})
}

func (a *App) inputBox(gtx layout.Context, editor *widget.Editor, hint string) layout.Dimensions {
	return a.box(gtx, a.colors.SurfaceMuted, layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(12), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		style := material.Editor(a.theme, editor, hint)
		style.Color = a.colors.Text
		style.HintColor = a.colors.TextMuted
		style.TextSize = unit.Sp(14)
		return style.Layout(gtx)
	})
}

func (a *App) primaryButton(gtx layout.Context, clickable *widget.Clickable, textValue string) layout.Dimensions {
	style := material.Button(a.theme, clickable, textValue)
	style.Background = a.colors.Primary
	style.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	if clickable.Hovered() {
		style.Background = a.colors.Accent
	}
	if clickable.Pressed() {
		style.Background = a.colors.Primary
	}
	if a.busy {
		gtx = gtx.Disabled()
		if a.activeTask == "PDF 분할 중" {
			style.Background = a.colors.Accent
			style.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
		} else {
			style.Background = a.colors.SurfaceAlt
			style.Color = a.colors.TextMuted
		}
	}
	style.CornerRadius = 0
	style.Inset = layout.Inset{Top: 14, Bottom: 14, Left: 16, Right: 16}
	return style.Layout(gtx)
}

func (a *App) secondaryButton(gtx layout.Context, clickable *widget.Clickable, textValue string) layout.Dimensions {
	style := material.Button(a.theme, clickable, textValue)
	style.Background = a.colors.SurfaceAlt
	style.Color = a.colors.Text
	if clickable.Hovered() {
		style.Background = a.colors.Hover
	}
	if clickable.Pressed() {
		style.Background = a.colors.Pressed
	}
	if a.busy {
		gtx = gtx.Disabled()
		style.Background = a.colors.SurfaceMuted
		style.Color = a.colors.TextMuted
	}
	style.CornerRadius = 0
	style.Inset = layout.Inset{Top: 12, Bottom: 12, Left: 14, Right: 14}
	return style.Layout(gtx)
}

func (a *App) themeModeButton(gtx layout.Context, clickable *widget.Clickable, textValue string, active bool) layout.Dimensions {
	bg := color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x24}
	textColor := color.NRGBA{R: 0xF2, G: 0xF6, B: 0xFB, A: 0xFF}
	if active {
		bg = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
		textColor = a.colors.Primary
	}
	if clickable.Hovered() && !active {
		bg = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x38}
	}
	if clickable.Pressed() {
		bg = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x4A}
	}

	return a.freeClickBox(gtx, clickable, bg, layout.Inset{
		Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(12), Right: unit.Dp(12),
	}, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(a.theme, unit.Sp(12), textValue)
		label.Color = textColor
		return label.Layout(gtx)
	})
}

func (a *App) windowControlButton(gtx layout.Context, clickable *widget.Clickable, textValue string, closeAction bool) layout.Dimensions {
	bg := color.NRGBA{}
	if clickable.Hovered() {
		bg = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x18}
	}
	if closeAction && clickable.Hovered() {
		bg = color.NRGBA{R: 0xD9, G: 0x2D, B: 0x20, A: 0xFF}
	}
	if clickable.Pressed() {
		bg = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x28}
	}
	if closeAction && clickable.Pressed() {
		bg = color.NRGBA{R: 0xB4, G: 0x23, B: 0x18, A: 0xFF}
	}

	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(38)), gtx.Dp(unit.Dp(30)))
		gtx.Constraints.Max = gtx.Constraints.Min
		return a.box(gtx, bg, layout.UniformInset(0), func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Label(a.theme, unit.Sp(13), textValue)
				label.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
				return label.Layout(gtx)
			})
		})
	})
}

func (a *App) modeButton(gtx layout.Context, clickable *widget.Clickable, textValue string, active bool) layout.Dimensions {
	bg := a.colors.SurfaceAlt
	textColor := a.colors.Text
	if active {
		bg = a.colors.PrimarySoft
		textColor = a.colors.Primary
	}
	if clickable.Hovered() {
		if active {
			bg = a.colors.Hover
		} else {
			bg = a.colors.SurfaceMuted
		}
	}
	if clickable.Pressed() {
		bg = a.colors.Pressed
	}

	return a.clickBox(gtx, clickable, bg, layout.Inset{
		Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(14), Right: unit.Dp(14),
	}, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(a.theme, unit.Sp(13), textValue)
		label.Color = textColor
		return label.Layout(gtx)
	})
}

func (a *App) boundaryButtonLabel(page int) string {
	if a.boundaries[page] {
		return "분할선 켜짐"
	}
	return "이 페이지 뒤 분할"
}

func (a *App) box(gtx layout.Context, bg color.NRGBA, inset layout.Inset, content layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := inset.Layout(gtx, content)
	call := macro.Stop()

	rect := image.Rectangle{Max: dims.Size}
	paint.FillShape(gtx.Ops, bg, clip.Rect(rect).Op())
	call.Add(gtx.Ops)
	return dims
}

func (a *App) clickBox(gtx layout.Context, clickable *widget.Clickable, bg color.NRGBA, inset layout.Inset, content layout.Widget) layout.Dimensions {
	if a.busy {
		gtx = gtx.Disabled()
	}
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return a.box(gtx, bg, inset, content)
	})
}

func (a *App) freeClickBox(gtx layout.Context, clickable *widget.Clickable, bg color.NRGBA, inset layout.Inset, content layout.Widget) layout.Dimensions {
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return a.box(gtx, bg, inset, content)
	})
}

func (a *App) statusColor() color.NRGBA {
	switch a.statusTone {
	case statusSuccess:
		return a.colors.Success
	case statusError:
		return a.colors.Error
	default:
		return a.colors.Primary
	}
}

func valueOrDash(v int) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", v)
}
