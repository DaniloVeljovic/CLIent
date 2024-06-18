package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Collection struct {
	Name string
}

type Request struct {
	Name string
}

func getCollections() []Collection {
	return []Collection{
		{Name: "Collection 1"},
		{Name: "Collection 2"},
	}
}

func getRequests(collection *Collection) Request {
	m := make(map[string]Request)
	m["Collection 1"] = Request{Name: "Request 1"}
	m["Collection 2"] = Request{Name: "Request 2"}
	return m[collection.Name]
}

func main() {
	app := tview.NewApplication()

	collections := getCollections()

	// Define UI components
	collectionsPanel := tview.NewList()

	collectionsPanel.SetBorder(true).SetTitle("Collections (c)")
	requestsPanel := tview.NewList()
	requestsPanel.SetBorder(true).SetTitle("Requests")
	requestEditor := tview.NewBox().SetBorder(true).SetTitle("Request Editor")
	responseViewer := tview.NewBox().SetBorder(true).SetTitle("Response")

	for _, collection := range collections {
		collectionsPanel.AddItem(collection.Name, "", 0, func() {
			r := getRequests(&collection)

			for i := range requestsPanel.GetItemCount() {
				requestsPanel.RemoveItem(i)
			}

			requestsPanel.AddItem(r.Name, "", 0, nil)
		})
	}

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
