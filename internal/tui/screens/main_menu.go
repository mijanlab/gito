package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gito/internal/git"
	"gito/internal/theme"
	"gito/internal/tui/components"
)

// MainMenuScreen renders the primary hub of Gito.
type MainMenuScreen struct {
	Theme  *theme.Theme
	Header *components.Header
	Footer *components.Footer
	Menu   *components.Menu
}

// NewMainMenuScreen creates a new MainMenuScreen.
func NewMainMenuScreen(theme *theme.Theme, version string) *MainMenuScreen {
	s := &MainMenuScreen{
		Theme:  theme,
		Header: components.NewHeader(theme, version),
		Footer: components.NewFooter(theme),
	}
	s.InitMenu(nil)
	return s
}

// InitMenu updates menu items based on git status.
func (s *MainMenuScreen) InitMenu(status *git.StatusInfo) {
	var pushBadge, pullBadge, commitBadge, changesBadge string
	if status != nil {
		if status.Ahead > 0 {
			pushBadge = fmt.Sprintf("↑ %d ahead", status.Ahead)
		}
		if status.Behind > 0 {
			pullBadge = fmt.Sprintf("↓ %d behind", status.Behind)
		}
		if len(status.Conflicts) > 0 {
			commitBadge = fmt.Sprintf("! %d conflict(s)", len(status.Conflicts))
			changesBadge = fmt.Sprintf("! %d conflict(s)", len(status.Conflicts))
		} else if len(status.Files) > 0 {
			commitBadge = fmt.Sprintf("%d uncommitted", len(status.Files))
			changesBadge = fmt.Sprintf("%d changed", len(status.Files))
		}
	}

	items := []components.MenuItem{
		{ID: "push", Title: "Push", Icon: "↑", Badge: pushBadge},
		{ID: "pull", Title: "Pull", Icon: "↓", Badge: pullBadge},
		{ID: "commit", Title: "Commit", Icon: "◇", Badge: commitBadge},
		{ID: "branch", Title: "Branch", Icon: "⎇"},
		{ID: "changes", Title: "Changes", Icon: "◌", Badge: changesBadge},
		{ID: "history", Title: "History", Icon: "◷"},
		{ID: "ai", Title: "AI Settings", Icon: "✦"},
		{ID: "exit", Title: "Exit", Icon: "✕"},
	}

	if s.Menu == nil {
		s.Menu = components.NewMenu(s.Theme, items)
	} else {
		// Preserve cursor position if possible
		currCursor := s.Menu.Cursor
		s.Menu.SetItems(items)
		s.Menu.Cursor = currCursor
	}
}

// Update handles keyboard navigation for MainMenuScreen.
func (s *MainMenuScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			s.Menu.MoveUp()
		case "down", "j":
			s.Menu.MoveDown()
		case "enter":
			selected := s.Menu.Selected()
			if selected != nil {
				return selected.ID, nil
			}
		case "q", "ctrl+c":
			return "exit", tea.Quit
		case "r":
			return "refresh", nil
		case "?":
			return "help", nil
		}
	}
	return "", nil
}

// Render renders the main menu screen.
func (s *MainMenuScreen) Render(repo *git.RepoInfo, status *git.StatusInfo, width int) string {
	headerBox := s.Header.Render(repo, status, width)

	prompt := s.Theme.SectionTitle.Render("What would you like to do?")
	menuBox := s.Menu.Render(width)

	footerBox := s.Footer.Render([]components.KeyHelp{
		{Key: "↑↓", Label: "Navigate"},
		{Key: "Enter", Label: "Select"},
		{Key: "r", Label: "Refresh"},
		{Key: "q", Label: "Quit"},
	}, width)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerBox,
		"",
		prompt,
		"",
		menuBox,
		"",
		footerBox,
	)

	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	return s.Theme.Card.Width(boxWidth).Render(content)
}
