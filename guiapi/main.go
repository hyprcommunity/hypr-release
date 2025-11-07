package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/hyprcommunity/hypr-release/guiapi/bridge"
	"github.com/hyprcommunity/hypr-release/guiapi/ui"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hypr Release Manager")
	w.Resize(fyne.NewSize(800, 600))

	// Bridge örneğini oluştur
	b := bridge.NewBridge()

	tabs := container.NewAppTabs(
		container.NewTabItem("System", ui.NewSystemTab(w, b)),
		container.NewTabItem("Dotfiles", ui.NewDotfilesTab(w, b)),
		container.NewTabItem("Updates", ui.NewUpdatesTab(w, b)),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}
