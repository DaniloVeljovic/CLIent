package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
	"net/http"
	"os"
	"strings"
)

type Collection struct {
	Id       int
	Name     string
	Requests []Request
}

type Request struct {
	Id      int
	Name    string
	Verb    string
	Url     string
	Headers map[string]string
	Body    interface{}
}

func getCollections() []Collection {
	file, err := os.ReadFile("./db/collection.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}

	var collections []Collection

	err = json.Unmarshal(file, &collections)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}

	return collections
}

func saveCollections(collections []Collection) {
	file, err := os.Create("./db/collection.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	data, err := json.MarshalIndent(collections, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func main() {
	app := tview.NewApplication()
	collections := getCollections()
	collectionsPanel := tview.NewList()
	collectionsPanel.SetBorder(true).SetTitle("Collections (c)")
	requestsPanel := tview.NewList()
	requestsPanel.SetBorder(true).SetTitle("Requests (r)")
	requestEditor := tview.NewTextArea()
	requestEditor.SetBorder(true).SetTitle("Request Editor (e)")
	responseViewer := tview.NewTextView()
	responseViewer.SetBorder(true).SetTitle("Response (v)")

	var activeRequest *Request
	var activeCollection *Collection

	for i := range collections {
		collection := &collections[i]
		collectionsPanel.AddItem(collection.Name, "", 0, func() {
			requestsPanel.Clear()
			activeCollection = collection
			activeRequest = nil

			for j := range collection.Requests {
				request := &collection.Requests[j]

				requestsPanel.AddItem(request.Name, "", 0, func() {
					activeRequest = request
					builder := strings.Builder{}
					builder.WriteString(request.Verb)
					builder.WriteString("\n")
					builder.WriteString(request.Url)
					builder.WriteString("\n")
					headers, _ := json.Marshal(request.Headers)
					builder.WriteString(string(headers))
					builder.WriteString("\n")
					body, _ := json.Marshal(request.Body)
					builder.WriteString(string(body))
					requestEditor.SetText(builder.String(), true)
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
		if event.Key() == tcell.KeyEsc {
			if activeCollection == nil && activeRequest == nil {
				collection := Collection{
					Id:       0,
					Name:     requestEditor.GetText(),
					Requests: []Request{},
				}
				activeCollection = &collection
				collections = append(collections, collection)
				for i := range collections {
					collection := &collections[i]
					collectionsPanel.AddItem(collection.Name, "", 0, func() {
						requestsPanel.Clear()
						activeCollection = collection

						for j := range collection.Requests {
							request := &collection.Requests[j]

							requestsPanel.AddItem(request.Name, "", 0, func() {
								activeRequest = request
								builder := strings.Builder{}
								builder.WriteString(request.Verb)
								builder.WriteString("\n")
								builder.WriteString(request.Url)
								builder.WriteString("\n")
								headers, _ := json.Marshal(request.Headers)
								builder.WriteString(string(headers))
								builder.WriteString("\n")
								body, _ := json.Marshal(request.Body)
								builder.WriteString(string(body))
								requestEditor.SetText(builder.String(), true)
							})
						}
						app.SetFocus(requestsPanel)
					})
				}
				saveCollections(collections)
			} else if activeRequest != nil {
				text := requestEditor.GetText()
				lines := strings.SplitN(text, "\n", 4)
				if len(lines) == 4 {
					activeRequest.Verb = lines[0]
					activeRequest.Url = lines[1]
					json.Unmarshal([]byte(lines[2]), &activeRequest.Headers)
					json.Unmarshal([]byte(lines[3]), &activeRequest.Body)
					saveCollections(collections)
				}
			} else {
				text := requestEditor.GetText()
				lines := strings.SplitN(text, "\n", 4)
				if len(lines) == 4 {
					request := Request{}
					request.Verb = lines[0]
					request.Url = lines[1]
					json.Unmarshal([]byte(lines[2]), &request.Headers)
					json.Unmarshal([]byte(lines[3]), &request.Body)
					request.Name = "sample"
					activeRequest = &request
					activeCollection.Requests = append(activeCollection.Requests, request)
					saveCollections(collections)
				}
			}
			requestsPanel.Clear()
			for j := range activeCollection.Requests {
				request := &activeCollection.Requests[j]
				requestsPanel.AddItem(request.Name, "", 0, func() {
					activeRequest = request
					builder := strings.Builder{}
					builder.WriteString(request.Verb)
					builder.WriteString("\n")
					builder.WriteString(request.Url)
					builder.WriteString("\n")
					headers, _ := json.Marshal(request.Headers)
					builder.WriteString(string(headers))
					builder.WriteString("\n")
					body, _ := json.Marshal(request.Body)
					builder.WriteString(string(body))
					requestEditor.SetText(builder.String(), true)
				})
			}
			app.SetFocus(mainPanel)
		} else if event.Key() == tcell.KeyRune {
			if app.GetFocus() != requestEditor {
				if app.GetFocus() == requestsPanel {
					switch event.Rune() {
					case 'a':
						if activeRequest != nil {
							builder := strings.Builder{}
							builder.WriteString(activeRequest.Verb)
							builder.WriteString("\n")
							builder.WriteString(activeRequest.Url)
							builder.WriteString("\n")
							headers, _ := json.Marshal(activeRequest.Headers)
							builder.WriteString(string(headers))
							builder.WriteString("\n")
							body, _ := json.Marshal(activeRequest.Body)
							builder.WriteString(string(body))
							requestEditor.SetText(builder.String(), true)
						} else {
							requestEditor.SetText("", true)
						}
						app.SetFocus(requestEditor)
						return nil
					}
				} else if app.GetFocus() == collectionsPanel {
					switch event.Rune() {
					case 'a':
						if activeCollection != nil {
							requestEditor.SetText(activeCollection.Name, true)
						} else {
							requestEditor.SetText("", true)
							activeRequest = nil
							activeCollection = nil
						}
						app.SetFocus(requestEditor)
						return nil
					}
				}

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
					response := CallApi(requestEditor.GetText())
					responseViewer.SetText(response)
					return nil
				case 'v':
					app.SetFocus(responseViewer)
					return nil
				}
			}
		}

		return event
	})

	if err := app.SetRoot(mainPanel, true).SetFocus(mainPanel).Run(); err != nil {
		panic(err)
	}
}

func CallApi(text string) string {
	client := &http.Client{}
	params := strings.SplitN(text, "\n", 4)
	if len(params) < 4 {
		return "Invalid input format. Expected: <METHOD> <URL> <HEADERS_JSON> <BODY>"
	}

	method, url, headersJson, bodyJson := params[0], params[1], params[2], params[3]

	req, err := http.NewRequest(method, url, bytes.NewBufferString(bodyJson))
	if err != nil {
		return "Failed to create request: " + err.Error()
	}

	var headers map[string]string
	err = json.Unmarshal([]byte(headersJson), &headers)
	if err != nil {
		return "Failed to parse headers JSON: " + err.Error()
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "An error occurred: " + err.Error()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Failed to read response body: " + err.Error()
	}

	return string(body)
}
