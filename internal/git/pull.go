package git

import (
	"context"
	"time"
)

// Pull pulls changes from the configured remote and branch.
func (e *Executor) Pull(dir, remote, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	args := []string{"pull"}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}

	_, err := e.RunWithContext(ctx, dir, args...)
	return err
}

// PullCurrent pulls changes for the current branch.
func (e *Executor) PullCurrent(dir string) error {
	return e.Pull(dir, "", "")
}

// SwitchAndPull switches to the target branch and pulls changes.
func (e *Executor) SwitchAndPull(dir, branch string) error {
	if err := e.SwitchBranch(dir, branch); err != nil {
		return err
	}
	return e.PullCurrent(dir)
}

// PullBranchIntoCurrent pulls remote branch changes directly into current branch.
// WARNING: This merges remote branch into the checked out branch.
func (e *Executor) PullBranchIntoCurrent(dir, remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}
	return e.Pull(dir, remote, branch)
}
