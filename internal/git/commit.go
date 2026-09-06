package git

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CommitResult holds information about a newly created commit.
type CommitResult struct {
	Hash    string
	Subject string
}

// StageAll stages all current changes in the repository respecting .gitignore.
func (e *Executor) StageAll(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := e.RunWithContext(ctx, dir, "add", "-A")
	return err
}

// StageFiles stages a list of specific files.
func (e *Executor) StageFiles(dir string, files ...string) error {
	if len(files) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := append([]string{"add", "--"}, files...)
	_, err := e.RunWithContext(ctx, dir, args...)
	return err
}

// Commit creates a git commit with the given message.
// If stageAll is true, it stages all changes first.
func (e *Executor) Commit(dir, message string, stageAll bool) (*CommitResult, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("commit message cannot be empty")
	}

	if stageAll {
		if err := e.StageAll(dir); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := e.RunWithContext(ctx, dir, "commit", "-m", message)
	if err != nil {
		return nil, err
	}

	// Retrieve short commit hash
	hashCtx, hashCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer hashCancel()

	hashOut, err := e.RunWithContext(hashCtx, dir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return &CommitResult{Subject: message}, nil
	}

	return &CommitResult{
		Hash:    strings.TrimSpace(hashOut),
		Subject: message,
	}, nil
}
