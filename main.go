package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// P1Data holds the fields we care about from the HomeWizard P1 /api/v1/data response.
type P1Data struct {
	ActivePowerW   float64 `json:"active_power_w"`
	ActivePowerL1W float64 `json:"active_power_l1_w"`
	ActivePowerL2W float64 `json:"active_power_l2_w"`
	ActivePowerL3W float64 `json:"active_power_l3_w"`
}

// fetchP1Data makes a single GET request to the HomeWizard P1 meter and returns the parsed data.
func fetchP1Data(host string) (*P1Data, error) {
	if host == "" {
		return nil, fmt.Errorf("no host configured")
	}
	url := fmt.Sprintf("http://%s/api/v1/data", host)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("requesting P1 data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from P1 meter", resp.StatusCode)
	}

	var data P1Data
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding P1 response: %w", err)
	}
	return &data, nil
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

func loadHost() string {
	path, err := getConfigPath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cfg struct {
		Host string `json:"host"`
	}
	json.Unmarshal(b, &cfg)
	return cfg.Host
}

func saveHost(host string) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	cfg := struct {
		Host string `json:"host"`
	}{Host: host}
	b, _ := json.Marshal(cfg)
	return os.WriteFile(path, b, 0644)
}

func showPreferencesDialog(window *adw.ApplicationWindow, onUpdate func()) {
	dialog := adw.NewPreferencesDialog()
	dialog.SetSearchEnabled(false)

	page := adw.NewPreferencesPage()
	page.SetTitle("General")
	page.SetIconName("preferences-system-symbolic")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Connection")
	group.SetDescription("Configure the HomeWizard P1 meter")

	ipRow := adw.NewEntryRow()
	ipRow.SetTitle("IP Address")
	ipRow.SetText(loadHost())

	group.Add(ipRow)
	page.Add(group)
	dialog.Add(page)

	dialog.ConnectClosed(func() {
		saveHost(ipRow.Text())
		onUpdate()
	})

	dialog.Present(window)
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

		refreshData := func() {
			host := loadHost()
			data, err := fetchP1Data(host)
			if err != nil {
				log.Printf("Error fetching P1 data: %v", err)
				if host == "" {
					powerLabel.SetText("Please configure the P1 IP Address in Preferences.")
				} else {
					powerLabel.SetText(fmt.Sprintf("Error: %v", err))
				}
			} else {
				powerLabel.SetText(fmt.Sprintf("⚡ %.0f W", data.ActivePowerW))
			}
		}

		prefAction := gio.NewSimpleAction("preferences", nil)
		prefAction.ConnectActivate(func(p *glib.Variant) {
			showPreferencesDialog(window, refreshData)
		})
		app.AddAction(prefAction)

		refreshData()

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
