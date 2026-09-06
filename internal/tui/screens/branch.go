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

type branchViewMode int

const (
	branchModeMain branchViewMode = iota
	branchModeCreateInput
	branchModeCreateAction
	branchModeSwitchSelect
	branchModeDeleteSelect
	branchModeDeleteForceConfirm
	branchModeSuccess
	branchModeError
)

// BranchScreen manages branch creation, switching, and deletion.
type BranchScreen struct {
	Theme         *theme.Theme
	Header        *components.Header
	Footer        *components.Footer
	MainMenu      *components.Menu
	CreateMenu    *components.Menu
	SwitchMenu    *components.Menu
	DeleteMenu    *components.Menu
	ForceMenu     *components.Menu
	Mode          branchViewMode
	Status        *git.StatusInfo
	Branches      []git.Branch
	NewBranchName string
	TargetBranch  string
	SuccessMsg    string
	ErrorTitle    string
	ErrorDetail   string
}

// NewBranchScreen creates a new BranchScreen.
func NewBranchScreen(theme *theme.Theme, version string) *BranchScreen {
	return &BranchScreen{
		Theme:  theme,
		Header: components.NewHeader(theme, version),
		Footer: components.NewFooter(theme),
		Mode:   branchModeMain,
	}
}

// Reset initializes branch screen data.
func (s *BranchScreen) Reset(status *git.StatusInfo, branches []git.Branch) {
	s.Status = status
	s.Branches = branches
	s.Mode = branchModeMain
	s.NewBranchName = ""
	s.TargetBranch = ""
	s.SuccessMsg = ""
	s.ErrorTitle = ""
	s.ErrorDetail = ""

	s.MainMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "create", Title: "Create new branch", Icon: "+"},
		{ID: "switch", Title: "Switch branch", Icon: "⎇"},
		{ID: "delete", Title: "Delete branch", Icon: "✕"},
		{ID: "back", Title: "Back", Icon: "←"},
	})

	s.CreateMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "create_switch", Title: "Create & switch", Icon: "⎇"},
		{ID: "create_only", Title: "Create only", Icon: "+"},
		{ID: "cancel", Title: "Cancel", Icon: "✕"},
	})

	var switchItems, deleteItems []components.MenuItem
	for _, b := range branches {
		badge := ""
		if b.IsCurrent {
			badge = "(current)"
		} else if b.Upstream != "" {
			badge = "→ " + b.Upstream
		}

		if !b.IsCurrent {
			switchItems = append(switchItems, components.MenuItem{
				ID:       b.Name,
				Title:    b.Name,
				Subtitle: b.Subject,
				Icon:     "⎇",
				Badge:    badge,
			})
			deleteItems = append(deleteItems, components.MenuItem{
				ID:       b.Name,
				Title:    b.Name,
				Subtitle: b.Subject,
				Icon:     "✕",
				Badge:    badge,
			})
		}
	}
	switchItems = append(switchItems, components.MenuItem{ID: "cancel", Title: "Cancel", Icon: "←"})
	deleteItems = append(deleteItems, components.MenuItem{ID: "cancel", Title: "Cancel", Icon: "←"})

	s.SwitchMenu = components.NewMenu(s.Theme, switchItems)
	s.DeleteMenu = components.NewMenu(s.Theme, deleteItems)

	s.ForceMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "cancel", Title: "Cancel", Icon: "←"},
		{ID: "delete_force", Title: "Delete anyway (Force -D)", Icon: "⚠"},
	})
}

// SetSuccess sets success view.
func (s *BranchScreen) SetSuccess(msg string) {
	s.Mode = branchModeSuccess
	s.SuccessMsg = msg
}

// SetError sets error view.
func (s *BranchScreen) SetError(title, detail string) {
	s.Mode = branchModeError
	s.ErrorTitle = title
	s.ErrorDetail = detail
}

// SetForceConfirm switches to force delete confirmation mode.
func (s *BranchScreen) SetForceConfirm(branchName string) {
	s.TargetBranch = branchName
	s.Mode = branchModeDeleteForceConfirm
}

// Update handles keyboard navigation for BranchScreen.
func (s *BranchScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case branchModeMain:
			switch msg.String() {
			case "up", "k":
				s.MainMenu.MoveUp()
			case "down", "j":
				s.MainMenu.MoveDown()
			case "enter":
				selected := s.MainMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "create":
						s.Mode = branchModeCreateInput
						s.NewBranchName = ""
						return "", nil
					case "switch":
						s.Mode = branchModeSwitchSelect
						return "", nil
					case "delete":
						s.Mode = branchModeDeleteSelect
						return "", nil
					case "back":
						return "main_menu", nil
					}
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case branchModeCreateInput:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(s.NewBranchName) != "" {
					s.Mode = branchModeCreateAction
				}
				return "", nil
			case "esc":
				s.Mode = branchModeMain
				return "", nil
			case "backspace":
				if len(s.NewBranchName) > 0 {
					s.NewBranchName = s.NewBranchName[:len(s.NewBranchName)-1]
				}
				return "", nil
			default:
				if len(msg.String()) == 1 {
					s.NewBranchName += msg.String()
				}
				return "", nil
			}

		case branchModeCreateAction:
			switch msg.String() {
			case "up", "k":
				s.CreateMenu.MoveUp()
			case "down", "j":
				s.CreateMenu.MoveDown()
			case "enter":
				selected := s.CreateMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "create_switch":
						return "do_branch:create_switch:" + s.NewBranchName, nil
					case "create_only":
						return "do_branch:create_only:" + s.NewBranchName, nil
					case "cancel":
						s.Mode = branchModeMain
						return "", nil
					}
				}
			case "esc":
				s.Mode = branchModeCreateInput
				return "", nil
			}

		case branchModeSwitchSelect:
			switch msg.String() {
			case "up", "k":
				s.SwitchMenu.MoveUp()
			case "down", "j":
				s.SwitchMenu.MoveDown()
			case "enter":
				selected := s.SwitchMenu.Selected()
				if selected != nil {
					if selected.ID == "cancel" {
						s.Mode = branchModeMain
						return "", nil
					}
					return "do_branch:switch:" + selected.ID, nil
				}
			case "esc":
				s.Mode = branchModeMain
				return "", nil
			}

		case branchModeDeleteSelect:
			switch msg.String() {
			case "up", "k":
				s.DeleteMenu.MoveUp()
			case "down", "j":
				s.DeleteMenu.MoveDown()
			case "enter":
				selected := s.DeleteMenu.Selected()
				if selected != nil {
					if selected.ID == "cancel" {
						s.Mode = branchModeMain
						return "", nil
					}
					s.TargetBranch = selected.ID
					return "do_branch:delete_safe:" + selected.ID, nil
				}
			case "esc":
				s.Mode = branchModeMain
				return "", nil
			}

		case branchModeDeleteForceConfirm:
			switch msg.String() {
			case "up", "k":
				s.ForceMenu.MoveUp()
			case "down", "j":
				s.ForceMenu.MoveDown()
			case "enter":
				selected := s.ForceMenu.Selected()
				if selected != nil {
					if selected.ID == "delete_force" {
						return "do_branch:delete_force:" + s.TargetBranch, nil
					}
					s.Mode = branchModeMain
					return "", nil
				}
			case "esc":
				s.Mode = branchModeMain
				return "", nil
			}

		case branchModeSuccess, branchModeError:
			switch msg.String() {
			case "enter", "esc", "q":
				return "main_menu", nil
			}
		}
	}
	return "", nil
}

// Render renders the branch screen.
func (s *BranchScreen) Render(repo *git.RepoInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, s.Status, width)

	switch s.Mode {
	case branchModeSuccess:
		title := s.Theme.SuccessText.Bold(true).Render("✓ Branch Action Completed")
		msg := s.Theme.Value.Render(s.SuccessMsg)
		hint := s.Theme.Label.Render("Press Enter to continue")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", msg, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case branchModeError:
		title := s.Theme.ErrorText.Bold(true).Render("✕ Branch Operation Failed: " + s.ErrorTitle)
		detail := s.Theme.Label.Render("Details:\n  " + strings.ReplaceAll(s.ErrorDetail, "\n", "\n  "))
		hint := s.Theme.Label.Render("Press Enter to return")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", detail, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case branchModeCreateInput:
		title := s.Theme.SectionTitle.Render("Create New Branch")
		prompt := s.Theme.Label.Render("Enter branch name:")
		inputBox := s.Theme.Value.Bold(true).Render("> " + s.NewBranchName + "█")
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "Enter", Label: "Next"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", prompt, inputBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case branchModeCreateAction:
		title := s.Theme.SectionTitle.Render("Create Branch: " + s.NewBranchName)
		from := s.Theme.Label.Render("Create from: ● " + s.Status.Branch)
		menuBox := s.CreateMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", from, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case branchModeSwitchSelect:
		title := s.Theme.SectionTitle.Render("Switch Branch")
		menuBox := s.SwitchMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Switch"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case branchModeDeleteSelect:
		title := s.Theme.SectionTitle.Render("Delete Branch (Safe)")
		menuBox := s.DeleteMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Delete"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case branchModeDeleteForceConfirm:
		title := s.Theme.ErrorText.Bold(true).Render("⚠ Unmerged Branch Warning")
		desc := fmt.Sprintf("Branch '%s' contains commits that may not exist on another branch.\nDeleting it will destroy these unmerged commits.", s.TargetBranch)
		descBox := s.Theme.WarningText.Render(desc)
		menuBox := s.ForceMenu.Render(width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", descBox, "", menuBox)
		return s.Theme.Card.Width(boxWidth).Render(content)

	default: // branchModeMain
		title := s.Theme.SectionTitle.Render("Branch Management")
		var bLines []string
		bLines = append(bLines, s.Theme.Label.Render("Local branches:"))
		for _, b := range s.Branches {
			if b.IsCurrent {
				bLines = append(bLines, fmt.Sprintf("  ● %s %s", s.Theme.Value.Bold(true).Render(b.Name), s.Theme.StatusClean.Render("(current)")))
			} else {
				bLines = append(bLines, fmt.Sprintf("    %s", s.Theme.Label.Render(b.Name)))
			}
		}

		branchList := strings.Join(bLines, "\n")
		actionsTitle := s.Theme.Label.Render("Actions:")
		menuBox := s.MainMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", branchList, "", actionsTitle, menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}
}
