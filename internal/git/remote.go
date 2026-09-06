package git

import (
	"context"
	"strings"
	"time"
)

// Remote represents a configured Git remote repository.
type Remote struct {
	Name string
	URL  string
}

// ListRemotes returns all configured remotes.
func (e *Executor) ListRemotes(dir string) ([]Remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := e.RunWithContext(ctx, dir, "remote", "-v")
	if err != nil {
		return nil, err
	}

	remoteMap := make(map[string]string)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := fields[0]
			url := fields[1]
			remoteMap[name] = url
		}
	}

	var remotes []Remote
	// Ensure origin comes first if present
	if url, ok := remoteMap["origin"]; ok {
		remotes = append(remotes, Remote{Name: "origin", URL: url})
	}
	for name, url := range remoteMap {
		if name != "origin" {
			remotes = append(remotes, Remote{Name: name, URL: url})
		}
	}

	return remotes, nil
}

// DefaultRemote returns the most likely default remote (origin, or first available).
func (e *Executor) DefaultRemote(dir string) (string, error) {
	remotes, err := e.ListRemotes(dir)
	if err != nil {
		return "", err
	}
	if len(remotes) == 0 {
		return "", nil
	}
	return remotes[0].Name, nil
}
