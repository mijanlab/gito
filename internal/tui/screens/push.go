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

type pushViewMode int

const (
	pushModeMain pushViewMode = iota
	pushModeSelectBranch
	pushModeForceConfirm
	pushModeSuccess
	pushModeError
)

// PushScreen manages the Push workflow.
type PushScreen struct {
	Theme          *theme.Theme
	Header         *components.Header
	Footer         *components.Footer
	Menu           *components.Menu
	BranchMenu     *components.Menu
	Mode           pushViewMode
	Status         *git.StatusInfo
	Branches       []git.Branch
	TargetBranch   string
	ForceInput     string
	SuccessMsg     string
	ErrorTitle     string
	ErrorDetail    string
	IsUpstreamMiss bool
}

// NewPushScreen creates a new PushScreen.
func NewPushScreen(theme *theme.Theme, version string) *PushScreen {
	return &PushScreen{
		Theme:  theme,
		Header: components.NewHeader(theme, version),
		Footer: components.NewFooter(theme),
		Mode:   pushModeMain,
	}
}

// Reset re-initializes the push screen with status & branches.
func (s *PushScreen) Reset(status *git.StatusInfo, branches []git.Branch) {
	s.Status = status
	s.Branches = branches
	s.Mode = pushModeMain
	s.ForceInput = ""
	s.SuccessMsg = ""
	s.ErrorTitle = ""
	s.ErrorDetail = ""
	s.TargetBranch = ""

	if status != nil {
		s.TargetBranch = status.Branch
		s.IsUpstreamMiss = !status.HasUpstream
	}

	var items []components.MenuItem
	if s.IsUpstreamMiss {
		items = []components.MenuItem{
			{ID: "push_upstream", Title: "Push & set upstream", Icon: "↑"},
			{ID: "another_branch", Title: "Another branch", Icon: "⎇"},
			{ID: "back", Title: "Cancel", Icon: "✕"},
		}
	} else {
		items = []components.MenuItem{
			{ID: "push", Title: "Push", Icon: "↑"},
			{ID: "another_branch", Title: "Another branch", Icon: "⎇"},
			{ID: "force_push", Title: "Force push (with lease)", Icon: "⚠"},
			{ID: "back", Title: "Cancel", Icon: "✕"},
		}
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
	bItems = append(bItems, components.MenuItem{ID: "back_to_push", Title: "Back", Icon: "✕"})
	s.BranchMenu = components.NewMenu(s.Theme, bItems)
}

// SetSuccess sets success view.
func (s *PushScreen) SetSuccess(msg string) {
	s.Mode = pushModeSuccess
	s.SuccessMsg = msg
}

// SetError sets error view.
func (s *PushScreen) SetError(title, detail string) {
	s.Mode = pushModeError
	s.ErrorTitle = title
	s.ErrorDetail = detail
}

// Update handles keyboard navigation for PushScreen.
func (s *PushScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case pushModeMain:
			switch msg.String() {
			case "up", "k":
				s.Menu.MoveUp()
			case "down", "j":
				s.Menu.MoveDown()
			case "enter":
				selected := s.Menu.Selected()
				if selected != nil {
					switch selected.ID {
					case "push":
						return "do_push:current", nil
					case "push_upstream":
						return "do_push:upstream", nil
					case "another_branch":
						s.Mode = pushModeSelectBranch
						return "", nil
					case "force_push":
						s.Mode = pushModeForceConfirm
						s.ForceInput = ""
						return "", nil
					case "back":
						return "main_menu", nil
					}
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case pushModeSelectBranch:
			switch msg.String() {
			case "up", "k":
				s.BranchMenu.MoveUp()
			case "down", "j":
				s.BranchMenu.MoveDown()
			case "enter":
				selected := s.BranchMenu.Selected()
				if selected != nil {
					if selected.ID == "back_to_push" {
						s.Mode = pushModeMain
						return "", nil
					}
					s.TargetBranch = selected.ID
					return "do_push:branch:" + selected.ID, nil
				}
			case "esc":
				s.Mode = pushModeMain
				return "", nil
			}

		case pushModeForceConfirm:
			switch msg.String() {
			case "enter":
				if s.ForceInput == s.TargetBranch {
					s.Mode = pushModeMain
					return "do_push:force:" + s.TargetBranch, nil
				}
			case "esc":
				s.Mode = pushModeMain
				return "", nil
			case "backspace":
				if len(s.ForceInput) > 0 {
					s.ForceInput = s.ForceInput[:len(s.ForceInput)-1]
				}
			default:
				if len(msg.String()) == 1 {
					s.ForceInput += msg.String()
				}
			}

		case pushModeSuccess, pushModeError:
			switch msg.String() {
			case "enter", "esc", "q":
				return "main_menu", nil
			}
		}
	}
	return "", nil
}

// Render renders the Push screen based on current mode.
func (s *PushScreen) Render(repo *git.RepoInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, s.Status, width)

	switch s.Mode {
	case pushModeSuccess:
		title := s.Theme.SuccessText.Bold(true).Render("✓ Push Completed")
		msg := s.Theme.Value.Render(s.SuccessMsg)
		hint := s.Theme.Label.Render("Press Enter to continue")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", msg, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case pushModeError:
		title := s.Theme.ErrorText.Bold(true).Render("✕ Push Failed: " + s.ErrorTitle)
		detail := s.Theme.Label.Render("Details:\n  " + strings.ReplaceAll(s.ErrorDetail, "\n", "\n  "))
		hint := s.Theme.Label.Render("Press Enter to return")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", detail, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case pushModeForceConfirm:
		title := s.Theme.ErrorText.Bold(true).Render("⚠ Force Push Warning")
		warning := s.Theme.WarningText.Render(
			"You are about to overwrite remote history.\nThis can remove commits from the remote branch for all collaborators.",
		)
		prompt := s.Theme.Label.Render(fmt.Sprintf("Type the branch name '%s' to confirm:", s.TargetBranch))
		inputBox := s.Theme.Value.Bold(true).Render(s.ForceInput + "█")
		hint := s.Theme.Label.Render("Press Enter when ready, Esc to cancel")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", warning, "", prompt, inputBox, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case pushModeSelectBranch:
		title := s.Theme.SectionTitle.Render("Select Branch to Push")
		menuBox := s.BranchMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Push"},
			{Key: "Esc", Label: "Back"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	default:
		title := s.Theme.SectionTitle.Render("Push")
		var infoLines []string
		infoLines = append(infoLines, s.Theme.Label.Render("Current branch:"))
		infoLines = append(infoLines, "  ● "+s.Theme.Value.Bold(true).Render(s.TargetBranch))

		if s.IsUpstreamMiss {
			infoLines = append(infoLines, "", s.Theme.WarningText.Render("Branch '"+s.TargetBranch+"' is not connected to a remote branch."))
		} else {
			infoLines = append(infoLines, "", s.Theme.Label.Render("Remote:"))
			infoLines = append(infoLines, "  origin")
			if s.Status != nil && s.Status.Ahead > 0 {
				infoLines = append(infoLines, "", s.Theme.Label.Render("Status:"))
				infoLines = append(infoLines, fmt.Sprintf("  ↑ %d %s ahead", s.Status.Ahead, plural(s.Status.Ahead, "commit", "commits")))
			}
		}

		infoBox := strings.Join(infoLines, "\n")
		menuBox := s.Menu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", infoBox, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}
}

func plural(n int, s, p string) string {
	if n == 1 {
		return s
	}
	return p
}
