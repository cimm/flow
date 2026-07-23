package ui

import (
	"context"
	"fmt"
	"log"

	homewizard "codeberg.org/clambin/go-homewizard/client/v1"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func Activate(app *adw.Application) {
	window := adw.NewApplicationWindow(&app.Application)
	window.SetTitle("Flow")
	window.SetDefaultSize(300, 150)

	headerBar := adw.NewHeaderBar()

	menu := gio.NewMenu()
	menu.Append("Preferences", "app.preferences")

	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetMenuModel(menu)

	headerBar.PackEnd(menuBtn)

	powerLabel := gtk.NewLabel("")
	powerLabel.AddCSSClass("title-1")
	powerLabel.SetVExpand(true)

	var timeoutId glib.SourceHandle

	startPolling := func() {
		if timeoutId != 0 {
			glib.SourceRemove(timeoutId)
			timeoutId = 0
		}

		cfg := loadConfig()

		doUpdate := func() bool {
			client := homewizard.Client{Target: cfg.Host}
			data, err := client.GetRecentMeasurement(context.Background())
			if err != nil {
				log.Printf("Error fetching P1 data: %v", err)
				if cfg.Host == "" {
					powerLabel.SetText("Please configure the P1 Host URL in Preferences.")
				} else {
					powerLabel.SetText(fmt.Sprintf("Error: %v", err))
				}
			} else {
				powerLabel.SetText(fmt.Sprintf("⚡ %.0f W", data.ActivePowerW))
			}
			return true
		}

		// Call once immediately
		doUpdate()

		timeoutId = glib.TimeoutAdd(uint(cfg.Interval*1000), doUpdate)
	}

	prefAction := gio.NewSimpleAction("preferences", nil)
	prefAction.ConnectActivate(func(p *glib.Variant) {
		showPreferencesDialog(window, startPolling)
	})
	app.AddAction(prefAction)
	app.SetAccelsForAction("app.preferences", []string{"<Primary>comma"})

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		app.Quit()
	})
	app.AddAction(quitAction)
	app.SetAccelsForAction("app.quit", []string{"<Primary>q"})

	startPolling()

	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(headerBar)
	toolbarView.SetContent(powerLabel)

	window.SetContent(toolbarView)
	window.Show()
}
