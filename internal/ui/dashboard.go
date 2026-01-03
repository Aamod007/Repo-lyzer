package ui

import (
	"fmt"
	"time"

	"github.com/agnivo988/Repo-lyzer/internal/analyzer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardModel struct {
	data       AnalysisResult
	BackToMenu bool
	width      int
	height     int
	showExport bool
	statusMsg  string
}

func NewDashboardModel() DashboardModel {
	return DashboardModel{}
}

func (m *DashboardModel) SetData(data AnalysisResult) {
	m.data = data
	m.bridge = NewAnalyzerDataBridge(data)
}

type exportMsg struct {
	err error
	msg string
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case exportMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Export failed: %v", msg.err)
		} else {
			m.statusMsg = msg.msg
		}
		// Clear status after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg { return "clear_status" })

	case string:
		if msg == "clear_status" {
			m.statusMsg = ""
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.showExport {
				m.showExport = false
			} else {
				m.BackToMenu = true
			}
		case "e":
			m.showExport = !m.showExport
		case "j":
			if m.showExport {
				return m, func() tea.Msg {
					err := ExportJSON(m.data, "analysis.json")
					return exportMsg{err, "Exported to analysis.json"}
				}
			}
		case "m":
			if m.showExport {
				return m, func() tea.Msg {
					err := ExportMarkdown(m.data, "analysis.md")
					return exportMsg{err, "Exported to analysis.md"}
				}
			}
		}
	}
	return m, nil
}

func (m DashboardModel) View() string {
	if m.data.Repo == nil {
		return "No data"
	}

	// Header
	header := TitleStyle.Render(fmt.Sprintf("Analysis for %s", m.data.Repo.FullName))

	// Metrics Column
	metrics := fmt.Sprintf(
		"Health Score: %d\nBus Factor: %d (%s)\nMaturity: %s (%d)",
		m.data.HealthScore,
		m.data.BusFactor, m.data.BusRisk,
		m.data.MaturityLevel, m.data.MaturityScore,
	)
	metricsBox := BoxStyle.Render(metrics)

	// Charts
	activityData := analyzer.CommitsPerDay(m.data.Commits)
	chart := RenderCommitActivity(activityData, 10) // Show last 10 days
	chartBox := BoxStyle.Render(chart)

	// Layout
	content := lipgloss.JoinHorizontal(lipgloss.Top, metricsBox, chartBox)
	content = lipgloss.JoinVertical(lipgloss.Left, header, content)

	if m.showExport {
		exportMenu := BoxStyle.Render("Export Options:\n[J] JSON\n[M] Markdown")
		content = lipgloss.JoinVertical(lipgloss.Left, content, exportMenu)
	}

	if m.statusMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMsg))
	}

	content += "\n" + SubtleStyle.Render("e: export • q: back")

	if m.width == 0 {
		return content
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}
