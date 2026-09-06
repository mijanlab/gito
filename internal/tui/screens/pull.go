package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gito/internal/git"
	"gito/internal/theme"
	"gito/internal/tui/components"
)

type pullViewMode int

const (
	pullModeMain pullViewMode = iota
	pullModeSelectBranch
	pullModeSafetyDecision
	pullModeSuccess
	pullModeError
)

// PullScreen handles Git pull operations with safety guardrails.
type PullScreen struct {
	Theme          *theme.Theme
	Header         *components.Header
	Footer         *components.Footer
	Menu           *components.Menu
	BranchMenu     *components.Menu
	SafetyMenu     *components.Menu
	Mode           pullViewMode
	Status         *git.StatusInfo
	Branches       []git.Branch
	CurrentBranch  string
	SelectedBranch string
	SuccessMsg     string
	ErrorTitle     string
	ErrorDetail    string
}

// NewPullScreen creates a new PullScreen.
func NewPullScreen(theme *theme.Theme, version string) *PullScreen {
	return &PullScreen{
		Theme:  theme,
		Header: components.NewHeader(theme, version),
		Footer: components.NewFooter(theme),
		Mode:   pullModeMain,
	}
}

// Reset initializes the Pull screen.
func (s *PullScreen) Reset(status *git.StatusInfo, branches []git.Branch) {
	s.Status = status
	s.Branches = branches
	s.Mode = pullModeMain
	s.SuccessMsg = ""
	s.ErrorTitle = ""
	s.ErrorDetail = ""
	s.SelectedBranch = ""

	if status != nil {
		s.CurrentBranch = status.Branch
	}

	items := []components.MenuItem{
		{ID: "pull_current", Title: fmt.Sprintf("Current branch (%s)", s.CurrentBranch), Icon: "↓"},
		{ID: "another_branch", Title: "Another branch", Icon: "⎇"},
		{ID: "back", Title: "Cancel", Icon: "✕"},
	}
	s.Menu = components.NewMenu(s.Theme, items)

	// Prepare branch menu
	var bItems []components.MenuItem
	for _, b := range branches {
		badge := ""
		if b.IsCurrent {
			badge = "(current)"
		} else if b.Upstream != "" {
			badge = "→ " + b.Upstream
		}
		bItems = append(bItems, components.MenuItem{
			ID:    b.Name,
			Title: b.Name,
			Icon:  "⎇",
			Badge: badge,
		})
	}
	bItems = append(bItems, components.MenuItem{ID: "back_to_pull", Title: "Back", Icon: "✕"})
	s.BranchMenu = components.NewMenu(s.Theme, bItems)
}

// SetSuccess sets success view.
func (s *PullScreen) SetSuccess(msg string) {
	s.Mode = pullModeSuccess
	s.SuccessMsg = msg
}

// SetError sets error view.
func (s *PullScreen) SetError(title, detail string) {
	s.Mode = pullModeError
	s.ErrorTitle = title
	s.ErrorDetail = detail
}

// Update handles keyboard navigation for PullScreen.
func (s *PullScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case pullModeMain:
			switch msg.String() {
			case "up", "k":
				s.Menu.MoveUp()
			case "down", "j":
				s.Menu.MoveDown()
			case "enter":
				selected := s.Menu.Selected()
				if selected != nil {
					switch selected.ID {
					case "pull_current":
						return "do_pull:current", nil
					case "another_branch":
						s.Mode = pullModeSelectBranch
						return "", nil
					case "back":
						return "main_menu", nil
					}
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case pullModeSelectBranch:
			switch msg.String() {
			case "up", "k":
				s.BranchMenu.MoveUp()
			case "down", "j":
				s.BranchMenu.MoveDown()
			case "enter":
				selected := s.BranchMenu.Selected()
				if selected != nil {
					if selected.ID == "back_to_pull" {
						s.Mode = pullModeMain
						return "", nil
					}
					s.SelectedBranch = selected.ID
					if s.SelectedBranch == s.CurrentBranch {
						return "do_pull:current", nil
					}
					// Prompt safety decision when user picked a different branch
					s.SafetyMenu = components.NewMenu(s.Theme, []components.MenuItem{
						{ID: "switch_and_pull", Title: fmt.Sprintf("Switch to %s and pull", s.SelectedBranch), Icon: "⎇"},
						{ID: "pull_into_current", Title: fmt.Sprintf("Pull %s into %s", s.SelectedBranch, s.CurrentBranch), Icon: "⚠"},
						{ID: "cancel_safety", Title: "Cancel", Icon: "✕"},
					})
					s.Mode = pullModeSafetyDecision
					return "", nil
				}
			case "esc":
				s.Mode = pullModeMain
				return "", nil
			}

		case pullModeSafetyDecision:
			switch msg.String() {
			case "up", "k":
				s.SafetyMenu.MoveUp()
			case "down", "j":
				s.SafetyMenu.MoveDown()
			case "enter":
				selected := s.SafetyMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "switch_and_pull":
						return "do_pull:switch_and_pull:" + s.SelectedBranch, nil
					case "pull_into_current":
						return "do_pull:pull_into_current:" + s.SelectedBranch, nil
					case "cancel_safety":
						s.Mode = pullModeSelectBranch
						return "", nil
					}
				}
			case "esc":
				s.Mode = pullModeSelectBranch
				return "", nil
			}

		case pullModeSuccess, pullModeError:
			switch msg.String() {
			case "enter", "esc", "q":
				return "main_menu", nil
			}
		}
	}
	return "", nil
}

// Render renders the Pull screen.
func (s *PullScreen) Render(repo *git.RepoInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, s.Status, width)

	switch s.Mode {
	case pullModeSuccess:
		title := s.Theme.SuccessText.Bold(true).Render("✓ Pull Completed")
		msg := s.Theme.Value.Render(s.SuccessMsg)
		hint := s.Theme.Label.Render("Press Enter to continue")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", msg, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case pullModeError:
		title := s.Theme.ErrorText.Bold(true).Render("✕ Pull Failed: " + s.ErrorTitle)
		detail := s.Theme.Label.Render("Details:\n  " + strings.ReplaceAll(s.ErrorDetail, "\n", "\n  "))
		hint := s.Theme.Label.Render("Press Enter to return")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", detail, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case pullModeSafetyDecision:
		title := s.Theme.WarningText.Bold(true).Render("Branch Safety Confirmation")
		desc := fmt.Sprintf("You are currently on:\n  ● %s\n\nYou selected:\n  ● %s\n\nHow should Gito proceed?", s.CurrentBranch, s.SelectedBranch)
		descBox := s.Theme.Value.Render(desc)
		menuBox := s.SafetyMenu.Render(width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", descBox, "", menuBox)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case pullModeSelectBranch:
		title := s.Theme.SectionTitle.Render("Select Branch to Pull")
		menuBox := s.BranchMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	default:
		title := s.Theme.SectionTitle.Render("Pull")
		menuBox := s.Menu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}
}
