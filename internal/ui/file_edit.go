package ui

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileEditModel struct {
	filePath  string
	repoOwner string
	repoName  string
	isOwner   bool
	width     int
	height    int
	Done      bool
}

func NewFileEditModel(filePath, repoFullName string) FileEditModel {
	parts := strings.Split(repoFullName, "/")
	repoOwner := ""
	repoName := ""
	if len(parts) == 2 {
		repoOwner = parts[0]
		repoName = parts[1]
	}

	return FileEditModel{
		filePath:  filePath,
		repoOwner: repoOwner,
		repoName:  repoName,
	}
}

func (m *FileEditModel) SetOwnership(isOwner bool) {
	m.isOwner = isOwner
}

func (m FileEditModel) Init() tea.Cmd { return nil }

func (m FileEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "e", "enter":
			// Open file in editor
			return m, m.openInEditor()
		case "c":
			// Commit changes
			if m.isOwner {
				return m, m.commitChanges()
			}
		case "p":
			// Push changes
			if m.isOwner {
				return m, m.pushChanges()
			}
		case "C":
			// Commit and push changes
			if m.isOwner {
				return m, m.commitAndPushChanges()
			}
		case "esc":
			m.Done = true
		}
	}

	return m, nil
}

func (m FileEditModel) View() string {
	content := TitleStyle.Render("📝 FILE EDITOR") + "\n\n"

	content += fmt.Sprintf("File: %s\n", m.filePath)
	content += fmt.Sprintf("Repository: %s/%s\n", m.repoOwner, m.repoName)

	if m.isOwner {
		content += SuccessStyle.Render("✓ You are the repository owner\n\n")
		content += "Available actions:\n"
		content += "• Press 'e' to open in external editor\n"
		content += "• Press 'c' to commit changes\n"
		content += "• Press 'p' to push changes\n"
		content += "• Press 'C' to commit and push changes\n"
	} else {
		content += ErrorStyle.Render("✗ You are not the repository owner\n\n")
		content += "Available actions:\n"
		content += "• Press 'e' to open in external editor (read-only)\n"
		content += ErrorStyle.Render("• Contact repository owner for contribution\n")
		content += SubtleStyle.Render("  Repository: https://github.com/" + m.repoOwner + "/" + m.repoName + "\n")
		content += SubtleStyle.Render("  Owner: @" + m.repoOwner + "\n")
	}

	content += "\n" + SubtleStyle.Render("ESC back")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Left, lipgloss.Top,
		BoxStyle.Render(content),
	)
}

func (m FileEditModel) openInEditor() tea.Cmd {
	editors := []string{"code", "vim", "notepad++", "notepad"}

	for _, editor := range editors {
		if m.isEditorAvailable(editor) {
			return tea.ExecProcess(exec.Command(editor, m.filePath), func(err error) tea.Msg {
				return nil // Just return, editor opened or failed
			})
		}
	}

	// Fallback to system default
	return tea.ExecProcess(exec.Command("cmd", "/c", "start", m.filePath), func(err error) tea.Msg {
		return nil
	})
}

func (m FileEditModel) isEditorAvailable(editor string) bool {
	_, err := exec.LookPath(editor)
	return err == nil
}

func (m FileEditModel) commitChanges() tea.Cmd {
	return tea.ExecProcess(exec.Command("git", "add", m.filePath), func(err error) tea.Msg {
		if err != nil {
			return nil
		}
		return tea.ExecProcess(exec.Command("git", "commit", "-m", "Update "+m.filePath), func(err error) tea.Msg {
			return nil
		})
	})
}

func (m FileEditModel) pushChanges() tea.Cmd {
	return tea.ExecProcess(exec.Command("git", "push"), func(err error) tea.Msg {
		return nil
	})
}

func (m FileEditModel) commitAndPushChanges() tea.Cmd {
	return tea.ExecProcess(exec.Command("git", "add", m.filePath), func(err error) tea.Msg {
		if err != nil {
			return nil
		}
		return tea.ExecProcess(exec.Command("git", "commit", "-m", "Update "+m.filePath), func(err error) tea.Msg {
			if err != nil {
				return nil
			}
			return tea.ExecProcess(exec.Command("git", "push"), func(err error) tea.Msg {
				return nil
			})
		})
	})
}
