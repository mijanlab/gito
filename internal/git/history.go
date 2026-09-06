package git

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CommitInfo holds brief commit summary for timeline list.
type CommitInfo struct {
	Hash         string
	ShortHash    string
	AuthorName   string
	RelativeDate string
	Subject      string
}

// CommitFileChange represents a single file change in a commit.
type CommitFileChange struct {
	Status string // M, A, D, R
	Path   string
}

// CommitDetail contains full commit details.
type CommitDetail struct {
	Hash         string
	ShortHash    string
	AuthorName   string
	AuthorEmail  string
	RelativeDate string
	Date         string
	Subject      string
	Body         string
	Files        []CommitFileChange
}

// GetCommitHistory returns the recent commit history list.
func (e *Executor) GetCommitHistory(dir string, limit int) ([]CommitInfo, error) {
	if limit <= 0 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	format := "%H|%h|%an|%cr|%s"
	out, err := e.RunWithContext(ctx, dir, "log", fmt.Sprintf("-n%d", limit), "--format="+format)
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) >= 5 {
			commits = append(commits, CommitInfo{
				Hash:         parts[0],
				ShortHash:    parts[1],
				AuthorName:   parts[2],
				RelativeDate: parts[3],
				Subject:      parts[4],
			})
		}
	}

	return commits, nil
}

// GetCommitDetails returns detailed information for a single commit.
func (e *Executor) GetCommitDetails(dir, hash string) (*CommitDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	format := "%H%x00%h%x00%an%x00%ae%x00%cr%x00%cd%x00%s%x00%b"
	out, err := e.RunWithContext(ctx, dir, "show", "--no-patch", "--format="+format, hash)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(out, "\x00")
	if len(parts) < 8 {
		return nil, fmt.Errorf("unexpected commit show output format")
	}

	detail := &CommitDetail{
		Hash:         strings.TrimSpace(parts[0]),
		ShortHash:    strings.TrimSpace(parts[1]),
		AuthorName:   strings.TrimSpace(parts[2]),
		AuthorEmail:  strings.TrimSpace(parts[3]),
		RelativeDate: strings.TrimSpace(parts[4]),
		Date:         strings.TrimSpace(parts[5]),
		Subject:      strings.TrimSpace(parts[6]),
		Body:         strings.TrimSpace(parts[7]),
	}

	// Fetch changed files using name-status
	filesOut, err := e.RunWithContext(ctx, dir, "show", "--name-status", "--oneline", "--format=", hash)
	if err == nil {
		fLines := strings.Split(filesOut, "\n")
		for _, fLine := range fLines {
			fLine = strings.TrimSpace(fLine)
			if fLine == "" {
				continue
			}
			fParts := strings.Fields(fLine)
			if len(fParts) >= 2 {
				detail.Files = append(detail.Files, CommitFileChange{
					Status: fParts[0],
					Path:   strings.Join(fParts[1:], " "),
				})
			}
		}
	}

	return detail, nil
}

// RevertCommit creates a new revert commit for the specified commit hash.
func (e *Executor) RevertCommit(dir, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := e.RunWithContext(ctx, dir, "revert", "--no-edit", hash)
	return err
}
