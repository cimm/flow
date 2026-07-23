package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
)

func showPreferencesDialog(window *adw.ApplicationWindow, onUpdate func()) {
	dialog := adw.NewPreferencesDialog()
	dialog.SetSearchEnabled(false)

	page := adw.NewPreferencesPage()

	group := adw.NewPreferencesGroup()
	group.SetTitle("Connection")
	group.SetDescription("Configure the HomeWizard P1 meter")

	cfg := loadConfig()

	ipRow := adw.NewEntryRow()
	ipRow.SetTitle("Host URL (e.g. http://192.168.1.x)")
	ipRow.SetText(cfg.Host)

	intervalRow := adw.NewSpinRowWithRange(1, 3600, 1)
	intervalRow.SetTitle("Update Interval in Seconds")
	intervalRow.SetValue(float64(cfg.Interval))

	group.Add(ipRow)
	group.Add(intervalRow)
	page.Add(group)
	dialog.Add(page)

	dialog.ConnectClosed(func() {
		newCfg := AppConfig{
			Host:     ipRow.Text(),
			Interval: int(intervalRow.Value()),
		}
		saveConfig(newCfg)
		onUpdate()
	})

	dialog.Present(window)
}
