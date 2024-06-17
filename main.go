package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	// Define UI components
	collectionsPanel := tview.NewList().AddItem("List Item 1", "", '0', nil).AddItem("List Item 1", "ddd", 'q', nil)
	collectionsPanel.SetBorder(true).SetTitle("Collections (c)")
	requestsPanel := tview.NewBox().SetBorder(true).SetTitle("Requests")
	requestEditor := tview.NewBox().SetBorder(true).SetTitle("Request Editor")
	responseViewer := tview.NewBox().SetBorder(true).SetTitle("Response")

	// Layout the panels
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(collectionsPanel, 0, 1, false).
		AddItem(requestsPanel, 0, 1, false)

	mainPanel := tview.NewFlex().
		AddItem(leftPanel, 0, 1, false).
		AddItem(requestEditor, 0, 2, false).
		AddItem(responseViewer, 0, 2, false)

	// Capture keyboard events
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'c':
				app.SetFocus(collectionsPanel)
				return nil
			case 'q':
				app.Stop()
				return nil
			case 'r':
				app.SetFocus(requestsPanel)
				return nil
			case 'a':
				if app.GetFocus() == collectionsPanel {
					app.SetFocus(requestEditor)
				} else {
					app.SetFocus(responseViewer)
				}

				return nil
			}

		}
		return event
	})

	// Set the root panel
	if err := app.SetRoot(mainPanel, true).SetFocus(mainPanel).Run(); err != nil {
		panic(err)
	}

}
