package ui

import (
	"github.com/hyprcommunity/hypr-release/guiapi/bridge"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// NewSystemTab GUI'deki "System" sekmesini oluşturur.
func NewSystemTab(win fyne.Window, b *bridge.Bridge) fyne.CanvasObject {
	table := widget.NewTable(
		func() (int, int) { return 0, 2 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {},
	)

	refreshBtn := widget.NewButton("Refresh System Info", func() {
		jsonStr, err := b.SystemInfo()
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		label := widget.NewMultiLineEntry()
		label.SetText(jsonStr)

		dialog.ShowCustom("System Info", "Close", label, win)
	})

	infoLabel := widget.NewLabel("Hyprland System Information")
	content := container.NewBorder(
		container.NewVBox(infoLabel, refreshBtn),
		nil, nil, nil, table,
	)

	return content
}
