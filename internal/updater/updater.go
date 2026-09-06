package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gito/internal/config"
)

const RepoOwner = "mijanlab"
const RepoName = "gito"

// ReleaseInfo contains GitHub release metadata.
type ReleaseInfo struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
}

// CheckLatestVersion queries GitHub API for the latest release.
func CheckLatestVersion(currentVersion string) (*ReleaseInfo, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "gito-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No release exists on GitHub yet
		return nil, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, false, err
	}

	latestTag := strings.TrimPrefix(rel.TagName, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")

	hasUpdate := isNewer(latestTag, currentClean)
	return &rel, hasUpdate, nil
}

// isNewer returns true if remote version is strictly newer than current version.
func isNewer(latest, current string) bool {
	if latest == "" || current == "" || latest == current {
		return false
	}
	// Simple semver comparison (e.g. 0.2.0 > 0.1.0)
	lParts := strings.Split(latest, ".")
	cParts := strings.Split(current, ".")

	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		if lParts[i] > cParts[i] {
			return true
		} else if lParts[i] < cParts[i] {
			return false
		}
	}
	return len(lParts) > len(cParts)
}

// SelfUpdate downloads and replaces the current gito binary with the newest release.
func SelfUpdate(currentVersion string) error {
	rel, hasUpdate, err := CheckLatestVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("could not check latest version: %w", err)
	}

	if !hasUpdate {
		fmt.Printf("✓ You are already using the latest version of Gito (v%s).\n", strings.TrimPrefix(currentVersion, "v"))
		return nil
	}

	latestVersion := strings.TrimPrefix(rel.TagName, "v")
	fmt.Printf("Updating Gito from v%s to v%s...\n", strings.TrimPrefix(currentVersion, "v"), latestVersion)

	osName := runtime.GOOS
	archName := runtime.GOARCH

	binaryAssetName := fmt.Sprintf("gito-%s-%s", osName, archName)
	if osName == "windows" {
		binaryAssetName += ".exe"
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", RepoOwner, RepoName, rel.TagName, binaryAssetName)

	// Fetch current executable location
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve symlink for executable: %w", err)
	}

	// Download binary to temp file
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "gito-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update download returned HTTP %d for %s", resp.StatusCode, downloadURL)
	}

	tempDir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(tempDir, "gito-update-*")
	if err != nil {
		// Fallback to system temp
		tmpFile, err = os.CreateTemp("", "gito-update-*")
		if err != nil {
			return fmt.Errorf("could not create temporary update file: %w", err)
		}
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write update file: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("could not set executable permissions: %w", err)
	}

	// Windows doesn't allow overwriting a running executable, so rename old binary first
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("could not move existing binary: %w", err)
		}
		if err := copyFile(tmpPath, execPath); err != nil {
			// Rollback
			_ = os.Rename(oldPath, execPath)
			return fmt.Errorf("could not install new binary: %w", err)
		}
	} else {
		// Unix atomic rename
		if err := os.Rename(tmpPath, execPath); err != nil {
			// If cross-device, copy
			if err := copyFile(tmpPath, execPath); err != nil {
				return fmt.Errorf("could not replace binary at %s: %w", execPath, err)
			}
		}
	}

	fmt.Printf("✓ Gito successfully updated to v%s at %s\n", latestVersion, execPath)
	return nil
}

// Uninstall removes the gito binary and optionally cleans configuration.
func Uninstall(purgeConfig bool) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	fmt.Printf("Removing Gito binary at %s...\n", execPath)
	if err := os.Remove(execPath); err != nil {
		return fmt.Errorf("could not remove binary at %s: %w (try running with sudo / admin)", execPath, err)
	}

	if purgeConfig {
		cfgPath, err := config.GetConfigPath()
		if err == nil {
			cfgDir := filepath.Dir(cfgPath)
			if err := os.RemoveAll(cfgDir); err == nil {
				fmt.Printf("Removed configuration directory at %s\n", cfgDir)
			}
		}
	}

	fmt.Println("✓ Gito uninstalled successfully.")
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
