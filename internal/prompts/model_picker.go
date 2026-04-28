package prompts

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const maxVisible = 12

// --- Provider Selector ---

type providerSelector struct {
	context   string // agent or group name
	providers []string
	cursor    int
	offset    int // scroll offset
	allModels map[string][]string
	done      bool
}

func newProviderSelector(context string, allModels map[string][]string) providerSelector {
	providers := make([]string, 0, len(allModels))
	for p := range allModels {
		providers = append(providers, p)
	}
	sort.Strings(providers)
	return providerSelector{
		context:   context,
		providers: providers,
		allModels: allModels,
	}
}

func (m providerSelector) Init() tea.Cmd { return nil }

func (m providerSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.providers)-1 {
				m.cursor++
				if m.cursor >= m.offset+maxVisible {
					m.offset = m.cursor - maxVisible + 1
				}
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m providerSelector) View() string {
	s := titleStyle.Render("Select provider") + "\n"
	s += groupStyle.Render("for: "+m.context) + "\n"
	s += dimStyle.Render("──────────────────────────────") + "\n\n"

	visible := m.providers
	start := m.offset
	end := m.offset + maxVisible
	if end > len(visible) {
		end = len(visible)
	}

	if start > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↑ %d more\n", start))
	}

	for i := start; i < end; i++ {
		p := m.providers[i]
		count := len(m.allModels[p])
		cursor := "  "
		style := dimStyle
		if i == m.cursor {
			cursor = "▶ "
			style = activeStyle
		}
		s += style.Render(fmt.Sprintf("%s%-20s %s", cursor, p, dimStyle.Render(fmt.Sprintf("(%d models)", count)))) + "\n"
	}

	if end < len(visible) {
		s += dimStyle.Render(fmt.Sprintf("  ↓ %d more\n", len(visible)-end))
	}

	s += "\n"
	s += dimStyle.Render("↑↓ navigate   enter select   esc cancel") + "\n"
	return s
}

// --- Model Selector ---

type modelSelector struct {
	context  string // agent or group name
	provider string
	models   []string
	cursor   int
	offset   int
	done     bool
}

func newModelSelector(context, provider string, models []string) modelSelector {
	return modelSelector{
		context:  context,
		provider: provider,
		models:   models,
	}
}

func (m modelSelector) Init() tea.Cmd { return nil }

func (m modelSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.models)-1 {
				m.cursor++
				if m.cursor >= m.offset+maxVisible {
					m.offset = m.cursor - maxVisible + 1
				}
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m modelSelector) View() string {
	// Strip provider prefix from model name for cleaner display
	shortName := func(id string) string {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			return parts[1]
		}
		return id
	}

	s := titleStyle.Render("Select model") + "\n"
	s += groupStyle.Render("for: "+m.context) + "  " + dimStyle.Render("provider: "+m.provider) + "\n"
	s += dimStyle.Render("──────────────────────────────") + "\n\n"

	start := m.offset
	end := m.offset + maxVisible
	if end > len(m.models) {
		end = len(m.models)
	}

	if start > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↑ %d more\n", start))
	}

	for i := start; i < end; i++ {
		model := m.models[i]
		cursor := "  "
		style := dimStyle
		if i == m.cursor {
			cursor = "▶ "
			style = activeStyle
		}
		s += style.Render(fmt.Sprintf("%s%s", cursor, shortName(model))) + "\n"
	}

	if end < len(m.models) {
		s += dimStyle.Render(fmt.Sprintf("  ↓ %d more\n", len(m.models)-end))
	}

	s += "\n"
	s += dimStyle.Render("↑↓ navigate   enter select   esc back to providers") + "\n"
	return s
}

// RunModelPicker shows provider → model picker.
// context is the agent or group name for display in title.
// Returns the full model ID (provider/model) or "" if cancelled.
func RunModelPicker(context string, allModels map[string][]string) (string, error) {
	if len(allModels) == 0 {
		return "", fmt.Errorf("no models available")
	}

	// 1. Provider selection
	ps := newProviderSelector(context, allModels)
	prog := tea.NewProgram(ps, tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		return "", err
	}
	result := final.(providerSelector)
	if !result.done {
		return "", nil
	}
	selectedProvider := result.providers[result.cursor]

	// 2. Model selection
	ms := newModelSelector(context, selectedProvider, allModels[selectedProvider])
	prog = tea.NewProgram(ms, tea.WithAltScreen())
	final, err = prog.Run()
	if err != nil {
		return "", err
	}
	mResult := final.(modelSelector)
	if !mResult.done {
		return "", nil // cancelled — go back
	}
	return mResult.models[mResult.cursor], nil
}
