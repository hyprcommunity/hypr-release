package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "github.com/hyprcommunity/hypr-release/guiapi/bridge"
    "github.com/hyprcommunity/hypr-release/guiapi/ui"
)

func main() {
    a := app.New()
    w := a.NewWindow("Hypr Release Manager")
    w.Resize(fyne.NewSize(800, 600))

    bridge := backend.NewBridge()

    tabs := container.NewAppTabs(
        container.NewTabItem("System", ui.NewSystemTab(w, bridge)),
        container.NewTabItem("Dotfiles", ui.NewDotfilesTab(w, bridge)),
        container.NewTabItem("Updates", ui.NewUpdatesTab(w, bridge)),
    )

    w.SetContent(tabs)
    w.ShowAndRun()
}
