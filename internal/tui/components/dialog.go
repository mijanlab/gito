package components

import (
	"strings"

	"gito/internal/theme"
)

// DialogType indicates the severity of the dialog.
type DialogType int

const (
	DialogInfo DialogType = iota
	DialogWarning
	DialogDanger
	DialogSuccess
)

// Dialog renders a clean modal box for confirmations and alerts.
type Dialog struct {
	Theme       *theme.Theme
	Type        DialogType
	Title       string
	Message     string
	Details     string
	Options     []MenuItem
	Cursor      int
	InputPrompt string
	InputValue  string
}

// NewDialog creates a new dialog.
func NewDialog(theme *theme.Theme, dType DialogType, title, message string) *Dialog {
	return &Dialog{
		Theme:   theme,
		Type:    dType,
		Title:   title,
		Message: message,
		Options: []MenuItem{
			{ID: "cancel", Title: "Cancel"},
		},
		Cursor: 0,
	}
}

// MoveUp moves cursor up in dialog options.
func (d *Dialog) MoveUp() {
	if len(d.Options) == 0 {
		return
	}
	d.Cursor--
	if d.Cursor < 0 {
		d.Cursor = len(d.Options) - 1
	}
}

// MoveDown moves cursor down in dialog options.
func (d *Dialog) MoveDown() {
	if len(d.Options) == 0 {
		return
	}
	d.Cursor++
	if d.Cursor >= len(d.Options) {
		d.Cursor = 0
	}
}

// Selected returns selected option id.
func (d *Dialog) Selected() string {
	if len(d.Options) == 0 || d.Cursor < 0 || d.Cursor >= len(d.Options) {
		return ""
	}
	return d.Options[d.Cursor].ID
}

// Render renders the dialog box.
func (d *Dialog) Render(width int) string {
	var titleStyled string
	switch d.Type {
	case DialogDanger:
		titleStyled = d.Theme.ErrorText.Bold(true).Render("⚠ " + d.Title)
	case DialogWarning:
		titleStyled = d.Theme.WarningText.Bold(true).Render("! " + d.Title)
	case DialogSuccess:
		titleStyled = d.Theme.SuccessText.Bold(true).Render("✓ " + d.Title)
	default:
		titleStyled = d.Theme.Title.Render("◈ " + d.Title)
	}

	var sections []string
	sections = append(sections, titleStyled, "")

	if d.Message != "" {
		sections = append(sections, d.Theme.Value.Render(d.Message), "")
	}

	if d.InputPrompt != "" {
		prompt := d.Theme.Label.Render(d.InputPrompt)
		input := d.Theme.Value.Bold(true).Render(d.InputValue + "█")
		sections = append(sections, prompt, input, "")
	}

	if d.Details != "" {
		detailsBox := d.Theme.Label.Render("Details:\n  " + strings.ReplaceAll(d.Details, "\n", "\n  "))
		sections = append(sections, detailsBox, "")
	}

	// Options
	if len(d.Options) > 0 {
		var optLines []string
		for i, opt := range d.Options {
			if i == d.Cursor {
				optLines = append(optLines, d.Theme.MenuCursor.Render("❯ ")+d.Theme.MenuSelected.Render(opt.Title))
			} else {
				optLines = append(optLines, "  "+d.Theme.MenuItem.Render(opt.Title))
			}
		}
		sections = append(sections, strings.Join(optLines, "\n"))
	}

	content := strings.Join(sections, "\n")
	boxWidth := width - 6
	if boxWidth < 40 {
		boxWidth = 40
	}

	return d.Theme.Card.Width(boxWidth).Render(content)
}
