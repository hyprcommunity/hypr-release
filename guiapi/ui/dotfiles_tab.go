package ui

import (
    "fmt"
    "github.com/hyprcommunity/hypr-release/guiapi/bridge"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

// NewDotfilesTab GUI'deki "Dotfiles" sekmesini oluşturur.
func NewDotfilesTab(win fyne.Window, bridge *backend.Bridge) fyne.CanvasObject {
    list := widget.NewList(
        func() int { return 0 },
        func() fyne.CanvasObject {
            return container.NewVBox(
                widget.NewLabel("Name"),
                widget.NewLabel("Author"),
                widget.NewLabel("Branch"),
                widget.NewButton("Install", nil),
                widget.NewButton("Update", nil),
            )
        },
        func(i widget.ListItemID, o fyne.CanvasObject) {},
    )

    refresh := widget.NewButton("Load Dotfiles", func() {
        entries, err := bridge.GetDotfiles()
        if err != nil {
            dialog.ShowError(err, win)
            return
        }

        list.Length = func() int { return len(entries) }
        list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
            box := o.(*fyne.Container)
            name := box.Objects[0].(*widget.Label)
            author := box.Objects[1].(*widget.Label)
            branch := box.Objects[2].(*widget.Label)
            install := box.Objects[3].(*widget.Button)
            update := box.Objects[4].(*widget.Button)

            entry := entries[i]
            name.SetText(fmt.Sprintf("Name: %s", entry.Name))
            author.SetText(fmt.Sprintf("Author: %s", entry.Author))
            branch.SetText(fmt.Sprintf("Branch: %s", entry.Branch))

            install.OnTapped = func() {
                go func() {
                    err := bridge.InstallDotfile(entry.Name)
                    if err != nil {
                        dialog.ShowError(err, win)
                    } else {
                        dialog.ShowInformation("Success", "Installed "+entry.Name, win)
                    }
                }()
            }
            update.OnTapped = func() {
                go func() {
                    err := bridge.UpdateDotfile(entry.Name)
                    if err != nil {
                        dialog.ShowError(err, win)
                    } else {
                        dialog.ShowInformation("Updated", "Dotfile updated successfully", win)
                    }
                }()
            }
        }

        list.Refresh()
    })

    return container.NewBorder(container.NewVBox(refresh), nil, nil, nil, list)
}
