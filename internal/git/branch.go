package git

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Branch represents a local Git branch.
type Branch struct {
	Name        string
	IsCurrent   bool
	Upstream    string
	Subject     string
	CommitHash  string
	HasUpstream bool
}

// ListBranches lists all local branches with their upstream and current status.
func (e *Executor) ListBranches(dir string) ([]Branch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	format := "%(refname:short)|%(HEAD)|%(upstream:short)|%(subject)|%(objectname:short)"
	out, err := e.RunWithContext(ctx, dir, "for-each-ref", "--format="+format, "refs/heads")
	if err != nil {
		return nil, err
	}

	var branches []Branch
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) >= 1 {
			b := Branch{
				Name: parts[0],
			}
			if len(parts) > 1 && strings.TrimSpace(parts[1]) == "*" {
				b.IsCurrent = true
			}
			if len(parts) > 2 && strings.TrimSpace(parts[2]) != "" {
				b.Upstream = strings.TrimSpace(parts[2])
				b.HasUpstream = true
			}
			if len(parts) > 3 {
				b.Subject = strings.TrimSpace(parts[3])
			}
			if len(parts) > 4 {
				b.CommitHash = strings.TrimSpace(parts[4])
			}
			branches = append(branches, b)
		}
	}

	return branches, nil
}

// CreateBranch creates a new branch, optionally switching to it.
func (e *Executor) CreateBranch(dir, name string, switchAfter bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if switchAfter {
		_, err := e.RunWithContext(ctx, dir, "switch", "-c", name)
		return err
	}

	_, err := e.RunWithContext(ctx, dir, "branch", name)
	return err
}

// SwitchBranch switches to an existing branch.
func (e *Executor) SwitchBranch(dir, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.RunWithContext(ctx, dir, "switch", name)
	return err
}

// DeleteBranch deletes a branch (safe mode by default, force with flag).
func (e *Executor) DeleteBranch(dir, name string, force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	flag := "-d"
	if force {
		flag = "-D"
	}

	_, err := e.RunWithContext(ctx, dir, "branch", flag, name)
	return err
}

// ValidateBranchName checks if a branch name is valid in Git.
func (e *Executor) ValidateBranchName(dir, name string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := e.RunWithContext(ctx, dir, "check-ref-format", "--branch", name)
	return err == nil
}
