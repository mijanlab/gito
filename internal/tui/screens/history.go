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

type historyViewMode int

const (
	historyModeList historyViewMode = iota
	historyModeDetail
	historyModeDiff
	historyModeRevertConfirm
	historyModeSuccess
	historyModeError
)

// HistoryScreen manages commit history exploration, diff inspection, and reverting.
type HistoryScreen struct {
	Theme         *theme.Theme
	Header        *components.Header
	Footer        *components.Footer
	HistoryMenu   *components.Menu
	DetailMenu    *components.Menu
	RevertMenu    *components.Menu
	DiffViewer    *components.DiffView
	Mode          historyViewMode
	Status        *git.StatusInfo
	Commits       []git.CommitInfo
	CurrentDetail *git.CommitDetail
	SuccessMsg    string
	ErrorTitle    string
	ErrorDetail   string
}

// NewHistoryScreen creates a new HistoryScreen.
func NewHistoryScreen(theme *theme.Theme, version string) *HistoryScreen {
	return &HistoryScreen{
		Theme:      theme,
		Header:     components.NewHeader(theme, version),
		Footer:     components.NewFooter(theme),
		DiffViewer: components.NewDiffView(theme, "", 18),
		Mode:       historyModeList,
	}
}

// Reset initializes commit history list.
func (s *HistoryScreen) Reset(status *git.StatusInfo, commits []git.CommitInfo) {
	s.Status = status
	s.Commits = commits
	s.Mode = historyModeList
	s.CurrentDetail = nil
	s.SuccessMsg = ""
	s.ErrorTitle = ""
	s.ErrorDetail = ""

	var items []components.MenuItem
	for _, c := range commits {
		items = append(items, components.MenuItem{
			ID:       c.Hash,
			Title:    fmt.Sprintf("%s  %s", c.ShortHash, c.Subject),
			Subtitle: fmt.Sprintf("%s · %s", c.AuthorName, c.RelativeDate),
			Icon:     "●",
		})
	}
	s.HistoryMenu = components.NewMenu(s.Theme, items)
	s.HistoryMenu.MaxHeight = 12

	s.DetailMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "view_diff", Title: "View diff", Icon: "◌"},
		{ID: "revert_confirm", Title: "Revert commit", Icon: "↺"},
		{ID: "back_to_list", Title: "Back", Icon: "←"},
	})

	s.RevertMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "revert_execute", Title: "Revert commit", Icon: "↺"},
		{ID: "cancel_revert", Title: "Cancel", Icon: "✕"},
	})
}

// SetDetail sets the commit detail view.
func (s *HistoryScreen) SetDetail(detail *git.CommitDetail) {
	s.CurrentDetail = detail
	s.Mode = historyModeDetail
}

// SetCommitDiff sets commit diff and opens the diff viewer.
func (s *HistoryScreen) SetCommitDiff(diff string) {
	s.DiffViewer.SetDiff(diff)
	s.Mode = historyModeDiff
}

// SetSuccess sets success view.
func (s *HistoryScreen) SetSuccess(msg string) {
	s.Mode = historyModeSuccess
	s.SuccessMsg = msg
}

// SetError sets error view.
func (s *HistoryScreen) SetError(title, detail string) {
	s.Mode = historyModeError
	s.ErrorTitle = title
	s.ErrorDetail = detail
}

// Update handles keyboard navigation for HistoryScreen.
func (s *HistoryScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case historyModeList:
			switch msg.String() {
			case "up", "k":
				s.HistoryMenu.MoveUp()
			case "down", "j":
				s.HistoryMenu.MoveDown()
			case "enter":
				selected := s.HistoryMenu.Selected()
				if selected != nil {
					return "view_commit_detail:" + selected.ID, nil
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case historyModeDetail:
			switch msg.String() {
			case "up", "k":
				s.DetailMenu.MoveUp()
			case "down", "j":
				s.DetailMenu.MoveDown()
			case "enter":
				selected := s.DetailMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "view_diff":
						if s.CurrentDetail != nil {
							return "view_commit_diff:" + s.CurrentDetail.Hash, nil
						}
					case "revert_confirm":
						s.Mode = historyModeRevertConfirm
						return "", nil
					case "back_to_list":
						s.Mode = historyModeList
						return "", nil
					}
				}
			case "esc", "q":
				s.Mode = historyModeList
				return "", nil
			}

		case historyModeDiff:
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
				s.Mode = historyModeDetail
				return "", nil
			}

		case historyModeRevertConfirm:
			switch msg.String() {
			case "up", "k":
				s.RevertMenu.MoveUp()
			case "down", "j":
				s.RevertMenu.MoveDown()
			case "enter":
				selected := s.RevertMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "revert_execute":
						if s.CurrentDetail != nil {
							return "do_revert:" + s.CurrentDetail.Hash, nil
						}
					case "cancel_revert":
						s.Mode = historyModeDetail
						return "", nil
					}
				}
			case "esc":
				s.Mode = historyModeDetail
				return "", nil
			}

		case historyModeSuccess, historyModeError:
			switch msg.String() {
			case "enter", "esc", "q":
				s.Mode = historyModeList
				return "refresh", nil
			}
		}
	}
	return "", nil
}

// Render renders the History screen.
func (s *HistoryScreen) Render(repo *git.RepoInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, s.Status, width)

	switch s.Mode {
	case historyModeSuccess:
		title := s.Theme.SuccessText.Bold(true).Render("✓ Revert Completed")
		msg := s.Theme.Value.Render(s.SuccessMsg)
		hint := s.Theme.Label.Render("Press Enter to return to history")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", msg, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case historyModeError:
		title := s.Theme.ErrorText.Bold(true).Render("✕ Revert Failed: " + s.ErrorTitle)
		detail := s.Theme.Label.Render("Details:\n  " + strings.ReplaceAll(s.ErrorDetail, "\n", "\n  "))
		hint := s.Theme.Label.Render("Press Enter to return")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", detail, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case historyModeDiff:
		title := s.Theme.SectionTitle.Render(fmt.Sprintf("Diff for %s", s.CurrentDetail.ShortHash))
		diffBox := s.DiffViewer.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓/PgUp/PgDn", Label: "Scroll"},
			{Key: "Esc", Label: "Back to Commit"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", diffBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case historyModeRevertConfirm:
		title := s.Theme.WarningText.Bold(true).Render("Revert Commit Confirmation")
		desc := fmt.Sprintf("Are you sure you want to revert commit %s?\n\n\"%s\"\n\nThis will create a new commit that undoes the changes.", s.CurrentDetail.ShortHash, s.CurrentDetail.Subject)
		descBox := s.Theme.Value.Render(desc)
		menuBox := s.RevertMenu.Render(width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", descBox, "", menuBox)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case historyModeDetail:
		if s.CurrentDetail == nil {
			return ""
		}
		title := s.Theme.SectionTitle.Render("Commit " + s.CurrentDetail.ShortHash)
		subj := s.Theme.Value.Bold(true).Render(s.CurrentDetail.Subject)
		author := s.Theme.Label.Render(fmt.Sprintf("Author: %s <%s>", s.CurrentDetail.AuthorName, s.CurrentDetail.AuthorEmail))
		date := s.Theme.Label.Render(fmt.Sprintf("Date:   %s (%s)", s.CurrentDetail.Date, s.CurrentDetail.RelativeDate))

		var filesList []string
		filesList = append(filesList, s.Theme.Label.Render("Files:"))
		for i, f := range s.CurrentDetail.Files {
			if i >= 6 {
				filesList = append(filesList, s.Theme.Label.Render(fmt.Sprintf("  ... and %d more files", len(s.CurrentDetail.Files)-6)))
				break
			}
			filesList = append(filesList, fmt.Sprintf("  %s  %s", s.Theme.Title.Render(f.Status), s.Theme.Value.Render(f.Path)))
		}
		filesBox := strings.Join(filesList, "\n")
		menuBox := s.DetailMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back to List"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", subj, "", author, date, "", filesBox, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	default: // historyModeList
		title := s.Theme.SectionTitle.Render("History")
		if len(s.Commits) == 0 {
			emptyMsg := s.Theme.Label.Render("  (No commits yet in this repository)")
			footer := s.Footer.Render([]components.KeyHelp{{Key: "Esc/q", Label: "Back"}}, width)
			content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", emptyMsg, "", footer)
			return s.Theme.Card.Width(boxWidth).Render(content)
		}

		menuBox := s.HistoryMenu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "View Details"},
			{Key: "Esc/q", Label: "Back"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}
}
