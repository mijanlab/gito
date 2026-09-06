package cli

import (
	"os"
	"path/filepath"
	"testing"

	"gito/internal/git"
)

func TestRunnerPushAndPull(t *testing.T) {
	exec, err := git.NewExecutor()
	if err != nil {
		t.Fatalf("failed to create git executor: %v", err)
	}

	// 1. Create a remote bare repo
	remoteDir := t.TempDir()
	_, err = exec.Run(remoteDir, "init", "--bare")
	if err != nil {
		t.Fatalf("failed to init bare remote: %v", err)
	}

	// 2. Create a local repo
	localDir := t.TempDir()
	_ = exec.InitRepository(localDir)
	_, _ = exec.Run(localDir, "config", "user.name", "Tester")
	_, _ = exec.Run(localDir, "config", "user.email", "tester@gito.dev")
	_, _ = exec.Run(localDir, "remote", "add", "origin", remoteDir)

	// Create initial file
	testFile := filepath.Join(localDir, "hello.txt")
	_ = os.WriteFile(testFile, []byte("hello world\n"), 0644)

	runner, err := NewRunner(localDir)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Test Status
	err = runner.Status()
	if err != nil {
		t.Fatalf("runner.Status failed: %v", err)
	}

	// Test Push with autoAccept
	err = runner.Push(true)
	if err != nil {
		t.Fatalf("runner.Push failed: %v", err)
	}

	// Test Pull
	err = runner.Pull()
	if err != nil {
		t.Fatalf("runner.Pull failed: %v", err)
	}
}
