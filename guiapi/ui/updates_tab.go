package ui

import (
    "fmt"
    "github.com/hyprcommunity/hypr-release/guiapi/bridge"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

// NewUpdatesTab GUI'deki "Updates" sekmesini oluşturur.
func NewUpdatesTab(win fyne.Window, bridge *backend.Bridge) fyne.CanvasObject {
    versionLabel := widget.NewLabel("Current Version: unknown")
    progress := widget.NewProgressBar()
    progress.Hide()
    logArea := widget.NewMultiLineEntry()
    logArea.SetPlaceHolder("Update log...")

    checkBtn := widget.NewButton("Check Release", func() {
        progress.Show()
        progress.SetValue(0.3)
        version, err := bridge.CheckRelease()
        if err != nil {
            dialog.ShowError(err, win)
            progress.Hide()
            return
        }
        versionLabel.SetText(fmt.Sprintf("Current Version: %s", version))
        progress.SetValue(1)
        progress.Hide()
    })

    exportBtn := widget.NewButton("Export Release JSON", func() {
        json, err := bridge.ExportReleaseJSON()
        if err != nil {
            dialog.ShowError(err, win)
            return
        }
        logArea.SetText(json)
        dialog.ShowInformation("Exported", "Release data exported successfully.", win)
    })

    controls := container.NewHBox(checkBtn, exportBtn)
    return container.NewBorder(container.NewVBox(versionLabel, controls), nil, nil, nil, container.NewVSplit(progress, logArea))
}
