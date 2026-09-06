package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gito/internal/theme"
)

// KeyHelp represents a shortcut description.
type KeyHelp struct {
	Key   string
	Label string
}

// Footer renders responsive keyboard navigation hints at the bottom.
type Footer struct {
	Theme *theme.Theme
}

// NewFooter creates a new Footer component.
func NewFooter(theme *theme.Theme) *Footer {
	return &Footer{Theme: theme}
}

// Render renders the keyboard shortcuts horizontally.
func (f *Footer) Render(keys []KeyHelp, width int) string {
	if len(keys) == 0 {
		keys = []KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
			{Key: "q", Label: "Quit"},
		}
	}

	var items []string
	for _, k := range keys {
		item := f.Theme.KeyBadge.Render(k.Key) + " " + f.Theme.Label.Render(k.Label)
		items = append(items, item)
	}

	line := strings.Join(items, "   ")
	if width > 0 && lipgloss.Width(line) > width-4 {
		// More compact separator for narrow widths
		line = strings.Join(items, "  ")
	}

	return f.Theme.HelpBar.Render(line)
}
