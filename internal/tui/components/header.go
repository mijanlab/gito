package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gito/internal/git"
	"gito/internal/theme"
)

// Header renders the top application title bar and repository context.
type Header struct {
	Theme            *theme.Theme
	Version          string
	AvailableVersion string
}

// NewHeader creates a new Header component.
func NewHeader(theme *theme.Theme, version string) *Header {
	return &Header{
		Theme:   theme,
		Version: version,
	}
}

// SetAvailableVersion updates the available version string.
func (h *Header) SetAvailableVersion(ver string) {
	h.AvailableVersion = ver
}

// Render renders the header for the given repository info and status.
func (h *Header) Render(repo *git.RepoInfo, status *git.StatusInfo, width int) string {
	if width < 40 {
		width = 40
	}
	innerWidth := width - 4

	// Top bar: ◈ Gito ... v0.1.0 [Available: v0.2.0]
	brand := h.Theme.Title.Render("◈ Gito")
	verStr := "v" + h.Version
	if h.AvailableVersion != "" && h.AvailableVersion != h.Version {
		verStr += " " + h.Theme.AIText.Render("• Available: v"+strings.TrimPrefix(h.AvailableVersion, "v"))
	}
	ver := h.Theme.Version.Render(verStr)

	spaceCount := innerWidth - lipgloss.Width(brand) - lipgloss.Width(ver)
	if spaceCount < 1 {
		spaceCount = 1
	}
	topBar := brand + strings.Repeat(" ", spaceCount) + ver

	if repo == nil || !repo.IsInside {
		return topBar
	}

	// Repo & Branch info line
	repoName := h.Theme.RepoName.Render(repo.Name)
	branchName := "detached"
	if status != nil {
		if status.IsDetached {
			branchName = fmt.Sprintf("detached (%s)", status.DetachedOID)
		} else if status.Branch != "" {
			branchName = status.Branch
		}
	}
	branchLine := h.Theme.Branch.Render("└─ " + branchName)

	var statusBadge string
	if status != nil {
		switch {
		case len(status.Conflicts) > 0:
			statusBadge = h.Theme.StatusError.Render(fmt.Sprintf("! %d conflict(s)", len(status.Conflicts)))
		case status.State != git.StateClean && status.State != git.StateModified:
			statusBadge = h.Theme.StatusWarning.Render(fmt.Sprintf("⚠ %s", status.State))
		case status.IsClean:
			statusBadge = h.Theme.StatusClean.Render("● Clean")
		default:
			count := len(status.Files)
			statusBadge = h.Theme.StatusModified.Render(fmt.Sprintf("● %d modified", count))
		}

		if status.Ahead > 0 || status.Behind > 0 {
			var syncParts []string
			if status.Ahead > 0 {
				syncParts = append(syncParts, fmt.Sprintf("↑ %d", status.Ahead))
			}
			if status.Behind > 0 {
				syncParts = append(syncParts, fmt.Sprintf("↓ %d", status.Behind))
			}
			statusBadge += "  " + h.Theme.Label.Render(strings.Join(syncParts, " "))
		}
	}

	spaceCount2 := innerWidth - lipgloss.Width(branchLine) - lipgloss.Width(statusBadge)
	if spaceCount2 < 1 {
		spaceCount2 = 1
	}
	branchAndStatus := branchLine + strings.Repeat(" ", spaceCount2) + statusBadge

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		"",
		repoName,
		branchAndStatus,
	)
}
