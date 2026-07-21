package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

func showPreferencesWindow(app *gtk.Application, mainWindow *gtk.ApplicationWindow, onUpdate func()) {
	win := gtk.NewWindow()
	win.SetTitle("Preferences")
	win.SetDefaultSize(300, 100)
	win.SetTransientFor(&mainWindow.Window)
	win.SetModal(true)

	vbox := gtk.NewBox(gtk.OrientationVertical, 10)
	vbox.SetMarginTop(10)
	vbox.SetMarginBottom(10)
	vbox.SetMarginStart(10)
	vbox.SetMarginEnd(10)

	label := gtk.NewLabel("P1 IP Address:")
	label.SetHAlign(gtk.AlignStart)

	entry := gtk.NewEntry()
	entry.SetText(loadHost())

	win.ConnectCloseRequest(func() bool {
		saveHost(entry.Text())
		onUpdate()
		return false
	})

	vbox.Append(label)
	vbox.Append(entry)

	win.SetChild(vbox)
	win.Show()
}

func main() {
	app := gtk.NewApplication("io.github.cimm.flow.app", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		window := gtk.NewApplicationWindow(app)
		window.SetTitle("Flow")
		window.SetDefaultSize(300, 150)

		headerBar := gtk.NewHeaderBar()
		window.SetTitlebar(headerBar)

		menu := gio.NewMenu()
		menu.Append("Preferences", "app.preferences")

		menuBtn := gtk.NewMenuButton()
		menuBtn.SetIconName("open-menu-symbolic")
		menuBtn.SetMenuModel(menu)

		headerBar.PackEnd(menuBtn)

		powerLabel := gtk.NewLabel("")
		powerLabel.AddCSSClass("power-label")
		powerLabel.SetWrap(true)

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
			showPreferencesWindow(app, window, refreshData)
		})
		app.AddAction(prefAction)

		refreshData()

		// Apply some basic CSS so it doesn't look completely bare.
		css := gtk.NewCSSProvider()
		css.LoadFromString(`
			.power-label {
				font-size: 32px;
				font-weight: bold;
				margin: 24px;
				text-align: center;
			}
		`)
		display := window.Widget.Display()
		gtk.StyleContextAddProviderForDisplay(display, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

		window.SetChild(powerLabel)
		window.Show()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
