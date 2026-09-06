package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gito/internal/ai"
	"gito/internal/config"
	"gito/internal/git"
)

// Runner executes direct CLI commands like `gt push`, `gt pull`, `gt status`.
type Runner struct {
	GitExec *git.Executor
	Config  *config.AppConfig
	WorkDir string
}

// NewRunner creates a new CLI command runner.
func NewRunner(workDir string) (*Runner, error) {
	gitExec, err := git.NewExecutor()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultAppConfig()
	}

	return &Runner{
		GitExec: gitExec,
		Config:  cfg,
		WorkDir: workDir,
	}, nil
}

// Push handles `gt push`:
// 1. If uncommitted changes exist, generates an AI commit message, commits them, and pushes.
// 2. If commits exist locally, pushes them to remote.
// 3. Handles setting upstream automatically if not configured.
func (r *Runner) Push(autoAccept bool) error {
	repoInfo, err := r.GitExec.DetectRepository(r.WorkDir)
	if err != nil || !repoInfo.IsInside {
		return fmt.Errorf("fatal: not a git repository (or any of the parent directories)")
	}

	status, err := r.GitExec.GetStatus(r.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}

	if status.IsDetached {
		return fmt.Errorf("cannot push in detached HEAD state")
	}

	// 1. Check if uncommitted changes exist
	if len(status.Files) > 0 {
		fmt.Printf("● Detected %d uncommitted %s on '%s'.\n", len(status.Files), plural(len(status.Files), "change", "changes"), status.Branch)
		for i, f := range status.Files {
			if i >= 5 {
				fmt.Printf("  ... and %d more files\n", len(status.Files)-5)
				break
			}
			fmt.Printf("  %s %s\n", f.DisplayStatus(), f.Path)
		}

		var commitMsg string
		if r.Config.AI.Enabled {
			fmt.Print("✦ Generating AI commit message from diff... ")
			diff, err := r.GitExec.GetComprehensiveDiff(r.WorkDir)
			if err == nil && diff != "" {
				provider := ai.NewProvider(r.Config.AI)
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				msg, aiErr := provider.GenerateCommitMessage(ctx, diff)
				cancel()
				if aiErr == nil && msg != "" {
					commitMsg = msg
					fmt.Printf("Done!\n\n✦ AI suggested message:\n  \"%s\"\n\n", commitMsg)
				} else {
					fmt.Printf("Failed (%v)\n", aiErr)
				}
			} else {
				fmt.Println("Empty diff")
			}
		}

		if commitMsg == "" {
			if autoAccept {
				commitMsg = fmt.Sprintf("Update %d %s", len(status.Files), plural(len(status.Files), "file", "files"))
			} else {
				fmt.Print("Enter commit message: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				commitMsg = strings.TrimSpace(input)
				if commitMsg == "" {
					return fmt.Errorf("aborted: commit message cannot be empty")
				}
			}
		} else if !autoAccept {
			fmt.Print("Press Enter to use AI message, type a new message, or 'c' to cancel: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if strings.ToLower(input) == "c" {
				fmt.Println("Push cancelled.")
				return nil
			}
			if input != "" {
				commitMsg = input
			}
		}

		// Commit changes
		fmt.Printf("Commiting changes... ")
		res, err := r.GitExec.Commit(r.WorkDir, commitMsg, true)
		if err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}
		fmt.Printf("✓ [%s] %s\n", res.Hash, res.Subject)
	}

	// 2. Refresh status before push
	status, err = r.GitExec.GetStatus(r.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to refresh status: %w", err)
	}

	// Check if up-to-date
	if status.HasUpstream && status.Ahead == 0 && len(status.Files) == 0 {
		fmt.Printf("✓ Branch '%s' is up-to-date with '%s'. Nothing to push.\n", status.Branch, status.Upstream)
		return nil
	}

	// 3. Push to remote
	remote := "origin"
	remotes, _ := r.GitExec.ListRemotes(r.WorkDir)
	if len(remotes) > 0 {
		remote = remotes[0].Name
	}

	if !status.HasUpstream {
		fmt.Printf("Pushing '%s' and setting upstream to %s/%s... ", status.Branch, remote, status.Branch)
		err = r.GitExec.Push(r.WorkDir, remote, status.Branch, true, false)
		if err != nil {
			if gitErr, ok := err.(*git.GitError); ok {
				title, exp := gitErr.UserFriendlyError()
				return fmt.Errorf("%s: %s\nDetails: %s", title, exp, gitErr.Stderr)
			}
			return err
		}
		fmt.Printf("✓ Done!\n✓ %s → %s/%s\n", status.Branch, remote, status.Branch)
	} else {
		fmt.Printf("Pushing commits on '%s' to %s... ", status.Branch, status.Upstream)
		err = r.GitExec.Push(r.WorkDir, remote, status.Branch, false, false)
		if err != nil {
			if gitErr, ok := err.(*git.GitError); ok {
				title, exp := gitErr.UserFriendlyError()
				return fmt.Errorf("%s: %s\nDetails: %s", title, exp, gitErr.Stderr)
			}
			return err
		}
		fmt.Printf("✓ Done!\n✓ Pushed to %s\n", status.Upstream)
	}

	return nil
}

// Pull handles `gt pull`:
// Pulls newest changes from upstream tracking branch cleanly.
func (r *Runner) Pull() error {
	repoInfo, err := r.GitExec.DetectRepository(r.WorkDir)
	if err != nil || !repoInfo.IsInside {
		return fmt.Errorf("fatal: not a git repository (or any of the parent directories)")
	}

	status, err := r.GitExec.GetStatus(r.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}

	fmt.Printf("Pulling latest changes for branch '%s'... ", status.Branch)
	err = r.GitExec.PullCurrent(r.WorkDir)
	if err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			title, exp := gitErr.UserFriendlyError()
			return fmt.Errorf("\n✕ %s: %s\nDetails: %s", title, exp, gitErr.Stderr)
		}
		return err
	}

	fmt.Printf("✓ Done!\n✓ Branch '%s' is up-to-date.\n", status.Branch)
	return nil
}

// Status prints a concise, clean summary of repository state.
func (r *Runner) Status() error {
	repoInfo, err := r.GitExec.DetectRepository(r.WorkDir)
	if err != nil || !repoInfo.IsInside {
		return fmt.Errorf("fatal: not a git repository (or any of the parent directories)")
	}

	status, err := r.GitExec.GetStatus(r.WorkDir)
	if err != nil {
		return err
	}

	fmt.Printf("On branch %s\n", status.Branch)
	if status.HasUpstream {
		fmt.Printf("Tracking %s (Ahead: %d, Behind: %d)\n", status.Upstream, status.Ahead, status.Behind)
	} else {
		fmt.Printf("No upstream configured.\n")
	}

	if status.IsClean {
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}

	fmt.Printf("\nChanges (%d files):\n", len(status.Files))
	for _, f := range status.Files {
		fmt.Printf("  %s %s\n", f.DisplayStatus(), f.Path)
	}

	return nil
}

func plural(n int, s, p string) string {
	if n == 1 {
		return s
	}
	return p
}
