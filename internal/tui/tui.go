package tui

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"sort"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"

	"github.com/go-goal/tagger/internal/model"
	"github.com/go-goal/tagger/internal/prediction"
	"github.com/go-goal/tagger/internal/tfidf"
)

var (
	// Color scheme
	bgColor        = lipgloss.Color("#1E1E2E")
	accentColor    = lipgloss.Color("#89B4FA")
	textColor      = lipgloss.Color("#CDD6F4")
	highlightColor = lipgloss.Color("#F5E0DC")
	successColor   = lipgloss.Color("#A6E3A1")
	errorColor     = lipgloss.Color("#F38BA8")

	titleStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Background(accentColor).
		Padding(0, 1).
		Bold(true)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1).
		MarginTop(1)

	inputStyle = lipgloss.NewStyle().
		Foreground(textColor)

	resultStyle = lipgloss.NewStyle().
		Foreground(successColor)

	errorStyle = lipgloss.NewStyle().
		Foreground(errorColor)

	listItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2)

	// New styles for the list
	listHeaderStyle = lipgloss.NewStyle().
		Foreground(highlightColor).
		Bold(true).
		PaddingBottom(1)

	listStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 1, 1, 2)
)

type TaggerApp struct {
	predictor *prediction.Predictor
	textInput textinput.Model
	result    map[string]string
	err       error
	width     int
	height    int
	lastInput string // Add this line
}

func NewTaggerApp(tfidfPath, labelsDir, modelsDir string) (*TaggerApp, error) {
	ti := textinput.New()
	ti.Placeholder = "Enter rate name"
	ti.Focus()

	predictor, err := prediction.NewPredictor(tfidfPath, labelsDir, modelsDir)
	if err != nil {
		return nil, err
	}

	return &TaggerApp{
		predictor: predictor,
		textInput: ti,
	}, nil
}

func (app *TaggerApp) Init() tea.Cmd {
	return textinput.Blink
}

func (app *TaggerApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return app, tea.Quit
		case tea.KeyEnter:
			return app, app.predict
		}

	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height

	case predictionMsg:
		app.result = msg.results
		app.err = nil
		app.lastInput = app.textInput.Value() // Add this line

	case errMsg:
		app.err = msg.err
	}

	app.textInput, cmd = app.textInput.Update(msg)

	// Add this block
	if app.textInput.Value() != app.lastInput {
		return app, app.predict
	}

	return app, cmd
}

func (app *TaggerApp) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Rate Name Classifier") + "\n\n")

	// Input box
	inputBox := boxStyle.Render(fmt.Sprintf(
		"Enter a rate name to classify:\n\n%s",
		inputStyle.Render(app.textInput.View()),
	))
	b.WriteString(inputBox + "\n\n")

	// Result table
	if len(app.result) > 0 {
		resultTable := boxStyle.Render(
			listHeaderStyle.Render("Predicted classes:") + "\n" +
			app.formatResultsAsTable(),
		)
		b.WriteString(resultTable + "\n\n")
	}

	// Error box
	if app.err != nil {
		errorBox := boxStyle.Copy().BorderForeground(errorColor).Render(
			fmt.Sprintf("Error: %v", app.err),
			)
		b.WriteString(errorBox + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(textColor).Faint(true).Render("Press Ctrl+C to quit.\n"))

	// Center the content
	return lipgloss.NewStyle().
		Background(bgColor).
		Render(lipgloss.Place(app.width, app.height,
			lipgloss.Center, lipgloss.Center,
			b.String()))
}

func (app *TaggerApp) formatResultsAsTable() string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Model", "Prediction"})
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	models := make([]string, 0, len(app.result))
	for model := range app.result {
		models = append(models, model)
	}
	sort.Strings(models)

	for _, model := range models {
		modelName := strings.TrimPrefix(model, "catboost_model_")
		table.Append([]string{modelName, app.result[model]})
	}

	table.Render()
	return buf.String()
}

func (app *TaggerApp) predict() tea.Msg {
	results, err := app.predictor.Predict(app.textInput.Value())
	if err != nil {
		return errMsg{err}
	}
	return predictionMsg{results}
}

type tfIdfDataMsg struct {
	data tfidf.TfIdfData
}

type classLabelsMsg struct {
	labels []string
}

type modelMsg struct {
	model *model.CatBoostModel
}

type predictionMsg struct {
	results map[string]string
}

type errMsg struct {
	err error
}

type tickMsg time.Time

func (app *TaggerApp) Run() error {
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
