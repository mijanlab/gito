package tui

import (
	"os"
	"path/filepath"
	"testing"

	"gito/internal/git"
)

func TestAppInitAndState(t *testing.T) {
	exec, err := git.NewExecutor()
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	dir := t.TempDir()
	_ = exec.InitRepository(dir)
	_, _ = exec.Run(dir, "config", "user.name", "Tester")
	_, _ = exec.Run(dir, "config", "user.email", "test@example.com")
	_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
	_, _ = exec.Commit(dir, "Initial commit", true)

	app, err := NewApp("0.0.5", dir)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	cmd := app.checkAndRefreshCmd()
	if cmd == nil {
		t.Fatalf("expected non-nil checkAndRefreshCmd command")
	}

	msg := cmd()
	app.Update(msg)

	if app.RepoInfo == nil || !app.RepoInfo.IsInside {
		t.Errorf("expected app to recognize repo")
	}
	if app.ActiveScreen != ScreenMainMenu {
		t.Errorf("expected ActiveScreen ScreenMainMenu, got %v", app.ActiveScreen)
	}

	// Test non-git directory switches to ScreenNotGit
	nonRepo := t.TempDir()
	appNon, _ := NewApp("0.0.5", nonRepo)
	cmdNon := appNon.checkAndRefreshCmd()
	msgNon := cmdNon()
	appNon.Update(msgNon)

	if appNon.ActiveScreen != ScreenNotGit {
		t.Errorf("expected ActiveScreen ScreenNotGit for empty dir, got %v", appNon.ActiveScreen)
	}
}
