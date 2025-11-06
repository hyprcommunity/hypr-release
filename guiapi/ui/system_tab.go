package ui

import (
    "fmt"
    "github.com/hyprcommunity/hypr-release/guiapi/bridge"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

// NewSystemTab GUI'deki "System" sekmesini oluşturur.
func NewSystemTab(win fyne.Window, bridge *backend.Bridge) fyne.CanvasObject {
    table := widget.NewTable(
        func() (int, int) { return 0, 2 },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(i, j int, o fyne.CanvasObject) {},
    )

    refreshBtn := widget.NewButton("Refresh System Info", func() {
        data, err := bridge.SystemInfo()
        if err != nil {
            dialog.ShowError(err, win)
            return
        }

        rows := make([][]string, 0, len(data))
        for k, v := range data {
            rows = append(rows, []string{k, v})
        }

        table.Length = func() (int, int) { return len(rows), 2 }
        table.UpdateCell = func(i, j int, o fyne.CanvasObject) {
            o.(*widget.Label).SetText(rows[i][j])
        }
        table.Refresh()
    })

    infoLabel := widget.NewLabel("Hyprland System Information")
    content := container.NewBorder(
        container.NewVBox(infoLabel, refreshBtn),
        nil, nil, nil, table,
    )

    return content
}
