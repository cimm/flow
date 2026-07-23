package main

import (
	"os"

	"flow/ui"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

func main() {
	app := adw.NewApplication("io.github.cimm.flow.app", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		ui.Activate(app)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
