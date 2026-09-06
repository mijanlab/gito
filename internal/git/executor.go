package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitError wraps error messages from the git command with human-friendly descriptions.
type GitError struct {
	Command string
	Args    []string
	Stderr  string
	Err     error
}

func (e *GitError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("git %s failed: %s", strings.Join(e.Args, " "), strings.TrimSpace(e.Stderr))
	}
	return fmt.Sprintf("git %s failed: %v", strings.Join(e.Args, " "), e.Err)
}

// UserFriendlyError returns a helpful human-readable explanation of common git errors.
func (e *GitError) UserFriendlyError() (title string, explanation string) {
	stderr := strings.ToLower(e.Stderr)

	switch {
	case strings.Contains(stderr, "no upstream branch") || strings.Contains(stderr, "has no tracking branch"):
		return "No Upstream Branch", "This branch is not connected to a remote tracking branch. Use 'Push & set upstream' to connect it."
	case strings.Contains(stderr, "could not resolve host") || strings.Contains(stderr, "failed to connect") || strings.Contains(stderr, "network is unreachable"):
		return "Network Error", "Unable to connect to the remote repository. Check your internet connection and remote URL."
	case strings.Contains(stderr, "permission denied") || strings.Contains(stderr, "authentication failed"):
		return "Authentication Failed", "Git was unable to authenticate with the remote repository. Check your SSH keys or credentials."
	case strings.Contains(stderr, "non-fast-forward") || strings.Contains(stderr, "fetch first") || strings.Contains(stderr, "updates were rejected"):
		return "Remote Has Newer Changes", "The remote repository contains commits you don't have locally. Pull before pushing, or resolve diverged history."
	case strings.Contains(stderr, "local changes to the following files would be overwritten"):
		return "Uncommitted Local Changes", "You have local changes that would be overwritten. Commit or stash them before proceeding."
	case strings.Contains(stderr, "conflict") || strings.Contains(stderr, "merge conflict"):
		return "Merge Conflict", "There are conflicting changes between branches that need to be resolved."
	case strings.Contains(stderr, "not a git repository"):
		return "No Git Repository", "The current directory is not part of a Git repository."
	case strings.Contains(stderr, "branch") && strings.Contains(stderr, "not fully merged"):
		return "Unmerged Branch", "This branch contains commits that have not been merged into your current branch."
	case strings.Contains(stderr, "already exists"):
		return "Resource Already Exists", "A branch or reference with this name already exists."
	default:
		if e.Stderr != "" {
			return "Git Operation Failed", strings.TrimSpace(e.Stderr)
		}
		if e.Err != nil {
			return "Git Operation Failed", e.Err.Error()
		}
		return "Git Operation Failed", "An unknown git error occurred."
	}
}

// Executor executes git commands safely without invoking shell strings.
type Executor struct {
	GitPath string
}

// NewExecutor creates a new git executor.
func NewExecutor() (*Executor, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("git executable not found in PATH. Please install Git to use Gito.")
	}
	return &Executor{GitPath: gitPath}, nil
}

// Run executes a git command in the given directory and returns stdout.
func (e *Executor) Run(dir string, args ...string) (string, error) {
	return e.RunWithContext(context.Background(), dir, args...)
}

// RunWithContext executes a git command with context timeout / cancellation.
func (e *Executor) RunWithContext(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, e.GitPath, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", &GitError{
			Command: e.GitPath,
			Args:    args,
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	return stdout.String(), nil
}

// Version returns the installed Git version string.
func (e *Executor) Version() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := e.RunWithContext(ctx, "", "version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
