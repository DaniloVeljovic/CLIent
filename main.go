package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
	"net/http"
	_ "net/http"
	"strings"
	_ "strings"
)

type Verb string

const (
	GET    Verb = "GET"
	POST   Verb = "POST"
	PUT    Verb = "PUT"
	PATCH  Verb = "PATCH"
	DELETE Verb = "DELETE"
)

type Collection struct {
	Name string
}

type Request struct {
	Name    string
	Headers map[string]string
	Url     string
	Verb    Verb
	Body    string
}

func getCollections() []Collection {
	return []Collection{
		{Name: "Collection 1"},
		{Name: "Collection 2"},
	}
}

func getRequests(collection *Collection) []Request {
	m := make(map[string][]Request)
	m["Collection 1"] = []Request{
		{Name: "Request 1"},
		{Name: "Request 2"},
	}
	m["Collection 2"] = []Request{
		{Name: "Request 3", Verb: GET, Url: "https://www.google.com"},
		{Name: "Request 4", Verb: GET, Url: "https://www.postman-echo.com/get"},
	}
	return m[collection.Name]
}

func main() {
	app := tview.NewApplication()

	collections := getCollections()

	// Define UI components
	collectionsPanel := tview.NewList()

	collectionsPanel.SetBorder(true).SetTitle("Collections (c)")
	requestsPanel := tview.NewList()
	requestsPanel.SetBorder(true).SetTitle("Requests (r)")
	requestEditor := tview.NewTextArea()
	requestEditor.SetBorder(true).SetTitle("Request Editor (e)")
	responseViewer := tview.NewTextView()
	responseViewer.SetBorder(true).SetTitle("Response")

	for _, collection := range collections {
		collectionsPanel.AddItem(collection.Name, "", 0, func() {
			r := getRequests(&collection)

			for i := range requestsPanel.GetItemCount() {
				requestsPanel.RemoveItem(i)
			}

			for _, r := range r {
				requestsPanel.AddItem(r.Name, "", 0, func() {
					builder := strings.Builder{}
					builder.WriteString(string(r.Verb))
					builder.WriteString(" ")
					builder.WriteString(r.Url)

					requestEditor.SetText(builder.String(), false)
				})
			}
			app.SetFocus(requestsPanel)
		})
	}

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(collectionsPanel, 0, 1, false).
		AddItem(requestsPanel, 0, 1, false)

	mainPanel := tview.NewFlex().
		AddItem(leftPanel, 0, 1, false).
		AddItem(requestEditor, 0, 2, false).
		AddItem(responseViewer, 0, 2, false)

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
			case 'e':
				app.SetFocus(requestEditor)
				return nil
			case 'p':
				get, err := http.Get("https://www.postman-echo.com/get")
				if err != nil {
					return nil
				}
				defer get.Body.Close()
				body, _ := io.ReadAll(get.Body)
				responseViewer.SetText(string(body))
				return nil
			}

		}
		return event
	})

	if err := app.SetRoot(mainPanel, true).SetFocus(mainPanel).Run(); err != nil {
		panic(err)
	}
}
