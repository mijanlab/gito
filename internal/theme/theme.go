package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the visual styling and color palette of Gito.
type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Muted     lipgloss.Color
	Accent    lipgloss.Color
	AI        lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Danger    lipgloss.Color
	Border    lipgloss.Color
	BgSubtle  lipgloss.Color

	// Pre-built styles
	Title          lipgloss.Style
	Version        lipgloss.Style
	RepoName       lipgloss.Style
	Branch         lipgloss.Style
	StatusClean    lipgloss.Style
	StatusModified lipgloss.Style
	StatusWarning  lipgloss.Style
	StatusError    lipgloss.Style

	MenuItem     lipgloss.Style
	MenuSelected lipgloss.Style
	MenuIcon     lipgloss.Style
	MenuCursor   lipgloss.Style

	Card         lipgloss.Style
	SectionTitle lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	HelpBar      lipgloss.Style
	KeyBadge     lipgloss.Style

	SuccessText lipgloss.Style
	WarningText lipgloss.Style
	ErrorText   lipgloss.Style
	AIText      lipgloss.Style

	DiffAdded   lipgloss.Style
	DiffRemoved lipgloss.Style
	DiffHunk    lipgloss.Style
	DiffMeta    lipgloss.Style
}

// DefaultTheme returns the premium dark terminal palette.
func DefaultTheme() *Theme {
	t := &Theme{
		Primary:   lipgloss.Color("#F8FAFC"), // Crisp white
		Secondary: lipgloss.Color("#94A3B8"), // Cool slate
		Muted:     lipgloss.Color("#64748B"), // Muted slate
		Accent:    lipgloss.Color("#38BDF8"), // Sky blue accent
		AI:        lipgloss.Color("#C084FC"), // Soft purple
		Success:   lipgloss.Color("#34D399"), // Emerald green
		Warning:   lipgloss.Color("#FBBF24"), // Warm amber
		Danger:    lipgloss.Color("#F87171"), // Soft red
		Border:    lipgloss.Color("#334155"), // Dark slate border
		BgSubtle:  lipgloss.Color("#1E293B"), // Subtle dark background
	}

	t.Title = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	t.Version = lipgloss.NewStyle().
		Foreground(t.Muted)

	t.RepoName = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	t.Branch = lipgloss.NewStyle().
		Foreground(t.Secondary)

	t.StatusClean = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)

	t.StatusModified = lipgloss.NewStyle().
		Foreground(t.Warning)

	t.StatusWarning = lipgloss.NewStyle().
		Foreground(t.Warning).
		Bold(true)

	t.StatusError = lipgloss.NewStyle().
		Foreground(t.Danger).
		Bold(true)

	t.MenuItem = lipgloss.NewStyle().
		Foreground(t.Secondary).
		PaddingLeft(2)

	t.MenuSelected = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	t.MenuCursor = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	t.MenuIcon = lipgloss.NewStyle().
		Foreground(t.Accent).
		MarginRight(1)

	t.Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2)

	t.SectionTitle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		MarginBottom(1)

	t.Label = lipgloss.NewStyle().
		Foreground(t.Muted)

	t.Value = lipgloss.NewStyle().
		Foreground(t.Primary)

	t.HelpBar = lipgloss.NewStyle().
		Foreground(t.Muted).
		MarginTop(1)

	t.KeyBadge = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	t.SuccessText = lipgloss.NewStyle().
		Foreground(t.Success)

	t.WarningText = lipgloss.NewStyle().
		Foreground(t.Warning)

	t.ErrorText = lipgloss.NewStyle().
		Foreground(t.Danger)

	t.AIText = lipgloss.NewStyle().
		Foreground(t.AI).
		Bold(true)

	t.DiffAdded = lipgloss.NewStyle().
		Foreground(t.Success)

	t.DiffRemoved = lipgloss.NewStyle().
		Foreground(t.Danger)

	t.DiffHunk = lipgloss.NewStyle().
		Foreground(t.Accent)

	t.DiffMeta = lipgloss.NewStyle().
		Foreground(t.Muted)

	return t
}
