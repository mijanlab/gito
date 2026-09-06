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

type commitViewMode int

const (
	commitModeChooseMethod commitViewMode = iota
	commitModeInputMessage
	commitModeAIReview
	commitModeConfirm
	commitModeViewDiff
	commitModeSuccess
	commitModeError
)

// CommitScreen manages the complete interactive commit workflow.
type CommitScreen struct {
	Theme         *theme.Theme
	Header        *components.Header
	Footer        *components.Footer
	MethodMenu    *components.Menu
	AIReviewMenu  *components.Menu
	ConfirmMenu   *components.Menu
	DiffViewer    *components.DiffView
	Mode          commitViewMode
	Status        *git.StatusInfo
	DiffText      string
	CommitMessage string
	SuccessHash   string
	SuccessMsg    string
	ErrorTitle    string
	ErrorDetail   string
	IsAIActive    bool
}

// NewCommitScreen creates a new CommitScreen.
func NewCommitScreen(theme *theme.Theme, version string) *CommitScreen {
	return &CommitScreen{
		Theme:      theme,
		Header:     components.NewHeader(theme, version),
		Footer:     components.NewFooter(theme),
		DiffViewer: components.NewDiffView(theme, "", 15),
		Mode:       commitModeChooseMethod,
	}
}

// Reset initializes the commit screen with status and diff.
func (s *CommitScreen) Reset(status *git.StatusInfo, diffText string, aiEnabled bool) {
	s.Status = status
	s.DiffText = diffText
	s.DiffViewer.SetDiff(diffText)
	s.CommitMessage = ""
	s.SuccessHash = ""
	s.SuccessMsg = ""
	s.ErrorTitle = ""
	s.ErrorDetail = ""
	s.Mode = commitModeChooseMethod
	s.IsAIActive = aiEnabled

	methodItems := []components.MenuItem{}
	if aiEnabled {
		methodItems = append(methodItems, components.MenuItem{ID: "ai_generate", Title: "✦ AI Generate", Icon: ""})
	}
	methodItems = append(methodItems,
		components.MenuItem{ID: "write_myself", Title: "✎ Write myself", Icon: ""},
		components.MenuItem{ID: "back", Title: "Cancel", Icon: "✕"},
	)
	s.MethodMenu = components.NewMenu(s.Theme, methodItems)

	s.AIReviewMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "commit_now", Title: "Commit", Icon: "✓"},
		{ID: "regenerate", Title: "Regenerate", Icon: "✦"},
		{ID: "edit_message", Title: "Edit message", Icon: "✎"},
		{ID: "write_myself", Title: "Write myself", Icon: "✎"},
		{ID: "cancel", Title: "Cancel", Icon: "✕"},
	})

	s.ConfirmMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "commit_now", Title: "Commit", Icon: "✓"},
		{ID: "review_diff", Title: "Review changes (diff)", Icon: "◌"},
		{ID: "change_message", Title: "Change message", Icon: "✎"},
		{ID: "cancel", Title: "Cancel", Icon: "✕"},
	})
}

// SetAIMessage sets the generated AI commit message and switches to AI review mode.
func (s *CommitScreen) SetAIMessage(msg string) {
	s.CommitMessage = msg
	s.Mode = commitModeAIReview
}

// SetSuccess sets commit success state.
func (s *CommitScreen) SetSuccess(hash, msg string) {
	s.Mode = commitModeSuccess
	s.SuccessHash = hash
	s.SuccessMsg = msg
}

// SetError sets commit error state.
func (s *CommitScreen) SetError(title, detail string) {
	s.Mode = commitModeError
	s.ErrorTitle = title
	s.ErrorDetail = detail
}

// Update handles keyboard navigation for CommitScreen.
func (s *CommitScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case commitModeChooseMethod:
			switch msg.String() {
			case "up", "k":
				s.MethodMenu.MoveUp()
			case "down", "j":
				s.MethodMenu.MoveDown()
			case "enter":
				selected := s.MethodMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "ai_generate":
						return "do_ai_generate", nil
					case "write_myself":
						s.Mode = commitModeInputMessage
						return "", nil
					case "back":
						return "main_menu", nil
					}
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case commitModeInputMessage:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(s.CommitMessage) != "" {
					s.Mode = commitModeConfirm
				}
				return "", nil
			case "esc":
				s.Mode = commitModeChooseMethod
				return "", nil
			case "backspace":
				if len(s.CommitMessage) > 0 {
					s.CommitMessage = s.CommitMessage[:len(s.CommitMessage)-1]
				}
				return "", nil
			default:
				if len(msg.String()) == 1 {
					s.CommitMessage += msg.String()
				}
				return "", nil
			}

		case commitModeAIReview:
			switch msg.String() {
			case "up", "k":
				s.AIReviewMenu.MoveUp()
			case "down", "j":
				s.AIReviewMenu.MoveDown()
			case "enter":
				selected := s.AIReviewMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "commit_now":
						return "do_commit:" + s.CommitMessage, nil
					case "regenerate":
						return "do_ai_generate", nil
					case "edit_message":
						s.Mode = commitModeInputMessage
						return "", nil
					case "write_myself":
						s.CommitMessage = ""
						s.Mode = commitModeInputMessage
						return "", nil
					case "cancel":
						return "main_menu", nil
					}
				}
			case "esc":
				return "main_menu", nil
			}

		case commitModeConfirm:
			switch msg.String() {
			case "up", "k":
				s.ConfirmMenu.MoveUp()
			case "down", "j":
				s.ConfirmMenu.MoveDown()
			case "enter":
				selected := s.ConfirmMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "commit_now":
						return "do_commit:" + s.CommitMessage, nil
					case "review_diff":
						s.Mode = commitModeViewDiff
						return "", nil
					case "change_message":
						s.Mode = commitModeInputMessage
						return "", nil
					case "cancel":
						return "main_menu", nil
					}
				}
			case "esc":
				return "main_menu", nil
			}

		case commitModeViewDiff:
			switch msg.String() {
			case "up", "k":
				s.DiffViewer.ScrollUp()
			case "down", "j":
				s.DiffViewer.ScrollDown()
			case "pgup":
				s.DiffViewer.PageUp()
			case "pgdown":
				s.DiffViewer.PageDown()
			case "esc", "enter", "q":
				s.Mode = commitModeConfirm
				return "", nil
			}

		case commitModeSuccess, commitModeError:
			switch msg.String() {
			case "enter", "esc", "q":
				return "main_menu", nil
			}
		}
	}
	return "", nil
}

// Render renders the Commit screen.
func (s *CommitScreen) Render(repo *git.RepoInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, s.Status, width)

	switch s.Mode {
	case commitModeSuccess:
		title := s.Theme.SuccessText.Bold(true).Render("✓ Commit Created")
		commitSummary := fmt.Sprintf("%s  %s", s.Theme.Title.Render(s.SuccessHash), s.Theme.Value.Render(s.SuccessMsg))
		hint := s.Theme.Label.Render("Press Enter to continue")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", commitSummary, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case commitModeError:
		title := s.Theme.ErrorText.Bold(true).Render("✕ Commit Failed: " + s.ErrorTitle)
		detail := s.Theme.Label.Render("Details:\n  " + strings.ReplaceAll(s.ErrorDetail, "\n", "\n  "))
		hint := s.Theme.Label.Render("Press Enter to return")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", detail, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case commitModeInputMessage:
		title := s.Theme.SectionTitle.Render("Commit Message")
		prompt := s.Theme.Label.Render("Type your commit message:")
		inputBox := s.Theme.Value.Bold(true).Render("> " + s.CommitMessage + "█")
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "Enter", Label: "Save"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", prompt, inputBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case commitModeAIReview:
		title := s.Theme.SectionTitle.Render("Commit")
		filesCount := 0
		if s.Status != nil {
			filesCount = len(s.Status.Files)
		}
		stats := s.Theme.Label.Render(fmt.Sprintf("%d %s changed", filesCount, plural(filesCount, "file", "files")))
		aiLabel := s.Theme.AIText.Render("✦ AI generated:")
		msgBox := s.Theme.Value.Bold(true).Render(fmt.Sprintf("\"%s\"", s.CommitMessage))
		menuBox := s.AIReviewMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Cancel"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, stats, "", aiLabel, msgBox, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case commitModeConfirm:
		title := s.Theme.SectionTitle.Render("Changes to Commit")
		fileList := s.renderFileList()
		msgLabel := s.Theme.Label.Render("Commit message:")
		msgBox := s.Theme.Value.Bold(true).Render(fmt.Sprintf("\"%s\"", s.CommitMessage))
		menuBox := s.ConfirmMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Cancel"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", fileList, "", msgLabel, msgBox, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case commitModeViewDiff:
		title := s.Theme.SectionTitle.Render("Review Diff")
		diffBox := s.DiffViewer.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓/PgUp/PgDn", Label: "Scroll"},
			{Key: "Esc", Label: "Back to Commit"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", diffBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	default: // commitModeChooseMethod
		title := s.Theme.SectionTitle.Render("Commit")
		filesCount := 0
		if s.Status != nil {
			filesCount = len(s.Status.Files)
		}
		if filesCount == 0 {
			cleanMsg := s.Theme.SuccessText.Render("Working tree is clean. Nothing to commit.")
			footer := s.Footer.Render([]components.KeyHelp{{Key: "Esc/Enter", Label: "Back"}}, width)
			content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", cleanMsg, "", footer)
			return s.Theme.Card.Width(boxWidth).Render(content)
		}

		stats := s.Theme.Label.Render(fmt.Sprintf("%d %s changed", filesCount, plural(filesCount, "file", "files")))
		fileList := s.renderFileList()
		prompt := s.Theme.Value.Render("How do you want to create the commit message?")
		menuBox := s.MethodMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Cancel"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, stats, "", fileList, "", prompt, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}
}

func (s *CommitScreen) renderFileList() string {
	if s.Status == nil || len(s.Status.Files) == 0 {
		return s.Theme.Label.Render("  (No files)")
	}

	var lines []string
	maxFiles := 8
	for i, f := range s.Status.Files {
		if i >= maxFiles {
			lines = append(lines, s.Theme.Label.Render(fmt.Sprintf("  ... and %d more files", len(s.Status.Files)-maxFiles)))
			break
		}
		statusChar := f.DisplayStatus()
		var charStyled string
		switch statusChar {
		case "A", "?":
			charStyled = s.Theme.SuccessText.Render(statusChar)
		case "D":
			charStyled = s.Theme.ErrorText.Render(statusChar)
		case "U":
			charStyled = s.Theme.ErrorText.Bold(true).Render(statusChar)
		default:
			charStyled = s.Theme.WarningText.Render(statusChar)
		}
		lines = append(lines, fmt.Sprintf("  %s  %s", charStyled, s.Theme.Value.Render(f.Path)))
	}
	return strings.Join(lines, "\n")
}
