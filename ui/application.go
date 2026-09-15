package ui

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	homewizard "codeberg.org/clambin/go-homewizard/client/v1"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	pollRequestTimeout = 5 * time.Second
	pollSlowInterval   = 30 * time.Second
)

func Activate(app *adw.Application) {
	window := adw.NewApplicationWindow(&app.Application)
	window.SetTitle("Flow")
	window.SetDefaultSize(340, 250)

	cssProvider := gtk.NewCSSProvider()
	cssProvider.LoadFromString(`
		.power-positive { color: var(--warning-color); }
		.power-negative { color: var(--success-color); }
	`)
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	headerBar := adw.NewHeaderBar()

	menu := gio.NewMenu()
	menu.Append("Preferences", "app.preferences")

	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetMenuModel(menu)

	headerBar.PackEnd(menuBtn)

	powerLabel := gtk.NewLabel("")
	powerLabel.AddCSSClass("title-1")
	powerLabel.SetMarginTop(12)
	powerLabel.SetMarginBottom(12)

	powerBox := gtk.NewBox(gtk.OrientationVertical, 0)
	powerBox.AddCSSClass("card")
	powerBox.SetMarginTop(12)
	powerBox.SetMarginBottom(12)
	powerBox.SetMarginStart(12)
	powerBox.SetMarginEnd(12)
	powerBox.SetHExpand(true)

	powerGraph := NewGraph(powerBox)

	powerBox.Append(powerLabel)
	powerBox.Append(powerGraph)
	carousel := adw.NewCarousel()
	carousel.Append(powerBox)

	dots := adw.NewCarouselIndicatorDots()
	dots.SetCarousel(carousel)
	dots.SetHExpand(true)
	dots.SetHAlign(gtk.AlignCenter)
	dots.SetMarginTop(6)
	dots.SetMarginBottom(6)

	outerBox := gtk.NewBox(gtk.OrientationVertical, 0)
	outerBox.Append(carousel)
	outerBox.Append(dots)

	httpClient := &http.Client{Timeout: pollRequestTimeout}
	var pollCancel context.CancelFunc

	startPolling := func() {
		if pollCancel != nil {
			pollCancel()
		}

		cfg := loadConfig()

		pollCtx, cancel := context.WithCancel(context.Background())
		pollCancel = cancel

		renderMeasurement := func(data *homewizard.RecentMeasurement, host string, err error) {
			if err != nil {
				if host == "" {
					powerLabel.SetText("Please configure the P1 Host URL in Preferences.")
				} else {
					powerLabel.SetText(fmt.Sprintf("Error: %v", err))
				}
				return
			}
			powerLabel.SetText(fmt.Sprintf("%.0f W", data.ActivePowerW))
			if data.ActivePowerW > 0 {
				powerLabel.RemoveCSSClass("power-negative")
				powerGraph.RemoveCSSClass("power-negative")
				powerLabel.AddCSSClass("power-positive")
				powerGraph.AddCSSClass("power-positive")
			} else {
				powerLabel.RemoveCSSClass("power-positive")
				powerGraph.RemoveCSSClass("power-positive")
				powerLabel.AddCSSClass("power-negative")
				powerGraph.AddCSSClass("power-negative")
			}
			powerGraph.AddValue(data.ActivePowerW)
		}

		doUpdate := func() error {
			reqCtx, reqCancel := context.WithTimeout(pollCtx, pollRequestTimeout)
			defer reqCancel()
			client := homewizard.Client{HTTPClient: httpClient, Target: cfg.Host}
			data, err := client.GetRecentMeasurement(reqCtx)
			if err == nil {
				d := data
				glib.IdleAdd(func() { renderMeasurement(&d, cfg.Host, nil) })
				return nil
			}
			if pollCtx.Err() != nil {
				return err
			}
			log.Printf("Error fetching P1 data: %v", err)
			host := cfg.Host
			e := err
			glib.IdleAdd(func() { renderMeasurement(nil, host, e) })
			return err
		}

		go func() {
			slowMode := doUpdate() != nil

			nextInterval := func() time.Duration {
				if slowMode {
					return pollSlowInterval
				}
				return time.Duration(cfg.Interval) * time.Second
			}

			timer := time.NewTimer(nextInterval())
			defer timer.Stop()

			for {
				select {
				case <-pollCtx.Done():
					return
				case <-timer.C:
					slowMode = doUpdate() != nil
					timer.Reset(nextInterval())
				}
			}
		}()
	}

	prefAction := gio.NewSimpleAction("preferences", nil)
	prefAction.ConnectActivate(func(p *glib.Variant) {
		showPreferencesDialog(window, startPolling)
	})
	app.AddAction(prefAction)
	app.SetAccelsForAction("app.preferences", []string{"<Primary>comma"})

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		if pollCancel != nil {
			pollCancel()
		}
		app.Quit()
	})
	app.AddAction(quitAction)
	app.SetAccelsForAction("app.quit", []string{"<Primary>q"})

	startPolling()

	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(headerBar)
	toolbarView.SetContent(outerBox)

	window.SetContent(toolbarView)
	window.Show()
}
