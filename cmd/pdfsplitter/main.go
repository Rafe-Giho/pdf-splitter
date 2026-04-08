package main

import (
	"log"

	"gioui.org/app"
	"gioui.org/unit"

	"pdfsplitter/internal/ui"
)

func main() {
	go func() {
		window := new(app.Window)
		window.Option(
			app.Title("PDF Splitter"),
			app.Decorated(false),
			app.Size(unit.Dp(1180), unit.Dp(820)),
			app.MinSize(unit.Dp(760), unit.Dp(560)),
		)

		if err := ui.New(window).Run(); err != nil {
			log.Fatal(err)
		}
	}()

	app.Main()
}
