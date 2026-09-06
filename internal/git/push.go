package git

import (
	"context"
	"fmt"
	"time"
)

// Push pushes commits to a remote repository.
func (e *Executor) Push(dir, remote, branch string, setUpstream bool, force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if remote == "" {
		remote = "origin"
	}

	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u")
	}
	if force {
		args = append(args, "--force-with-lease")
	}

	args = append(args, remote)
	if branch != "" {
		args = append(args, branch)
	}

	_, err := e.RunWithContext(ctx, dir, args...)
	if err != nil {
		return err
	}
	return nil
}

// PushCurrent pushes the current branch. If setUpstream is true, sets tracking upstream.
func (e *Executor) PushCurrent(dir string, setUpstream bool, force bool) error {
	status, err := e.GetStatus(dir)
	if err != nil {
		return err
	}
	if status.Branch == "" || status.IsDetached {
		return fmt.Errorf("cannot push in detached HEAD state")
	}

	remote := "origin"
	remotes, _ := e.ListRemotes(dir)
	if len(remotes) > 0 {
		remote = remotes[0].Name
	}

	return e.Push(dir, remote, status.Branch, setUpstream, force)
}
