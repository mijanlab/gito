package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gito/internal/theme"
)

// DiffView renders a syntax-colored, scrollable git diff.
type DiffView struct {
	Theme        *theme.Theme
	Lines        []string
	ScrollOffset int
	MaxHeight    int
}

// NewDiffView creates a new DiffView.
func NewDiffView(theme *theme.Theme, diffText string, maxHeight int) *DiffView {
	lines := strings.Split(diffText, "\n")
	if maxHeight <= 0 {
		maxHeight = 20
	}
	return &DiffView{
		Theme:        theme,
		Lines:        lines,
		ScrollOffset: 0,
		MaxHeight:    maxHeight,
	}
}

// SetDiff updates the diff content.
func (d *DiffView) SetDiff(diffText string) {
	d.Lines = strings.Split(diffText, "\n")
	d.ScrollOffset = 0
}

// ScrollUp moves scroll up.
func (d *DiffView) ScrollUp() {
	if d.ScrollOffset > 0 {
		d.ScrollOffset--
	}
}

// ScrollDown moves scroll down.
func (d *DiffView) ScrollDown() {
	maxScroll := len(d.Lines) - d.MaxHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.ScrollOffset < maxScroll {
		d.ScrollOffset++
	}
}

// PageUp moves up by page height.
func (d *DiffView) PageUp() {
	d.ScrollOffset -= d.MaxHeight
	if d.ScrollOffset < 0 {
		d.ScrollOffset = 0
	}
}

// PageDown moves down by page height.
func (d *DiffView) PageDown() {
	maxScroll := len(d.Lines) - d.MaxHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	d.ScrollOffset += d.MaxHeight
	if d.ScrollOffset > maxScroll {
		d.ScrollOffset = maxScroll
	}
}

// Render renders the visible window of diff lines with syntax coloring.
func (d *DiffView) Render(width int) string {
	if len(d.Lines) == 0 || (len(d.Lines) == 1 && d.Lines[0] == "") {
		return d.Theme.Label.Render("  (No diff to display)")
	}

	start := d.ScrollOffset
	end := start + d.MaxHeight
	if end > len(d.Lines) {
		end = len(d.Lines)
	}

	var renderedLines []string
	for i := start; i < end; i++ {
		line := d.Lines[i]
		colored := d.colorLine(line)
		renderedLines = append(renderedLines, colored)
	}

	// Status line at bottom: [Line 1-20 of 142]
	pos := fmt.Sprintf("[%d-%d / %d lines]", start+1, end, len(d.Lines))
	statusLine := d.Theme.Label.Render(pos)

	content := strings.Join(renderedLines, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, content, "", statusLine)
}

func (d *DiffView) colorLine(line string) string {
	switch {
	case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index "):
		return d.Theme.DiffMeta.Render(line)
	case strings.HasPrefix(line, "+"):
		return d.Theme.DiffAdded.Render(line)
	case strings.HasPrefix(line, "-"):
		return d.Theme.DiffRemoved.Render(line)
	case strings.HasPrefix(line, "@@"):
		return d.Theme.DiffHunk.Render(line)
	default:
		return d.Theme.Value.Render(line)
	}
}
