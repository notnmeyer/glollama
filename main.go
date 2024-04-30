package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ollama/ollama/api"

	"github.com/notnmeyer/glollama/internal/history"
	"github.com/notnmeyer/glollama/internal/ollama"
)

var viewWidth = 0

const defaultModel = "codellama"

type responseMsg string

type response struct {
	// the server may split the response across several... responses.
	// accumulate them here. we use this to trigger updates.
	acc chan string
	// there's no way to read the current contents of a viewport, so
	// we append acc's messages here to update the viewport with.
	all string
}

type app struct {
	client  *ollama.Chat
	history *history.History
	ta      textarea.Model
	vp      viewport.Model
	resp    *response
}

func (a app) helpView() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\n  ↑/↓: Navigate • x: Reset •  q: Quit\n")
}

func newApp() (*app, error) {
	model := os.Getenv("MODEL")
	if model == "" {
		model = defaultModel
	}
	client := ollama.New(model)

	// the textarea where the query is entered
	ta := textarea.New()
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Prompt = fmt.Sprintf("%s > ", client.Model)
	ta.Focus()

	// the textarea where the response is displayed
	vp := viewport.New(20, 40)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return &app{
		client:  client,
		history: history.New(),
		ta:      ta,
		vp:      vp,
		resp: &response{
			acc: make(chan string),
			all: "",
		},
	}, nil
}

func (a app) Init() tea.Cmd {
	return tea.Batch(activityMonitor(a.resp.acc))
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var taCmd, vpCmd tea.Cmd

	switch msg := msg.(type) {
	case responseMsg:
		// TODO: we want the responses in the history, but this breaks subsequent queries
		// a.history.Append(&api.Message{
		// 	Role:    "assistant",
		// 	Content: strings.TrimSpace(string(msg)),
		// })

		a.resp.all += string(msg)
		update, err := render(a.resp.all)
		if err != nil {
			log.Fatal(err)
		}
		a.vp.SetContent(update)
		return a, activityMonitor(a.resp.acc)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return a, tea.Quit
		case "enter":
			value := strings.TrimSpace(a.ta.Value())
			a.ta.Reset()

			a.history.Append(&api.Message{
				Role:    "user",
				Content: value,
			})

			go func() {
				a.resp.acc <- fmt.Sprintf("# %s\n", value)
				a.client.Chat(a.history, a.respFunc)
				a.resp.all += "\n"
			}()

			a.vp.SetContent(a.resp.all)

			a.vp, vpCmd = a.vp.Update(msg)
			a.ta, taCmd = a.ta.Update(msg)
			return a, tea.Batch(vpCmd, taCmd)
		case "up":
			a.vp.LineUp(1)
			a.vp, vpCmd = a.vp.Update(msg)
			return a, tea.Batch(vpCmd)
		case "down":
			a.vp.LineDown(1)
			a.vp, vpCmd = a.vp.Update(msg)
			return a, tea.Batch(vpCmd)
		default:
			a.ta, taCmd = a.ta.Update(msg)
			return a, tea.Batch(taCmd)
		}
	case tea.WindowSizeMsg:
		viewWidth = msg.Width - 2 // 2 seems to be the magic value
		// TODO: calculate height based on length of input
		// a.query.SetHeight(msg.Height - 10)
		a.ta.SetWidth(viewWidth)
		a.vp.Width = viewWidth
		return a, nil
	default:
		return a, nil
	}
}

func (a app) View() string {
	return fmt.Sprintf(
		"%s\n%s",
		a.ta.View(),
		a.vp.View()+a.helpView(),
	) + "\n\n"
}

func main() {
	model, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}

// called each time a new response is streamed from the server
func (a app) respFunc(resp api.ChatResponse) error {
	a.resp.acc <- resp.Message.Content
	return nil
}

// generates a message for the update loop
func activityMonitor(ch chan string) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-ch)
	}
}

// render markdown with glamour
func render(str string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(viewWidth-10),
	)
	if err != nil {
		return "", err
	}

	rendered, err := renderer.Render(str)
	if err != nil {
		return "", err
	}

	return rendered, nil
}
