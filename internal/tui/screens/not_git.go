package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gito/internal/theme"
	"gito/internal/tui/components"
)

// NotGitScreen handles directories that are not yet Git repositories.
type NotGitScreen struct {
	Theme       *theme.Theme
	Dir         string
	Menu        *components.Menu
	Footer      *components.Footer
	InputActive bool
	InputPath   string
	StatusMsg   string
}

// NewNotGitScreen creates a new NotGitScreen.
func NewNotGitScreen(theme *theme.Theme, dir string) *NotGitScreen {
	menu := components.NewMenu(theme, []components.MenuItem{
		{ID: "init", Title: "Initialize repository", Icon: "✦"},
		{ID: "choose", Title: "Choose another directory", Icon: "📁"},
		{ID: "exit", Title: "Exit", Icon: "✕"},
	})

	return &NotGitScreen{
		Theme:     theme,
		Dir:       dir,
		Menu:      menu,
		Footer:    components.NewFooter(theme),
		InputPath: dir,
	}
}

// Update handles keyboard navigation for NotGitScreen.
func (s *NotGitScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.InputActive {
			switch msg.String() {
			case "enter":
				s.InputActive = false
				return "change_dir:" + strings.TrimSpace(s.InputPath), nil
			case "esc":
				s.InputActive = false
				return "", nil
			case "backspace":
				if len(s.InputPath) > 0 {
					s.InputPath = s.InputPath[:len(s.InputPath)-1]
				}
				return "", nil
			default:
				if len(msg.String()) == 1 {
					s.InputPath += msg.String()
				}
				return "", nil
			}
		}

		switch msg.String() {
		case "up", "k":
			s.Menu.MoveUp()
		case "down", "j":
			s.Menu.MoveDown()
		case "enter":
			selected := s.Menu.Selected()
			if selected != nil {
				switch selected.ID {
				case "init":
					return "init_repo", nil
				case "choose":
					s.InputActive = true
					return "", nil
				case "exit":
					return "exit", nil
				}
			}
		case "q", "esc":
			return "exit", nil
		}
	}
	return "", nil
}

// Render renders the not-git screen.
func (s *NotGitScreen) Render(width int) string {
	brand := s.Theme.Title.Render("◈ Gito")
	title := s.Theme.SectionTitle.Render("No Git repository found.")
	dirLabel := s.Theme.Label.Render("Current directory:")
	dirVal := s.Theme.Value.Render(s.Dir)

	var content []string
	content = append(content, brand, "", title, "", dirLabel, dirVal, "")

	if s.InputActive {
		prompt := s.Theme.Label.Render("Enter directory path:")
		inputVal := s.Theme.Value.Bold(true).Render(s.InputPath + "█")
		hint := s.Theme.Label.Render("Press Enter to switch directory, Esc to cancel")
		content = append(content, prompt, inputVal, "", hint)
	} else {
		content = append(content, s.Menu.Render(width))
	}

	if s.StatusMsg != "" {
		content = append(content, "", s.Theme.ErrorText.Render(s.StatusMsg))
	}

	content = append(content, "", s.Footer.Render([]components.KeyHelp{
		{Key: "↑↓", Label: "Navigate"},
		{Key: "Enter", Label: "Select"},
		{Key: "q", Label: "Quit"},
	}, width))

	cardWidth := width - 6
	if cardWidth < 40 {
		cardWidth = 40
	}

	return lipgloss.Place(width, 24, lipgloss.Center, lipgloss.Center, s.Theme.Card.Width(cardWidth).Render(strings.Join(content, "\n")))
}
