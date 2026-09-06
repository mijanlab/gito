package git

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DiffStats holds summary line additions/deletions/files changed.
type DiffStats struct {
	FilesChanged int
	Additions    int
	Deletions    int
}

// GetDiff retrieves the git diff for working tree or staged changes.
func (e *Executor) GetDiff(dir string, staged bool, filePath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	if filePath != "" {
		args = append(args, "--", filePath)
	}

	out, err := e.RunWithContext(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	return out, nil
}

// GetCommitDiff retrieves the diff introduced by a specific commit.
func (e *Executor) GetCommitDiff(dir, hash string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := e.RunWithContext(ctx, dir, "show", "--patch", "--format=", hash)
	if err != nil {
		return "", err
	}
	return out, nil
}

// GetDiffStats retrieves diffstat counts for changes in the working tree.
func (e *Executor) GetDiffStats(dir string) (DiffStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stats DiffStats
	out, err := e.RunWithContext(ctx, dir, "diff", "HEAD", "--shortstat")
	if err != nil {
		// If HEAD doesn't exist yet (empty initial repo), try unstaged + staged
		out, _ = e.RunWithContext(ctx, dir, "diff", "--shortstat")
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return stats, nil
	}

	// Format: " 3 files changed, 42 insertions(+), 8 deletions(-)"
	parts := strings.Split(out, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) >= 2 {
			count, _ := strconv.Atoi(fields[0])
			if strings.Contains(fields[1], "file") {
				stats.FilesChanged = count
			} else if strings.Contains(fields[1], "insertion") || strings.Contains(fields[1], "addition") {
				stats.Additions = count
			} else if strings.Contains(fields[1], "deletion") {
				stats.Deletions = count
			}
		}
	}

	return stats, nil
}

// GetComprehensiveDiff gathers all changes (staged, unstaged, untracked) for AI commit generation.
func (e *Executor) GetComprehensiveDiff(dir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Try git diff HEAD to capture all staged and tracked changes
	out, err := e.RunWithContext(ctx, dir, "diff", "HEAD")
	if err != nil {
		// If HEAD is unborn, get staged and unstaged diffs
		stagedDiff, _ := e.RunWithContext(ctx, dir, "diff", "--staged")
		unstagedDiff, _ := e.RunWithContext(ctx, dir, "diff")
		out = stagedDiff + "\n" + unstagedDiff
	}

	var sb strings.Builder
	sb.WriteString(out)

	// 2. Also check untracked files and append their first few lines for context
	status, err := e.GetStatus(dir)
	if err == nil && len(status.Untracked) > 0 {
		for _, u := range status.Untracked {
			sb.WriteString("\n--- /dev/null\n+++ b/" + u.Path + "\n")
			fullPath := filepath.Join(dir, u.Path)
			data, readErr := os.ReadFile(fullPath)
			if readErr == nil {
				// Only include up to 4KB per untracked file to prevent overloading context
				if len(data) > 4096 {
					data = data[:4096]
				}
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					sb.WriteString("+" + line + "\n")
				}
			}
		}
	}

	return strings.TrimSpace(sb.String()), nil
}
