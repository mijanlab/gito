package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gito/internal/git"
	"gito/internal/theme"
)

func TestMainMenuNavigation(t *testing.T) {
	th := theme.DefaultTheme()
	screen := NewMainMenuScreen(th, "0.0.5")

	status := &git.StatusInfo{
		Branch:  "main",
		Ahead:   2,
		Behind:  0,
		IsClean: true,
		State:   git.StateClean,
	}
	screen.InitMenu(status)

	// Test cursor moving down
	action, _ := screen.Update(tea.KeyMsg{Type: tea.KeyDown})
	if action != "" {
		t.Errorf("expected empty action on KeyDown, got '%s'", action)
	}
	if screen.Menu.Cursor != 1 {
		t.Errorf("expected cursor to be at index 1, got %d", screen.Menu.Cursor)
	}

	// Test pressing Enter on second item (pull)
	action, _ = screen.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action != "pull" {
		t.Errorf("expected action 'pull', got '%s'", action)
	}
}

func TestCommitScreenTransitions(t *testing.T) {
	th := theme.DefaultTheme()
	screen := NewCommitScreen(th, "0.0.5")

	status := &git.StatusInfo{
		Branch:  "main",
		IsClean: false,
		Files: []git.FileStatus{
			{Path: "main.go", WorkingCode: 'M'},
		},
	}
	screen.Reset(status, "diff --git a/main.go ...", true)

	// Selecting "Write myself"
	screen.MethodMenu.Cursor = 1
	action, _ := screen.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action != "" {
		t.Errorf("expected empty action, got '%s'", action)
	}
	if screen.Mode != commitModeInputMessage {
		t.Errorf("expected mode commitModeInputMessage, got %v", screen.Mode)
	}

	// Type commit message
	screen.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("F")})
	screen.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	screen.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if screen.CommitMessage != "Fix" {
		t.Errorf("expected CommitMessage 'Fix', got '%s'", screen.CommitMessage)
	}

	// Submit input message
	screen.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if screen.Mode != commitModeConfirm {
		t.Errorf("expected mode commitModeConfirm, got %v", screen.Mode)
	}
}

func TestPullSafetyDecision(t *testing.T) {
	th := theme.DefaultTheme()
	screen := NewPullScreen(th, "0.0.5")

	status := &git.StatusInfo{
		Branch: "main",
	}
	branches := []git.Branch{
		{Name: "main", IsCurrent: true},
		{Name: "development"},
	}

	screen.Reset(status, branches)

	// Choose Another branch
	screen.Menu.Cursor = 1
	screen.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if screen.Mode != pullModeSelectBranch {
		t.Fatalf("expected mode pullModeSelectBranch, got %v", screen.Mode)
	}

	// Select 'development' branch (index 1)
	screen.BranchMenu.Cursor = 1
	screen.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Must trigger safety decision rather than blindly pulling
	if screen.Mode != pullModeSafetyDecision {
		t.Fatalf("expected safety decision mode when selecting a different branch, got %v", screen.Mode)
	}

	// Select Switch and pull
	action, _ := screen.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action != "do_pull:switch_and_pull:development" {
		t.Errorf("expected 'do_pull:switch_and_pull:development', got '%s'", action)
	}
}
