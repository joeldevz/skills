package prompts

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // orange for (not set)
	checkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))  // green for set models
	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Bold(true)
)

// shortModel strips provider prefix: "anthropic/claude-haiku-4-5" → "claude-haiku-4-5"
func shortModel(id string) string {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return id
}

// helpBar renders a consistent help bar at the bottom
func helpBar(keys ...string) string {
	parts := make([]string, 0, len(keys))
	for i := 0; i < len(keys)-1; i += 2 {
		k := keyStyle.Render(keys[i])
		v := dimStyle.Render(keys[i+1])
		parts = append(parts, k+" "+v)
	}
	return "\n" + strings.Join(parts, dimStyle.Render("   "))
}

// --- Profile Name Form ---

type profileNameForm struct {
	textInput textinput.Model
	err       string
	cancelled bool
}

func newProfileNameForm(initial string) profileNameForm {
	ti := textinput.New()
	ti.Placeholder = "e.g. backend, front-v2"
	ti.SetValue(initial)
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 30
	return profileNameForm{textInput: ti}
}

func (m profileNameForm) Init() tea.Cmd { return textinput.Blink }

func (m profileNameForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.textInput.Value())
			if err := validateProfileName(name); err != nil {
				m.err = err.Error()
				return m, nil
			}
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	// Live validation
	name := strings.TrimSpace(m.textInput.Value())
	if name != "" {
		if err := validateProfileName(name); err != nil {
			m.err = err.Error()
		} else {
			m.err = ""
		}
	} else {
		m.err = ""
	}
	return m, cmd
}

func (m profileNameForm) View() string {
	s := "\n"
	s += titleStyle.Render("  New Profile") + "\n"
	s += dimStyle.Render("  ──────────────────────────────") + "\n\n"
	s += fmt.Sprintf("  Profile name: %s\n\n", m.textInput.View())
	if m.err != "" {
		s += "  " + errorStyle.Render("✗ "+m.err) + "\n"
	} else {
		s += "  " + dimStyle.Render("lowercase letters and hyphens only") + "\n"
	}
	s += helpBar("enter", "confirm", "esc", "cancel")
	return s
}

func validateProfileName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 32 {
		return fmt.Errorf("max 32 characters")
	}
	if name == "default" {
		return fmt.Errorf("\"default\" is reserved")
	}
	if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(name) {
		return fmt.Errorf("use lowercase letters, numbers and hyphens (e.g. backend, front-v2)")
	}
	return nil
}

func RunProfileNameForm(initial string) (string, error) {
	m := newProfileNameForm(initial)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	result := final.(profileNameForm)
	if result.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	name := strings.TrimSpace(result.textInput.Value())
	if name == "" {
		return "", fmt.Errorf("cancelled")
	}
	return name, nil
}

// --- Mode Selector ---

type ConfigMode int

const (
	ModeSimple ConfigMode = iota
	ModeAdvanced
)

type modeSelector struct {
	cursor    int
	cancelled bool
}

func (m modeSelector) Init() tea.Cmd { return nil }

func (m modeSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m modeSelector) View() string {
	s := "\n"
	s += titleStyle.Render("  Configure Models") + "\n"
	s += dimStyle.Render("  ──────────────────────────────") + "\n\n"

	options := []struct {
		label string
		desc  string
		hint  string
	}{
		{
			"Simple",
			"Set one model per role",
			"Orchestrator · Workers · Advisor",
		},
		{
			"Advanced",
			"Set model per agent individually",
			"10 agents: orchestrator, coder, tech-planner...",
		},
	}

	for i, opt := range options {
		indicator := "  ○"
		labelStyle := dimStyle
		if i == m.cursor {
			indicator = "  ●"
			labelStyle = activeStyle
		}
		s += labelStyle.Render(fmt.Sprintf("%s  %s", indicator, opt.label)) + "\n"
		s += dimStyle.Render(fmt.Sprintf("     %s", opt.desc)) + "\n"
		s += dimStyle.Render(fmt.Sprintf("     %s", opt.hint)) + "\n\n"
	}

	s += helpBar("↑↓", "navigate", "enter", "select", "esc", "cancel")
	return s
}

func RunModeSelector() (ConfigMode, error) {
	m := modeSelector{}
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return ModeSimple, err
	}
	result := final.(modeSelector)
	if result.cancelled {
		return ModeSimple, fmt.Errorf("cancelled")
	}
	if result.cursor == 1 {
		return ModeAdvanced, nil
	}
	return ModeSimple, nil
}

// --- Simple Model Picker ---

type simpleGroup struct {
	key    string   // display name
	desc   string   // what agents it affects
	agents []string // agent names in this group
}

var simpleGroups = []simpleGroup{
	{
		key:    "Orchestrator",
		desc:   "Plans, coordinates and delegates all work",
		agents: []string{"orchestrator", "manager"},
	},
	{
		key:    "Workers",
		desc:   "Execute tasks: plan, code, verify, review",
		agents: []string{"tech-planner", "product-planner", "coder", "verifier", "test-reviewer", "security", "skill-validator"},
	},
	{
		key:    "Advisor",
		desc:   "Senior strategic consultant (use best model)",
		agents: []string{"advisor"},
	},
}

type simpleModelPicker struct {
	groups    []simpleGroup
	models    map[string]string // groupKey -> modelID
	cursor    int
	allModels map[string][]string
	cancelled bool
}

func newSimpleModelPicker(initial map[string]string) simpleModelPicker {
	allModels, _ := LoadOpencodeModels()
	if initial == nil {
		initial = make(map[string]string)
	}
	return simpleModelPicker{
		groups:    simpleGroups,
		models:    initial,
		allModels: allModels,
	}
}

func (m simpleModelPicker) Init() tea.Cmd { return nil }

func (m simpleModelPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.groups)-1 {
				m.cursor++
			}
		case "enter":
			// Open model picker for selected group
			g := m.groups[m.cursor]
			modelID, err := RunModelPicker(g.key, m.allModels)
			if err == nil && modelID != "" {
				m.models[g.key] = modelID
			}
			return m, nil
		case "s":
			// Set all groups to same model
			modelID, err := RunModelPicker("all groups", m.allModels)
			if err == nil && modelID != "" {
				for _, g := range m.groups {
					m.models[g.key] = modelID
				}
			}
			return m, nil
		case "c":
			// Confirm/save
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m simpleModelPicker) View() string {
	// Count configured groups
	configured := 0
	for _, g := range m.groups {
		if m.models[g.key] != "" {
			configured++
		}
	}
	total := len(m.groups)

	s := "\n"
	s += titleStyle.Render("  Model Setup — Simple") + "\n"
	if configured == total {
		s += checkStyle.Render(fmt.Sprintf("  ✓ All %d roles configured — press c to save", total)) + "\n"
	} else {
		s += warnStyle.Render(fmt.Sprintf("  %d/%d roles configured", configured, total)) + "\n"
	}
	s += dimStyle.Render("  ──────────────────────────────") + "\n\n"

	for i, g := range m.groups {
		isSelected := i == m.cursor
		modelID := m.models[g.key]

		// Group label
		cursor := "  "
		labelStyle := dimStyle
		if isSelected {
			cursor = "▶ "
			labelStyle = activeStyle
		}
		s += labelStyle.Render(fmt.Sprintf("%s%s", cursor, g.key)) + "\n"

		// Description
		s += dimStyle.Render(fmt.Sprintf("     %s", g.desc)) + "\n"
		s += dimStyle.Render(fmt.Sprintf("     Agents: %s", strings.Join(g.agents, ", "))) + "\n"

		// Current model
		if modelID == "" {
			s += "     " + warnStyle.Render("⚠  not set — press enter to pick a model") + "\n"
		} else {
			s += "     " + checkStyle.Render("✓  "+shortModel(modelID)) + "  " + dimStyle.Render("("+modelID+")") + "\n"
		}
		s += "\n"
	}

	s += helpBar(
		"enter", "pick model for selected role",
		"s", "set all to same model",
		"c", "save profile",
		"esc", "cancel",
	)
	return s
}

// --- Advanced Model Picker ---

type agentConfig struct {
	name  string
	model string // full ID or ""
}

var agentDescriptions = map[string]string{
	"orchestrator":    "Coordinates all agents, decides strategy",
	"tech-planner":    "Writes PLAN.md with technical steps",
	"product-planner": "Writes SPEC.md with business context",
	"coder":           "Implements code changes",
	"manager":         "Executes plan step by step",
	"verifier":        "Runs lint, build, tests",
	"test-reviewer":   "Reviews test quality",
	"security":        "Adversarial security judge",
	"skill-validator": "Validates code conventions",
	"advisor":         "Senior strategic consultant",
}

var agentList = []string{
	"orchestrator", "tech-planner", "product-planner",
	"coder", "manager", "verifier",
	"test-reviewer", "security", "skill-validator", "advisor",
}

type advancedModelPicker struct {
	agents    []agentConfig
	cursor    int
	allModels map[string][]string
	cancelled bool
}

func newAdvancedModelPicker(initial map[string]string) advancedModelPicker {
	allModels, _ := LoadOpencodeModels()
	agents := make([]agentConfig, len(agentList))
	for i, name := range agentList {
		agents[i] = agentConfig{name: name, model: initial[name]}
	}
	return advancedModelPicker{agents: agents, allModels: allModels}
}

func (m advancedModelPicker) Init() tea.Cmd { return nil }

func (m advancedModelPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.agents)-1 {
				m.cursor++
			}
		case "enter":
			agent := m.agents[m.cursor]
			modelID, err := RunModelPicker(agent.name, m.allModels)
			if err == nil && modelID != "" {
				m.agents[m.cursor].model = modelID
			}
			return m, nil
		case "s":
			modelID, err := RunModelPicker("all agents", m.allModels)
			if err == nil && modelID != "" {
				for i := range m.agents {
					m.agents[i].model = modelID
				}
			}
			return m, nil
		case "c":
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m advancedModelPicker) View() string {
	configured := 0
	for _, a := range m.agents {
		if a.model != "" {
			configured++
		}
	}
	total := len(m.agents)

	s := "\n"
	s += titleStyle.Render("  Model Setup — Advanced") + "\n"
	if configured == total {
		s += checkStyle.Render(fmt.Sprintf("  ✓ All %d agents configured — press c to save", total)) + "\n"
	} else {
		s += warnStyle.Render(fmt.Sprintf("  %d/%d agents configured", configured, total)) + "\n"
	}
	s += dimStyle.Render("  ──────────────────────────────") + "\n\n"

	for i, ac := range m.agents {
		isSelected := i == m.cursor
		cursor := "  "
		nameStyle := dimStyle
		if isSelected {
			cursor = "▶ "
			nameStyle = activeStyle
		}

		// Agent name + description on one line
		desc := agentDescriptions[ac.name]
		s += nameStyle.Render(fmt.Sprintf("%s%-18s", cursor, ac.name))
		s += dimStyle.Render(desc) + "\n"

		// Model on next line, indented
		if ac.model == "" {
			s += "     " + warnStyle.Render("⚠  not set") + "\n"
		} else {
			s += "     " + checkStyle.Render("✓  "+shortModel(ac.model)) + "\n"
		}
	}

	s += helpBar(
		"enter", "pick model",
		"s", "set all",
		"c", "save",
		"esc", "cancel",
	)
	return s
}

// --- Summary Screen ---

type summaryScreen struct {
	name      string
	models    map[string]string // agentName or groupKey -> modelID
	mode      ConfigMode
	confirmed bool
	cancelled bool
}

func (m summaryScreen) Init() tea.Cmd { return nil }

func (m summaryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "y", "c":
			m.confirmed = true
			return m, tea.Quit
		case "esc", "n", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m summaryScreen) View() string {
	s := "\n"
	s += titleStyle.Render("  Profile Summary") + "\n"
	s += dimStyle.Render("  ──────────────────────────────") + "\n\n"
	s += fmt.Sprintf("  Name:  %s\n", activeStyle.Render(m.name))
	s += fmt.Sprintf("  Mode:  %s\n\n", dimStyle.Render(map[ConfigMode]string{ModeSimple: "Simple", ModeAdvanced: "Advanced"}[m.mode]))

	if m.mode == ModeSimple {
		for _, g := range simpleGroups {
			modelID := m.models[g.key]
			if modelID == "" {
				s += fmt.Sprintf("  %-14s %s\n", g.key, warnStyle.Render("not set"))
			} else {
				s += fmt.Sprintf("  %-14s %s\n", g.key, checkStyle.Render(shortModel(modelID)))
				s += fmt.Sprintf("  %-14s %s\n", "", dimStyle.Render("→ "+strings.Join(g.agents, ", ")))
			}
		}
	} else {
		for _, name := range agentList {
			modelID := m.models[name]
			if modelID == "" {
				s += fmt.Sprintf("  %-18s %s\n", name, warnStyle.Render("not set"))
			} else {
				s += fmt.Sprintf("  %-18s %s\n", name, checkStyle.Render(shortModel(modelID)))
			}
		}
	}

	s += "\n"
	s += helpBar("enter", "save profile", "esc", "go back and edit")
	return s
}

// --- Runners ---

func RunSimpleModelPicker(initial map[string]string) (map[string]string, error) {
	m := newSimpleModelPicker(initial)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	result := final.(simpleModelPicker)
	if result.cancelled {
		return nil, fmt.Errorf("cancelled")
	}
	return result.models, nil
}

func RunAdvancedModelPicker(initial map[string]string) (map[string]string, error) {
	m := newAdvancedModelPicker(initial)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	result := final.(advancedModelPicker)
	if result.cancelled {
		return nil, fmt.Errorf("cancelled")
	}
	models := make(map[string]string)
	for _, ac := range result.agents {
		if ac.model != "" {
			models[ac.name] = ac.model
		}
	}
	return models, nil
}

// ProfileResult is the output of the creation flow.
type ProfileResult struct {
	Name   string
	Models map[string]string
}

// RunProfileCreationFlow runs the complete flow:
// name → mode → models → summary/confirm → ProfileResult
func RunProfileCreationFlow(initialModels map[string]string) (*ProfileResult, error) {
	// 1. Name
	name, err := RunProfileNameForm("")
	if err != nil {
		return nil, err
	}

	// 2. Mode
	mode, err := RunModeSelector()
	if err != nil {
		return nil, err
	}

	// 3. Models
	var models map[string]string
	if mode == ModeSimple {
		models, err = RunSimpleModelPicker(initialModels)
	} else {
		models, err = RunAdvancedModelPicker(initialModels)
	}
	if err != nil {
		return nil, err
	}

	// 4. Summary + confirm
	summary := summaryScreen{name: name, models: models, mode: mode}
	prog := tea.NewProgram(summary, tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		return nil, err
	}
	summaryResult := final.(summaryScreen)
	if summaryResult.cancelled {
		return nil, fmt.Errorf("cancelled")
	}

	return &ProfileResult{Name: name, Models: models}, nil
}

// LoadOpencodeModels executes "opencode models" and returns map[provider][]fullModelID.
func LoadOpencodeModels() (map[string][]string, error) {
	out, err := exec.Command("opencode", "models").Output()
	if err != nil {
		return defaultModels(), nil
	}
	result := make(map[string][]string)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "/", 2)
		if len(parts) == 2 {
			result[parts[0]] = append(result[parts[0]], line)
		}
	}
	if len(result) == 0 {
		return defaultModels(), nil
	}
	return result, nil
}

func defaultModels() map[string][]string {
	return map[string][]string{
		"anthropic": {
			"anthropic/claude-opus-4-6",
			"anthropic/claude-sonnet-4-6",
			"anthropic/claude-haiku-4-5",
		},
		"openai": {"openai/gpt-4o", "openai/o3"},
		"google": {
			"google/gemini-2.5-pro",
		},
	}
}
