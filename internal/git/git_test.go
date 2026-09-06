package git

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (*Executor, string) {
	t.Helper()
	exec, err := NewExecutor()
	if err != nil {
		t.Fatalf("failed to create git executor: %v", err)
	}

	dir := t.TempDir()
	if err := exec.InitRepository(dir); err != nil {
		t.Fatalf("failed to init test repo: %v", err)
	}

	// Configure git user inside test repo for committing
	_, _ = exec.Run(dir, "config", "user.name", "Gito Tester")
	_, _ = exec.Run(dir, "config", "user.email", "tester@gito.dev")

	return exec, dir
}

func TestDetectRepository(t *testing.T) {
	exec, dir := setupTestRepo(t)

	info, err := exec.DetectRepository(dir)
	if err != nil {
		t.Fatalf("DetectRepository failed: %v", err)
	}
	if !info.IsInside {
		t.Errorf("expected IsInside to be true")
	}

	// Test non-repo directory
	nonRepoDir := filepath.Join(t.TempDir(), "empty")
	_ = os.MkdirAll(nonRepoDir, 0755)
	infoNon, err := exec.DetectRepository(nonRepoDir)
	if err != nil {
		t.Fatalf("DetectRepository on non-repo returned error: %v", err)
	}
	if infoNon.IsInside {
		t.Errorf("expected IsInside to be false for non-repo")
	}
}

func TestStatusAndCommitFlow(t *testing.T) {
	exec, dir := setupTestRepo(t)

	// Initial clean status
	status, err := exec.GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if !status.IsClean {
		t.Errorf("expected initial repo to be clean")
	}

	// Create a new file
	testFile := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Status should now show untracked / modified
	status, err = exec.GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status.IsClean {
		t.Errorf("expected repo to not be clean after creating file")
	}
	if len(status.Untracked) != 1 || status.Untracked[0].Path != "hello.txt" {
		t.Errorf("expected 1 untracked file 'hello.txt', got: %+v", status.Untracked)
	}

	// Commit the file
	res, err := exec.Commit(dir, "Initial commit", true)
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
	if res.Subject != "Initial commit" {
		t.Errorf("expected commit subject 'Initial commit', got '%s'", res.Subject)
	}
	if res.Hash == "" {
		t.Errorf("expected non-empty commit hash")
	}

	// Status should now be clean
	status, err = exec.GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if !status.IsClean {
		t.Errorf("expected repo to be clean after commit")
	}
}

func TestBranchOperations(t *testing.T) {
	exec, dir := setupTestRepo(t)

	// Create initial commit on default branch
	testFile := filepath.Join(dir, "README.md")
	_ = os.WriteFile(testFile, []byte("# Test Repo\n"), 0644)
	_, err := exec.Commit(dir, "Add readme", true)
	if err != nil {
		t.Fatalf("initial commit failed: %v", err)
	}

	// Create new branch and switch
	err = exec.CreateBranch(dir, "feature/payment", true)
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	status, err := exec.GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status.Branch != "feature/payment" {
		t.Errorf("expected current branch 'feature/payment', got '%s'", status.Branch)
	}

	// List branches
	branches, err := exec.ListBranches(dir)
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}
	if len(branches) < 2 {
		t.Fatalf("expected at least 2 branches, got %d", len(branches))
	}

	// Switch back to original branch
	origBranch := branches[0].Name
	if origBranch == "feature/payment" && len(branches) > 1 {
		origBranch = branches[1].Name
	}

	err = exec.SwitchBranch(dir, origBranch)
	if err != nil {
		t.Fatalf("SwitchBranch failed: %v", err)
	}

	status, _ = exec.GetStatus(dir)
	if status.Branch != origBranch {
		t.Errorf("expected current branch '%s', got '%s'", origBranch, status.Branch)
	}

	// Delete branch
	err = exec.DeleteBranch(dir, "feature/payment", false)
	if err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}
}

func TestDiffAndHistory(t *testing.T) {
	exec, dir := setupTestRepo(t)

	// Commit 1
	file1 := filepath.Join(dir, "app.go")
	_ = os.WriteFile(file1, []byte("package main\n"), 0644)
	c1, err := exec.Commit(dir, "feat: initialize app", true)
	if err != nil {
		t.Fatalf("commit 1 failed: %v", err)
	}

	// Commit 2
	_ = os.WriteFile(file1, []byte("package main\n\nfunc main() {}\n"), 0644)
	c2, err := exec.Commit(dir, "feat: add main entry", true)
	if err != nil {
		t.Fatalf("commit 2 failed: %v", err)
	}

	// Check history
	history, err := exec.GetCommitHistory(dir, 10)
	if err != nil {
		t.Fatalf("GetCommitHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 commits in history, got %d", len(history))
	}
	if history[0].Hash != c2.Hash && history[0].ShortHash != c2.Hash {
		// Verify subject
		if history[0].Subject != "feat: add main entry" {
			t.Errorf("expected top commit 'feat: add main entry', got '%s'", history[0].Subject)
		}
	}

	// Check commit details
	detail, err := exec.GetCommitDetails(dir, history[0].Hash)
	if err != nil {
		t.Fatalf("GetCommitDetails failed: %v", err)
	}
	if detail.Subject != "feat: add main entry" {
		t.Errorf("expected commit detail subject 'feat: add main entry', got '%s'", detail.Subject)
	}

	// Revert commit 2
	err = exec.RevertCommit(dir, history[0].Hash)
	if err != nil {
		t.Fatalf("RevertCommit failed: %v", err)
	}

	historyAfter, _ := exec.GetCommitHistory(dir, 10)
	if len(historyAfter) != 3 {
		t.Errorf("expected 3 commits after revert, got %d", len(historyAfter))
	}
	_ = c1
}

func TestFriendlyErrors(t *testing.T) {
	err := &GitError{
		Stderr: "fatal: The current branch feature/login has no upstream branch.",
	}
	title, exp := err.UserFriendlyError()
	if title != "No Upstream Branch" {
		t.Errorf("expected 'No Upstream Branch', got '%s'", title)
	}
	if exp == "" {
		t.Errorf("expected non-empty explanation")
	}

	conflictErr := &GitError{
		Stderr: "error: you have merge conflicts in file.txt",
	}
	cTitle, _ := conflictErr.UserFriendlyError()
	if cTitle != "Merge Conflict" {
		t.Errorf("expected 'Merge Conflict', got '%s'", cTitle)
	}
}
