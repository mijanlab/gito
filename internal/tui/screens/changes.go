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

type changesViewMode int

const (
	changesModeSummary changesViewMode = iota
	changesModeDiffView
)

// ChangesScreen displays changed files and diff viewer.
type ChangesScreen struct {
	Theme       *theme.Theme
	Header      *components.Header
	Footer      *components.Footer
	ActionMenu  *components.Menu
	FileMenu    *components.Menu
	DiffViewer  *components.DiffView
	Mode        changesViewMode
	Status      *git.StatusInfo
	Stats       git.DiffStats
	CurrentDiff string
}

// NewChangesScreen creates a new ChangesScreen.
func NewChangesScreen(theme *theme.Theme, version string) *ChangesScreen {
	return &ChangesScreen{
		Theme:      theme,
		Header:     components.NewHeader(theme, version),
		Footer:     components.NewFooter(theme),
		DiffViewer: components.NewDiffView(theme, "", 18),
		Mode:       changesModeSummary,
	}
}

// Reset initializes the screen with status, stats, and diff text.
func (s *ChangesScreen) Reset(status *git.StatusInfo, stats git.DiffStats, fullDiff string) {
	s.Status = status
	s.Stats = stats
	s.CurrentDiff = fullDiff
	s.DiffViewer.SetDiff(fullDiff)
	s.Mode = changesModeSummary

	s.ActionMenu = components.NewMenu(s.Theme, []components.MenuItem{
		{ID: "view_diff", Title: "View full diff", Icon: "◌"},
		{ID: "commit", Title: "Commit changes", Icon: "◇"},
		{ID: "back", Title: "Back", Icon: "←"},
	})

	var fItems []components.MenuItem
	if status != nil {
		for _, f := range status.Files {
			statusBadge := f.DisplayStatus()
			fItems = append(fItems, components.MenuItem{
				ID:    f.Path,
				Title: f.Path,
				Badge: statusBadge,
				Icon:  "•",
			})
		}
	}
	s.FileMenu = components.NewMenu(s.Theme, fItems)
}

// SetFileDiff sets a specific file's diff and opens the diff viewer.
func (s *ChangesScreen) SetFileDiff(diff string) {
	s.CurrentDiff = diff
	s.DiffViewer.SetDiff(diff)
	s.Mode = changesModeDiffView
}

// Update handles keyboard navigation for ChangesScreen.
func (s *ChangesScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case changesModeSummary:
			switch msg.String() {
			case "up", "k":
				s.ActionMenu.MoveUp()
			case "down", "j":
				s.ActionMenu.MoveDown()
			case "enter":
				selected := s.ActionMenu.Selected()
				if selected != nil {
					switch selected.ID {
					case "view_diff":
						s.DiffViewer.SetDiff(s.CurrentDiff)
						s.Mode = changesModeDiffView
						return "", nil
					case "commit":
						return "commit", nil
					case "back":
						return "main_menu", nil
					}
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case changesModeDiffView:
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
				s.Mode = changesModeSummary
				return "", nil
			}
		}
	}
	return "", nil
}

// Render renders the changes screen.
func (s *ChangesScreen) Render(repo *git.RepoInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, s.Status, width)

	if s.Mode == changesModeDiffView {
		title := s.Theme.SectionTitle.Render("Diff View")
		diffBox := s.DiffViewer.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓/PgUp/PgDn", Label: "Scroll"},
			{Key: "Esc/q", Label: "Back to Changes"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", diffBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}

	title := s.Theme.SectionTitle.Render("Changes")

	if s.Status == nil || (len(s.Status.Files) == 0 && len(s.Status.Conflicts) == 0) {
		cleanBox := s.Theme.SuccessText.Render("Working tree clean. No local modifications.")
		footer := s.Footer.Render([]components.KeyHelp{{Key: "Esc/Enter", Label: "Back"}}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", cleanBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}

	// Categorize files
	var modifiedLines, addedLines, deletedLines, conflictLines []string
	for _, f := range s.Status.Files {
		switch {
		case f.IsConflict:
			conflictLines = append(conflictLines, "  "+s.Theme.ErrorText.Render(f.Path))
		case f.IsUntracked || f.StagedCode == 'A':
			addedLines = append(addedLines, "  "+s.Theme.SuccessText.Render(f.Path))
		case f.IsDeleted:
			deletedLines = append(deletedLines, "  "+s.Theme.ErrorText.Render(f.Path))
		default:
			modifiedLines = append(modifiedLines, "  "+s.Theme.WarningText.Render(f.Path))
		}
	}

	var sections []string
	if len(conflictLines) > 0 {
		sections = append(sections, s.Theme.ErrorText.Bold(true).Render("Conflicts:"), strings.Join(conflictLines, "\n"), "")
	}
	if len(modifiedLines) > 0 {
		sections = append(sections, s.Theme.WarningText.Bold(true).Render("Modified:"), strings.Join(modifiedLines, "\n"), "")
	}
	if len(addedLines) > 0 {
		sections = append(sections, s.Theme.SuccessText.Bold(true).Render("Added / Untracked:"), strings.Join(addedLines, "\n"), "")
	}
	if len(deletedLines) > 0 {
		sections = append(sections, s.Theme.ErrorText.Bold(true).Render("Deleted:"), strings.Join(deletedLines, "\n"), "")
	}

	// Summary stats
	statsStr := fmt.Sprintf("%d %s changed", len(s.Status.Files), plural(len(s.Status.Files), "file", "files"))
	if s.Stats.Additions > 0 || s.Stats.Deletions > 0 {
		statsStr += fmt.Sprintf(" • +%d -%d", s.Stats.Additions, s.Stats.Deletions)
	}
	statsBox := s.Theme.Label.Render(statsStr)
	sections = append(sections, statsBox, "")

	menuBox := s.ActionMenu.Render(width)
	sections = append(sections, menuBox)

	footer := s.Footer.Render([]components.KeyHelp{
		{Key: "↑↓", Label: "Navigate"},
		{Key: "Enter", Label: "Select"},
		{Key: "Esc", Label: "Back"},
	}, width)

	content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", strings.Join(sections, "\n"), "", footer)
	return s.Theme.Card.Width(boxWidth).Render(content)
}
