package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	homewizard "codeberg.org/clambin/go-homewizard/client/v1"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AppConfig struct {
	Host     string `json:"host"`
	Interval int    `json:"interval"`
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "flow")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

func loadConfig() AppConfig {
	path, err := getConfigPath()
	var cfg AppConfig
	if err == nil {
		b, err := os.ReadFile(path)
		if err == nil {
			json.Unmarshal(b, &cfg)
		}
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 1 // default to 1 second
	}
	return cfg
}

func saveConfig(cfg AppConfig) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	b, _ := json.Marshal(cfg)
	return os.WriteFile(path, b, 0644)
}

func main() {
	app := adw.NewApplication("io.github.cimm.flow.app", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
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
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
