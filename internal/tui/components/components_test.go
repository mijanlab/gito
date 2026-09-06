package components

import (
	"strings"
	"testing"

	"gito/internal/git"
	"gito/internal/theme"
)

func TestHeaderRendering(t *testing.T) {
	th := theme.DefaultTheme()
	header := NewHeader(th, "0.0.5")

	repo := &git.RepoInfo{
		Path:     "/home/user/project",
		Name:     "my-website",
		IsInside: true,
	}

	status := &git.StatusInfo{
		Branch:  "main",
		IsClean: true,
		Ahead:   1,
	}

	out := header.Render(repo, status, 80)
	if !strings.Contains(out, "Gito") {
		t.Errorf("expected header to contain 'Gito'")
	}
	if !strings.Contains(out, "my-website") {
		t.Errorf("expected header to contain repo name 'my-website'")
	}
	if !strings.Contains(out, "main") {
		t.Errorf("expected header to contain branch 'main'")
	}
}

func TestMenuNavigation(t *testing.T) {
	th := theme.DefaultTheme()
	items := []MenuItem{
		{ID: "1", Title: "First"},
		{ID: "2", Title: "Second"},
		{ID: "3", Title: "Third"},
	}
	menu := NewMenu(th, items)

	if menu.Selected().ID != "1" {
		t.Errorf("expected initial selection '1'")
	}

	menu.MoveDown()
	if menu.Selected().ID != "2" {
		t.Errorf("expected selection '2' after MoveDown")
	}

	menu.MoveUp()
	if menu.Selected().ID != "1" {
		t.Errorf("expected selection '1' after MoveUp")
	}

	// Wrap around test
	menu.MoveUp()
	if menu.Selected().ID != "3" {
		t.Errorf("expected wrap around to '3'")
	}
}

func TestDiffViewScrolling(t *testing.T) {
	th := theme.DefaultTheme()
	diff := "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8"
	dv := NewDiffView(th, diff, 3)

	if dv.ScrollOffset != 0 {
		t.Errorf("expected initial ScrollOffset 0")
	}

	dv.ScrollDown()
	if dv.ScrollOffset != 1 {
		t.Errorf("expected ScrollOffset 1")
	}

	dv.PageDown()
	if dv.ScrollOffset != 4 {
		t.Errorf("expected ScrollOffset 4 after PageDown")
	}

	dv.PageUp()
	if dv.ScrollOffset != 1 {
		t.Errorf("expected ScrollOffset 1 after PageUp")
	}
}
