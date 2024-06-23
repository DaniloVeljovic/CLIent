package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Collection struct {
	Name     string
	Requests []Request
}

func (c *Collection) RemoveRequest(requestName string) {
	for i, req := range c.Requests {
		if req.Name == requestName {
			c.Requests = append(c.Requests[:i], c.Requests[i+1:]...)
			return
		}
	}
}

type Request struct {
	Name    string
	Verb    string
	Url     string
	Headers map[string]string
	Body    interface{}
}

type EnvironmentVariables struct {
	Key   string
	Value string
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
	environmentVariables := getEnvironmentVariables()
	environmentPanel := tview.NewTextArea()
	environmentPanel.SetBorder(true).SetTitle("Environment (v)")
	requestEditor := tview.NewTextArea()
	requestEditor.SetBorder(true).SetTitle("Editor (e)")
	responseViewer := tview.NewTextView()
	responseViewer.SetBorder(true).SetTitle("Output (o)")

	var activeRequest *Request
	var activeCollection *Collection

	builder := strings.Builder{}
	for i := range environmentVariables {
		builder.WriteString(environmentVariables[i].Key + "=" + environmentVariables[i].Value + "\n")
	}
	environmentPanel.SetText(builder.String(), true)

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
					builder.WriteString(request.Name)
					builder.WriteString("\n")
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
		AddItem(environmentPanel, 0, 1, false).
		AddItem(requestsPanel, 0, 1, false)

	mainPanel := tview.NewFlex().
		AddItem(leftPanel, 0, 1, false).
		AddItem(requestEditor, 0, 2, false).
		AddItem(responseViewer, 0, 2, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			if app.GetFocus() == environmentPanel {
				variables := environmentPanel.GetText()
				lines := strings.Split(variables, "\n")

				var envVars []EnvironmentVariables
				for _, line := range lines {
					if line == "" {
						continue
					}
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						envVars = append(envVars, EnvironmentVariables{
							Key:   parts[0],
							Value: parts[1],
						})
					}
				}
				environmentVariables = envVars
				saveEnvVariables(environmentVariables)

			} else if activeCollection == nil && activeRequest == nil {
				collection := Collection{
					Name:     requestEditor.GetText(),
					Requests: []Request{},
				}
				collectionsPanel.Clear()
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
								builder.WriteString(request.Name)
								builder.WriteString("\n")
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
			} else if activeCollection != nil && activeRequest == nil {
				params := strings.Split(requestEditor.GetText(), "\n")
				if len(params) == 1 {
					activeCollection.Name = requestEditor.GetText()
				} else {
					request := Request{}
					request.Name = params[0]
					request.Verb = params[1]
					request.Url = params[2]
					json.Unmarshal([]byte(params[3]), &request.Headers)
					json.Unmarshal([]byte(params[4]), &request.Body)
					activeRequest = &request
					activeCollection.Requests = append(activeCollection.Requests, request)
					saveCollections(collections)
				}
				collectionsPanel.Clear()
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
								builder.WriteString(request.Name)
								builder.WriteString("\n")
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
			} else if activeCollection != nil && activeRequest != nil {
				text := requestEditor.GetText()
				lines := strings.SplitN(text, "\n", 5)
				if len(lines) == 5 {
					activeRequest.Name = lines[0]
					activeRequest.Verb = lines[1]
					activeRequest.Url = lines[2]
					json.Unmarshal([]byte(lines[3]), &activeRequest.Headers)
					json.Unmarshal([]byte(lines[4]), &activeRequest.Body)
					saveCollections(collections)
				}
			}
			requestsPanel.Clear()
			for j := range activeCollection.Requests {
				request := &activeCollection.Requests[j]
				requestsPanel.AddItem(request.Name, "", 0, func() {
					activeRequest = request
					builder := strings.Builder{}
					builder.WriteString(request.Name)
					builder.WriteString("\n")
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
		} else if app.GetFocus() == environmentPanel {

		} else if event.Key() == tcell.KeyRune {
			if app.GetFocus() != requestEditor {
				if app.GetFocus() == requestsPanel {
					switch event.Rune() {
					case 'a':
						requestEditor.SetText("", true)
						app.SetFocus(requestEditor)
						return nil
					case 'u':
						builder := strings.Builder{}
						builder.WriteString(activeRequest.Name)
						builder.WriteString("\n")
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
						app.SetFocus(requestEditor)
						return nil
					case 'd':
						activeCollection.RemoveRequest(activeRequest.Name)
						requestEditor.SetText("", true)
						requestsPanel.Clear()
						for j := range activeCollection.Requests {
							request := &activeCollection.Requests[j]

							requestsPanel.AddItem(request.Name, "", 0, func() {
								activeRequest = request
								builder := strings.Builder{}
								builder.WriteString(request.Name)
								builder.WriteString("\n")
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
						return nil
					}
				} else if app.GetFocus() == collectionsPanel {
					switch event.Rune() {
					case 'a':
						requestEditor.SetText("", true)
						activeRequest = nil
						activeCollection = nil
						app.SetFocus(requestEditor)
						return nil
					case 'u':
						builder := strings.Builder{}
						builder.WriteString(activeCollection.Name)
						requestEditor.SetText(builder.String(), true)
						app.SetFocus(requestEditor)
						return nil
					case 'd':
						for i, collection := range collections {
							if collection.Name == activeCollection.Name {
								collections = append(collections[:i], collections[i+1:]...)
								break
							}
						}
						app.SetFocus(collectionsPanel)
						requestsPanel.Clear()
						requestEditor.SetText("", true)
						collectionsPanel.Clear()
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
										builder.WriteString(request.Name)
										builder.WriteString("\n")
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
				case 'o':
					app.SetFocus(responseViewer)
					return nil
				case 'v':
					app.SetFocus(environmentPanel)
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

func getEnvironmentVariables() []EnvironmentVariables {
	file, err := os.ReadFile("./db/environment.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}

	var environmentVariables []EnvironmentVariables

	err = json.Unmarshal(file, &environmentVariables)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}

	return environmentVariables
}

func CallApi(text string) string {
	client := &http.Client{}
	params := strings.SplitN(text, "\n", 5)
	if len(params) < 4 {
		return "Invalid input format. Expected: <NAME> <METHOD> <URL> <HEADERS_JSON> <BODY>"
	}

	_, method, url, headersJson, bodyJson := params[0], params[1], params[2], params[3], params[4]

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

func saveEnvVariables(environmentVariables []EnvironmentVariables) {
	file, err := os.Create("./db/environment.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	data, err := json.MarshalIndent(environmentVariables, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
