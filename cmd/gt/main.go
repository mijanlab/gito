package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gito/internal/cli"
	"gito/internal/tui"
	"gito/internal/updater"
)

const version = "0.0.5"

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading working directory: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "push":
			autoAccept := false
			for _, arg := range os.Args[2:] {
				if arg == "-y" || arg == "--yes" {
					autoAccept = true
				}
			}
			runner, err := cli.NewRunner(workDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if err := runner.Push(autoAccept); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)

		case "pull":
			runner, err := cli.NewRunner(workDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if err := runner.Pull(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)

		case "status", "st":
			runner, err := cli.NewRunner(workDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if err := runner.Status(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)

		case "--version", "-v", "version":
			printVersion()
			os.Exit(0)
		case "update", "--update", "upgrade":
			runUpdate()
			os.Exit(0)
		case "uninstall", "--uninstall":
			runUninstall()
			os.Exit(0)
		case "--help", "-h", "help":
			printHelp()
			os.Exit(0)
		}
	}

	// Verify git executable exists on host system
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: 'git' is not installed or not in PATH.\nPlease install Git to use Gito: https://git-scm.com\n")
		os.Exit(1)
	}

	// Optional path argument
	if len(os.Args) > 1 && !isFlag(os.Args[1]) {
		workDir = os.Args[1]
	}

	app, err := tui.NewApp(version, workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Gito: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Gito: %v\n", err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("gt (Gito) version: v%s\n", strings.TrimPrefix(version, "v"))

	// Check if a newer version is available
	rel, hasUpdate, err := updater.CheckLatestVersion(version)
	if err == nil && rel != nil && hasUpdate {
		fmt.Printf("Available Version: %s\n", rel.TagName)
		fmt.Printf("Run 'gt update' to upgrade to the latest version.\n")
	}
}

func runUpdate() {
	fmt.Println("◈ Checking for Gito / gt updates...")
	if err := updater.SelfUpdate(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating Gito: %v\n", err)
		os.Exit(1)
	}
}

func runUninstall() {
	purgeConfig := false
	for _, arg := range os.Args[2:] {
		if arg == "--purge" || arg == "-p" {
			purgeConfig = true
		}
	}

	fmt.Print("Are you sure you want to uninstall gt / Gito? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Uninstall cancelled.")
		return
	}

	if err := updater.Uninstall(purgeConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error uninstalling Gito: %v\n", err)
		os.Exit(1)
	}
}

func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func printHelp() {
	fmt.Println(`gt (Gito) — Premium Interactive Git TUI & Fast CLI

Usage:
  gt [directory]
  gt [command] [flags]

Commands:
  push [-y]        AI-commit uncommitted changes & push to remote
  pull             Pull latest changes safely from upstream branch
  status           Show clean status of current branch and working tree
  update           Check for updates and self-upgrade to latest version
  uninstall        Remove binary from system (use --purge to remove config)
  version          Show current version and check for available updates

Flags:
  -y, --yes        Auto-accept AI commit message on push
  -v, --version    Show version
  -h, --help       Show help options

Keyboard Controls (Interactive Mode):
  ↑ / k            Navigate up
  ↓ / j            Navigate down
  Enter            Select / Confirm
  Esc              Back to previous screen
  r                Refresh repository state
  q                Quit Gito
  ?                Help`)
}
