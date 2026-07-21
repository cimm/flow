package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// P1Data holds the fields we care about from the HomeWizard P1 /api/v1/data response.
type P1Data struct {
	ActivePowerW     float64 `json:"active_power_w"`
	ActivePowerL1W   float64 `json:"active_power_l1_w"`
	ActivePowerL2W   float64 `json:"active_power_l2_w"`
	ActivePowerL3W   float64 `json:"active_power_l3_w"`
}

// fetchP1Data makes a single GET request to the HomeWizard P1 meter and returns the parsed data.
func fetchP1Data(host string) (*P1Data, error) {
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

func main() {
	host := os.Getenv("HOMEWIZARD_HOST")
	if host == "" {
		fmt.Println("Error: HOMEWIZARD_HOST environment variable is not set.")
		fmt.Println("Usage: HOMEWIZARD_HOST=<ip-or-hostname> flow")
		os.Exit(1)
	}

	app := gtk.NewApplication("io.github.cimm.flow.app", gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		window := gtk.NewApplicationWindow(app)
		window.SetTitle("Flow")
		window.SetDefaultSize(300, 150)

		// Fetch data (one-shot).
		data, err := fetchP1Data(host)

		var label *gtk.Label
		if err != nil {
			log.Printf("Error fetching P1 data: %v", err)
			label = gtk.NewLabel(fmt.Sprintf("Error: %v", err))
		} else {
			label = gtk.NewLabel(fmt.Sprintf("⚡ %.0f W", data.ActivePowerW))
		}

		label.AddCSSClass("power-label")

		// Apply some basic CSS so it doesn't look completely bare.
		css := gtk.NewCSSProvider()
		css.LoadFromString(`
			.power-label {
				font-size: 48px;
				font-weight: bold;
				margin: 24px;
			}
		`)
		display := window.Widget.Display()
		gtk.StyleContextAddProviderForDisplay(display, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

		window.SetChild(label)
		window.Show()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
