package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FileStatus represents the status of an individual file.
type FileStatus struct {
	Path        string
	OrigPath    string // Used for renames/copies
	StagedCode  byte   // 'M', 'A', 'D', 'R', 'C', ' ', '?'
	WorkingCode byte   // 'M', 'D', ' ', '?'
	IsUntracked bool
	IsConflict  bool
	IsStaged    bool
	IsModified  bool
	IsDeleted   bool
}

// DisplayStatus returns a short human-readable status badge for the file.
func (f FileStatus) DisplayStatus() string {
	if f.IsConflict {
		return "U"
	}
	if f.IsUntracked {
		return "?"
	}
	if f.StagedCode != ' ' && f.StagedCode != '.' {
		return string(f.StagedCode)
	}
	if f.WorkingCode != ' ' && f.WorkingCode != '.' {
		return string(f.WorkingCode)
	}
	return " "
}

// RepoState describes in-progress git operations.
type RepoState string

const (
	StateClean      RepoState = "Clean"
	StateModified   RepoState = "Modified"
	StateConflict   RepoState = "Conflict"
	StateRebase     RepoState = "Rebase in progress"
	StateMerge      RepoState = "Merge in progress"
	StateCherryPick RepoState = "Cherry-pick in progress"
	StateRevert     RepoState = "Revert in progress"
)

// StatusInfo encapsulates the full status of a git repository.
type StatusInfo struct {
	Branch       string
	Upstream     string
	IsDetached   bool
	DetachedOID  string
	Ahead        int
	Behind       int
	Files        []FileStatus
	StagedFiles  []FileStatus
	WorkingFiles []FileStatus
	Untracked    []FileStatus
	Conflicts    []FileStatus
	State        RepoState
	IsClean      bool
	HasUpstream  bool
}

// GetStatus queries git status and returns structured status information.
func (e *Executor) GetStatus(dir string) (*StatusInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := e.RunWithContext(ctx, dir, "status", "--branch", "--porcelain=v2", "-uall")
	if err != nil {
		return nil, err
	}

	info := parsePorcelainV2(out)

	// Check ongoing operations in git directory
	detectRepoState(dir, info)

	return info, nil
}

func parsePorcelainV2(output string) *StatusInfo {
	info := &StatusInfo{
		Files:        make([]FileStatus, 0),
		StagedFiles:  make([]FileStatus, 0),
		WorkingFiles: make([]FileStatus, 0),
		Untracked:    make([]FileStatus, 0),
		Conflicts:    make([]FileStatus, 0),
		State:        StateClean,
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "# ") {
			// Branch headers
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				switch parts[1] {
				case "branch.head":
					info.Branch = parts[2]
					if info.Branch == "(detached)" {
						info.IsDetached = true
					}
				case "branch.oid":
					if info.IsDetached && len(parts[2]) >= 7 {
						info.DetachedOID = parts[2][:7]
					}
				case "branch.upstream":
					info.Upstream = parts[2]
					info.HasUpstream = true
				case "branch.ab":
					if len(parts) >= 4 {
						aheadStr := strings.TrimPrefix(parts[2], "+")
						behindStr := strings.TrimPrefix(parts[3], "-")
						info.Ahead, _ = strconv.Atoi(aheadStr)
						info.Behind, _ = strconv.Atoi(behindStr)
					}
				}
			}
			continue
		}

		// File status lines
		if strings.HasPrefix(line, "1 ") {
			// Ordinary changed entries: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
			fields := strings.SplitN(line, " ", 9)
			if len(fields) >= 9 {
				xy := fields[1]
				path := fields[8]
				fs := FileStatus{
					Path:        path,
					StagedCode:  xy[0],
					WorkingCode: xy[1],
					IsStaged:    xy[0] != '.' && xy[0] != ' ',
					IsModified:  xy[1] != '.' && xy[1] != ' ',
					IsDeleted:   xy[0] == 'D' || xy[1] == 'D',
				}
				info.Files = append(info.Files, fs)
				if fs.IsStaged {
					info.StagedFiles = append(info.StagedFiles, fs)
				}
				if fs.IsModified || fs.IsDeleted {
					info.WorkingFiles = append(info.WorkingFiles, fs)
				}
			}
		} else if strings.HasPrefix(line, "2 ") {
			// Renamed / copied entries: 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
			fields := strings.SplitN(line, " ", 10)
			if len(fields) >= 10 {
				xy := fields[1]
				pathPart := fields[9]
				pathSub := strings.SplitN(pathPart, "\t", 2)
				path := pathSub[0]
				origPath := ""
				if len(pathSub) > 1 {
					origPath = pathSub[1]
				}

				fs := FileStatus{
					Path:        path,
					OrigPath:    origPath,
					StagedCode:  xy[0],
					WorkingCode: xy[1],
					IsStaged:    xy[0] != '.' && xy[0] != ' ',
					IsModified:  xy[1] != '.' && xy[1] != ' ',
				}
				info.Files = append(info.Files, fs)
				if fs.IsStaged {
					info.StagedFiles = append(info.StagedFiles, fs)
				}
				if fs.IsModified {
					info.WorkingFiles = append(info.WorkingFiles, fs)
				}
			}
		} else if strings.HasPrefix(line, "u ") {
			// Unmerged / conflict entries: u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
			fields := strings.SplitN(line, " ", 11)
			if len(fields) >= 11 {
				path := fields[10]
				fs := FileStatus{
					Path:       path,
					IsConflict: true,
				}
				info.Files = append(info.Files, fs)
				info.Conflicts = append(info.Conflicts, fs)
			}
		} else if strings.HasPrefix(line, "? ") {
			// Untracked entry: ? <path>
			path := strings.TrimPrefix(line, "? ")
			fs := FileStatus{
				Path:        path,
				WorkingCode: '?',
				IsUntracked: true,
			}
			info.Files = append(info.Files, fs)
			info.Untracked = append(info.Untracked, fs)
		}
	}

	info.IsClean = len(info.Files) == 0 && len(info.Conflicts) == 0
	if !info.IsClean {
		info.State = StateModified
	}
	if len(info.Conflicts) > 0 {
		info.State = StateConflict
	}

	return info
}

func detectRepoState(dir string, info *StatusInfo) {
	gitDir := filepath.Join(dir, ".git")

	// Check if gitDir is a file (e.g. submodule or worktree)
	fi, err := os.Stat(gitDir)
	if err == nil && !fi.IsDir() {
		// Read gitdir pointer
		content, err := os.ReadFile(gitDir)
		if err == nil {
			line := strings.TrimSpace(string(content))
			if strings.HasPrefix(line, "gitdir: ") {
				gitDir = strings.TrimPrefix(line, "gitdir: ")
				if !filepath.IsAbs(gitDir) {
					gitDir = filepath.Join(dir, gitDir)
				}
			}
		}
	}

	if exists(filepath.Join(gitDir, "rebase-merge")) || exists(filepath.Join(gitDir, "rebase-apply")) {
		info.State = StateRebase
	} else if exists(filepath.Join(gitDir, "MERGE_HEAD")) {
		info.State = StateMerge
	} else if exists(filepath.Join(gitDir, "CHERRY_PICK_HEAD")) {
		info.State = StateCherryPick
	} else if exists(filepath.Join(gitDir, "REVERT_HEAD")) {
		info.State = StateRevert
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SummaryString returns a concise human-readable summary of repository state.
func (s *StatusInfo) SummaryString() string {
	if len(s.Conflicts) > 0 {
		return fmt.Sprintf("! %d %s", len(s.Conflicts), plural(len(s.Conflicts), "conflict", "conflicts"))
	}
	if s.State != StateClean && s.State != StateModified {
		return string(s.State)
	}
	if s.IsClean {
		return "● Clean"
	}
	count := len(s.Files)
	return fmt.Sprintf("● %d %s changed", count, plural(count, "file", "files"))
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
