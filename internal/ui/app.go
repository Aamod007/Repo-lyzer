package ui

import (
	"fmt"
	"strings"

	"github.com/agnivo988/Repo-lyzer/internal/analyzer"
	"github.com/agnivo988/Repo-lyzer/internal/github"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateMenu sessionState = iota
	stateInput
	stateLoading
	stateDashboard
)

type MainModel struct {
	state         sessionState
	menu          EnhancedMenuModel
	input         string // Repository input
	spinner       spinner.Model
	dashboard     DashboardModel
	settings      SettingsModel
	help          HelpModel
	history       HistoryModel
	progress      *ProgressTracker
	err           error
	windowWidth   int
	windowHeight  int
	analysisType  string // quick, detailed, custom
	appSettings   AppSettings
}

func NewMainModel() MainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return MainModel{
		state:     stateMenu,
		menu:      NewMenuModel(),
		spinner:   s,
		dashboard: NewDashboardModel(),
	}
}

func (m MainModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		// Propagate to children
		m.menu.Update(msg)
		m.dashboard.Update(msg)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateMenu:
		newMenu, newCmd := m.menu.Update(msg)
		m.menu = newMenu.(MenuModel)
		cmds = append(cmds, newCmd)

		if m.menu.SelectedOption == 0 && m.menu.Done { // Analyze
			m.state = stateInput
			m.menu.Done = false // Reset for back navigation
		} else if m.menu.SelectedOption == 2 && m.menu.Done { // Exit
			return m, tea.Quit
		}

	case stateInput:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				if m.input != "" {
					m.state = stateLoading
					cmds = append(cmds, m.analyzeRepo(m.input))
				}
			case tea.KeyBackspace:
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case tea.KeyRunes:
				m.input += string(msg.Runes)
			case tea.KeyEsc:
				m.state = stateMenu
			}
		}

	case stateLoading:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		if result, ok := msg.(AnalysisResult); ok {
			m.dashboard.SetData(result)
			m.state = stateDashboard
			m.progress = nil
		}
		if err, ok := msg.(error); ok {
			m.err = err
			m.state = stateInput // Go back to input on error
			m.progress = nil
		}

	case stateDashboard:
		newDash, newCmd := m.dashboard.Update(msg)
		m.dashboard = newDash.(DashboardModel)
		cmds = append(cmds, newCmd)

		if m.dashboard.BackToMenu {
			m.state = stateMenu
			m.dashboard.BackToMenu = false
			m.input = ""
		}
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	switch m.state {
	case stateMenu:
		return m.menu.View()
	case stateInput:
		return m.inputView()
	case stateLoading:
		loadMsg := fmt.Sprintf("📊 Analyzing %s", m.input)
		if m.analysisType != "" {
			loadMsg += fmt.Sprintf(" (%s mode)", strings.ToUpper(m.analysisType))
		}

		statusView := fmt.Sprintf("%s %s...", m.spinner.View(), loadMsg)

		// Show progress stages if available
		if m.progress != nil {
			stages := m.progress.GetAllStages()
			statusView += "\n\n"
			for _, stage := range stages {
				prefix := "⏳ "
				if stage.IsComplete {
					prefix = "✅ "
				} else if stage.IsActive {
					prefix = "⚙️  "
				}
				statusView += prefix + stage.Name + "\n"
			}

			// Add elapsed time
			elapsed := m.progress.GetElapsedTime()
			statusView += fmt.Sprintf("\n⏱️  %ds elapsed", int(elapsed.Seconds()))
		}

		statusView += "\n\n" + SubtleStyle.Render("Press ESC to cancel")

		return lipgloss.Place(
			m.windowWidth, m.windowHeight,
			lipgloss.Center, lipgloss.Center,
			statusView,
		)
	case stateDashboard:
		return m.dashboard.View()
	}
	return ""
}

func (m MainModel) inputView() string {
	inputContent :=
		TitleStyle.Render("📥 ENTER REPOSITORY") + "\n\n" +
			InputStyle.Render("> "+m.input) + "\n\n" +
			SubtleStyle.Render("Format: owner/repo  •  Press Enter to run")

	if m.err != nil {
		inputContent += "\n\n" + ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	box := BoxStyle.Render(inputContent)

	if m.windowWidth == 0 {
		return box
	}

	return lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func (m MainModel) analyzeRepo(repoName string) tea.Cmd {
	return func() tea.Msg {
		parts := strings.Split(repoName, "/")
		if len(parts) != 2 {
			return fmt.Errorf("repository must be in owner/repo format")
		}

		tracker := NewProgressTracker()

		// Stage 1: Fetch repository
		client := github.NewClient()
		repo, err := client.GetRepo(parts[0], parts[1])
		if err != nil {
			return err
		}
		tracker.NextStage()

		// Stage 2: Analyze commits
		commits, _ := client.GetCommits(parts[0], parts[1], 365)
		tracker.NextStage()

		// Stage 3: Analyze contributors
		contributors, _ := client.GetContributors(parts[0], parts[1])
		tracker.NextStage()

		// Stage 4: Analyze languages
		languages, _ := client.GetLanguages(parts[0], parts[1])
		tracker.NextStage()

		// Stage 5: Compute metrics
		score := analyzer.CalculateHealth(repo, commits)
		busFactor, busRisk := analyzer.BusFactor(contributors)
		maturityScore, maturityLevel := analyzer.RepoMaturityScore(repo, len(commits), len(contributors), false)
		tracker.NextStage()

		// Mark complete
		tracker.NextStage()

		return AnalysisResult{
			Repo:          repo,
			Commits:       commits,
			Contributors:  contributors,
			Languages:     languages,
			HealthScore:   score,
			BusFactor:     busFactor,
			BusRisk:       busRisk,
			MaturityScore: maturityScore,
			MaturityLevel: maturityLevel,
		}
	}
}

func Run() error {
	p := tea.NewProgram(NewMainModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
