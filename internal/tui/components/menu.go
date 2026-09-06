package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gito/internal/theme"
)

// MenuItem represents an interactive option in a menu.
type MenuItem struct {
	ID       string
	Title    string
	Subtitle string
	Icon     string
	Badge    string
}

// Menu manages an interactive selectable list.
type Menu struct {
	Theme       *theme.Theme
	Items       []MenuItem
	Cursor      int
	MaxHeight   int
	OffsetStart int
}

// NewMenu creates a new Menu instance.
func NewMenu(theme *theme.Theme, items []MenuItem) *Menu {
	return &Menu{
		Theme:     theme,
		Items:     items,
		Cursor:    0,
		MaxHeight: 10,
	}
}

// MoveUp moves cursor up.
func (m *Menu) MoveUp() {
	if len(m.Items) == 0 {
		return
	}
	m.Cursor--
	if m.Cursor < 0 {
		m.Cursor = len(m.Items) - 1
	}
	m.adjustOffset()
}

// MoveDown moves cursor down.
func (m *Menu) MoveDown() {
	if len(m.Items) == 0 {
		return
	}
	m.Cursor++
	if m.Cursor >= len(m.Items) {
		m.Cursor = 0
	}
	m.adjustOffset()
}

func (m *Menu) adjustOffset() {
	if m.MaxHeight <= 0 || len(m.Items) <= m.MaxHeight {
		m.OffsetStart = 0
		return
	}
	if m.Cursor < m.OffsetStart {
		m.OffsetStart = m.Cursor
	} else if m.Cursor >= m.OffsetStart+m.MaxHeight {
		m.OffsetStart = m.Cursor - m.MaxHeight + 1
	}
}

// Selected returns the currently highlighted item.
func (m *Menu) Selected() *MenuItem {
	if len(m.Items) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Items) {
		return nil
	}
	return &m.Items[m.Cursor]
}

// SetItems updates items and resets/bounds cursor.
func (m *Menu) SetItems(items []MenuItem) {
	m.Items = items
	if m.Cursor >= len(items) {
		m.Cursor = len(items) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	m.adjustOffset()
}

// Render renders the menu items with cursor pointer.
func (m *Menu) Render(width int) string {
	if len(m.Items) == 0 {
		return m.Theme.Label.Render("  (No items)")
	}

	var lines []string
	start := m.OffsetStart
	end := start + m.MaxHeight
	if end > len(m.Items) || m.MaxHeight <= 0 {
		end = len(m.Items)
	}

	for i := start; i < end; i++ {
		item := m.Items[i]
		isSelected := (i == m.Cursor)

		var cursorStr string
		if isSelected {
			cursorStr = m.Theme.MenuCursor.Render("❯  ")
		} else {
			cursorStr = "   "
		}

		iconStr := ""
		if item.Icon != "" {
			if isSelected {
				iconStr = m.Theme.MenuIcon.Render(item.Icon) + " "
			} else {
				iconStr = m.Theme.Label.Render(item.Icon) + " "
			}
		}

		var titleStr string
		if isSelected {
			titleStr = m.Theme.MenuSelected.Render(item.Title)
		} else {
			titleStr = m.Theme.MenuItem.Render(item.Title)
		}

		lineContent := cursorStr + iconStr + titleStr

		if item.Badge != "" {
			badgeStr := "  " + m.Theme.Label.Render(item.Badge)
			lineContent += badgeStr
		}

		lines = append(lines, lineContent)

		if item.Subtitle != "" {
			subStr := "     " + m.Theme.Label.Render(item.Subtitle)
			lines = append(lines, subStr)
		}
	}

	if len(m.Items) > m.MaxHeight && m.MaxHeight > 0 {
		indicator := m.Theme.Label.Render(lipgloss.NewStyle().Faint(true).Render("  (scroll for more)"))
		lines = append(lines, indicator)
	}

	return strings.Join(lines, "\n")
}
