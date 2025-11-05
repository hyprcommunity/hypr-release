package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var HyprlandIcons = map[string]string{
	"system":   "󰍹", // monitor
	"dotfiles": "󰣇", // layers
	"updates":  "󰚰", // sync
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Hypr-Release Dashboard")
	w.Resize(fyne.NewSize(920, 580))

	// system tab
	systemLog := widget.NewMultiLineEntry()
	systemLog.SetPlaceHolder("system check logs...")
	systemTab := container.NewVBox(
		widget.NewLabelWithStyle(HyprlandIcons["system"]+"  System Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewButton("Check System", func() {
			// buraya CheckHyprSystem() entegre edilecek
			systemLog.SetText("Running system check...\n✅ hyprland: v0.42.1\n⬆️  hyprpaper: update available (v0.6.0 → v0.6.2)")
		}),
		systemLog,
	)

	// dotfiles tab
	dotLog := widget.NewMultiLineEntry()
	dotLog.SetPlaceHolder("dotfile metadata logs...")
	dotTab := container.NewVBox(
		widget.NewLabelWithStyle(HyprlandIcons["dotfiles"]+"  Dotfiles", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewButton("Check Dotfile", func() {
			// buraya CheckAll() + CheckTestingStatus() entegre edilecek
			dotLog.SetText("Checking HyDE...\nMain: v25.9.1\nBranch: testing\nChannel: stable")
		}),
		dotLog,
	)

	// updates tab
	updateLog := widget.NewMultiLineEntry()
	updateLog.SetPlaceHolder("update progress...")
	updateTab := container.NewVBox(
		widget.NewLabelWithStyle(HyprlandIcons["updates"]+"  Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewButton("Run Update", func() {
			// buraya updateing.WriteMetaFile() entegre edilecek
			updateLog.SetText("Pulling latest tags...\nBuilding new release...\n✅ Metadata written to /etc/hyprland-release")
		}),
		updateLog,
	)

	// tab container
	tabs := container.NewAppTabs(
		container.NewTabItem("System", systemTab),
		container.NewTabItem("Dotfiles", dotTab),
		container.NewTabItem("Updates", updateTab),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(tabs)
	w.ShowAndRun()
}
