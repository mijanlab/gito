package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RepoInfo contains general information about a git repository.
type RepoInfo struct {
	Path     string // Absolute path to repo root
	Name     string // Folder name of repo
	IsInside bool   // True if inside a valid git repo
}

// DetectRepository inspects the given directory to determine if it is a Git repository.
func (e *Executor) DetectRepository(dir string) (*RepoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := e.RunWithContext(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return &RepoInfo{
			Path:     dir,
			Name:     filepath.Base(dir),
			IsInside: false,
		}, nil
	}

	rootPath := strings.TrimSpace(out)
	return &RepoInfo{
		Path:     rootPath,
		Name:     filepath.Base(rootPath),
		IsInside: true,
	}, nil
}

// InitRepository initializes a new git repository in the specified directory.
func (e *Executor) InitRepository(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	_, err := e.RunWithContext(ctx, dir, "init")
	return err
}
