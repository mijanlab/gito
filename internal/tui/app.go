package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gito/internal/ai"
	"gito/internal/config"
	"gito/internal/git"
	"gito/internal/theme"
	"gito/internal/tui/components"
	"gito/internal/tui/screens"
	"gito/internal/updater"
)

// ScreenType identifies active screen.
type ScreenType int

const (
	ScreenNotGit ScreenType = iota
	ScreenMainMenu
	ScreenPush
	ScreenPull
	ScreenCommit
	ScreenBranch
	ScreenChanges
	ScreenHistory
	ScreenAIConfig
)

// App is the root Bubble Tea model managing all screens and Git interactions.
type App struct {
	Version      string
	WorkDir      string
	GitExec      *git.Executor
	RepoInfo     *git.RepoInfo
	Status       *git.StatusInfo
	Branches     []git.Branch
	AppConfig    *config.AppConfig
	Theme        *theme.Theme
	ActiveScreen ScreenType
	Loading      bool
	LoadingMsg   string
	Width        int
	Height       int

	// Screen models
	notGitScreen   *screens.NotGitScreen
	mainMenuScreen *screens.MainMenuScreen
	pushScreen     *screens.PushScreen
	pullScreen     *screens.PullScreen
	commitScreen   *screens.CommitScreen
	branchScreen   *screens.BranchScreen
	changesScreen  *screens.ChangesScreen
	historyScreen  *screens.HistoryScreen
	aiConfigScreen *screens.AIConfigScreen
}

// NewApp creates a new App model instance.
func NewApp(version, dir string) (*App, error) {
	gitExec, err := git.NewExecutor()
	if err != nil {
		return nil, err
	}

	appCfg, err := config.Load()
	if err != nil {
		appCfg = config.DefaultAppConfig()
	}

	th := theme.DefaultTheme()

	app := &App{
		Version:        version,
		WorkDir:        dir,
		GitExec:        gitExec,
		AppConfig:      appCfg,
		Theme:          th,
		ActiveScreen:   ScreenMainMenu,
		Width:          80,
		Height:         24,
		notGitScreen:   screens.NewNotGitScreen(th, dir),
		mainMenuScreen: screens.NewMainMenuScreen(th, version),
		pushScreen:     screens.NewPushScreen(th, version),
		pullScreen:     screens.NewPullScreen(th, version),
		commitScreen:   screens.NewCommitScreen(th, version),
		branchScreen:   screens.NewBranchScreen(th, version),
		changesScreen:  screens.NewChangesScreen(th, version),
		historyScreen:  screens.NewHistoryScreen(th, version),
		aiConfigScreen: screens.NewAIConfigScreen(th, version),
	}

	return app, nil
}

// Init triggers initial repository inspection and background update check on startup.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.checkAndRefreshCmd(),
		a.checkUpdateCmd(),
	)
}

// Messages for async operations
type repoStateMsg struct {
	RepoInfo *git.RepoInfo
	Status   *git.StatusInfo
	Branches []git.Branch
	Err      error
}

type updateAvailableMsg struct {
	HasUpdate bool
	Version   string
}

type asyncResultMsg struct {
	Action    string
	Success   bool
	Title     string
	Message   string
	ExtraHash string
	Err       error
}

type aiMessageGeneratedMsg struct {
	Message string
	Err     error
}

type aiTestMsg struct {
	Success bool
	Message string
}

type commitDetailLoadedMsg struct {
	Detail *git.CommitDetail
	Err    error
}

type commitDiffLoadedMsg struct {
	Diff string
	Err  error
}

// Update coordinates events across screens and background tasks.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width = msg.Width
		a.Height = msg.Height
		return a, nil

	case repoStateMsg:
		a.Loading = false
		if msg.Err != nil || msg.RepoInfo == nil || !msg.RepoInfo.IsInside {
			a.RepoInfo = msg.RepoInfo
			a.ActiveScreen = ScreenNotGit
			a.notGitScreen.Dir = a.WorkDir
			return a, nil
		}
		a.RepoInfo = msg.RepoInfo
		a.Status = msg.Status
		a.Branches = msg.Branches
		a.mainMenuScreen.InitMenu(a.Status)
		return a, nil

	case updateAvailableMsg:
		if msg.HasUpdate && msg.Version != "" {
			a.mainMenuScreen.Header.SetAvailableVersion(msg.Version)
			a.pushScreen.Header.SetAvailableVersion(msg.Version)
			a.pullScreen.Header.SetAvailableVersion(msg.Version)
			a.commitScreen.Header.SetAvailableVersion(msg.Version)
			a.branchScreen.Header.SetAvailableVersion(msg.Version)
			a.changesScreen.Header.SetAvailableVersion(msg.Version)
			a.historyScreen.Header.SetAvailableVersion(msg.Version)
			a.aiConfigScreen.Header.SetAvailableVersion(msg.Version)
		}
		return a, nil

	case aiMessageGeneratedMsg:
		a.Loading = false
		if msg.Err != nil {
			a.commitScreen.SetError("AI Generation Failed", msg.Err.Error())
		} else {
			a.commitScreen.SetAIMessage(msg.Message)
		}
		return a, nil

	case aiTestMsg:
		a.Loading = false
		a.aiConfigScreen.SetTestResult(msg.Success, msg.Message)
		return a, nil

	case commitDetailLoadedMsg:
		a.Loading = false
		if msg.Err == nil && msg.Detail != nil {
			a.historyScreen.SetDetail(msg.Detail)
		}
		return a, nil

	case commitDiffLoadedMsg:
		a.Loading = false
		if msg.Err == nil {
			a.historyScreen.SetCommitDiff(msg.Diff)
		}
		return a, nil

	case asyncResultMsg:
		a.Loading = false
		switch msg.Action {
		case "init":
			if msg.Success {
				return a, a.checkAndRefreshCmd()
			}
			a.notGitScreen.StatusMsg = msg.Message
		case "push":
			if msg.Success {
				a.pushScreen.SetSuccess(msg.Message)
			} else {
				a.pushScreen.SetError(msg.Title, msg.Message)
			}
		case "pull":
			if msg.Success {
				a.pullScreen.SetSuccess(msg.Message)
			} else {
				a.pullScreen.SetError(msg.Title, msg.Message)
			}
		case "commit":
			if msg.Success {
				a.commitScreen.SetSuccess(msg.ExtraHash, msg.Message)
			} else {
				a.commitScreen.SetError(msg.Title, msg.Message)
			}
		case "branch_create", "branch_switch":
			if msg.Success {
				a.branchScreen.SetSuccess(msg.Message)
			} else {
				a.branchScreen.SetError(msg.Title, msg.Message)
			}
		case "branch_delete_safe":
			if msg.Success {
				a.branchScreen.SetSuccess(msg.Message)
			} else {
				// Check if unmerged error
				if strings.Contains(strings.ToLower(msg.Message), "not fully merged") {
					a.branchScreen.SetForceConfirm(msg.ExtraHash)
				} else {
					a.branchScreen.SetError(msg.Title, msg.Message)
				}
			}
		case "branch_delete_force":
			if msg.Success {
				a.branchScreen.SetSuccess(msg.Message)
			} else {
				a.branchScreen.SetError(msg.Title, msg.Message)
			}
		case "revert":
			if msg.Success {
				a.historyScreen.SetSuccess(msg.Message)
			} else {
				a.historyScreen.SetError(msg.Title, msg.Message)
			}
		}
		return a, a.refreshSilentCmd()
	}

	if a.Loading {
		return a, nil
	}

	// Dispatch to active screen
	switch a.ActiveScreen {
	case ScreenNotGit:
		action, cmd := a.notGitScreen.Update(msg)
		if action == "init_repo" {
			a.Loading = true
			a.LoadingMsg = "Initializing repository..."
			return a, a.initRepoCmd(a.WorkDir)
		} else if strings.HasPrefix(action, "change_dir:") {
			newDir := strings.TrimPrefix(action, "change_dir:")
			a.WorkDir = newDir
			return a, a.checkAndRefreshCmd()
		} else if action == "exit" {
			return a, tea.Quit
		}
		return a, cmd

	case ScreenMainMenu:
		action, cmd := a.mainMenuScreen.Update(msg)
		switch action {
		case "push":
			a.pushScreen.Reset(a.Status, a.Branches)
			a.ActiveScreen = ScreenPush
			return a, nil
		case "pull":
			a.pullScreen.Reset(a.Status, a.Branches)
			a.ActiveScreen = ScreenPull
			return a, nil
		case "commit":
			diffText, _ := a.GitExec.GetDiff(a.WorkDir, false, "")
			a.commitScreen.Reset(a.Status, diffText, a.AppConfig.AI.Enabled)
			a.ActiveScreen = ScreenCommit
			return a, nil
		case "branch":
			a.branchScreen.Reset(a.Status, a.Branches)
			a.ActiveScreen = ScreenBranch
			return a, nil
		case "changes":
			diffText, _ := a.GitExec.GetDiff(a.WorkDir, false, "")
			stats, _ := a.GitExec.GetDiffStats(a.WorkDir)
			a.changesScreen.Reset(a.Status, stats, diffText)
			a.ActiveScreen = ScreenChanges
			return a, nil
		case "history":
			commits, _ := a.GitExec.GetCommitHistory(a.WorkDir, 50)
			a.historyScreen.Reset(a.Status, commits)
			a.ActiveScreen = ScreenHistory
			return a, nil
		case "ai":
			a.aiConfigScreen.Reset(a.AppConfig.AI)
			a.ActiveScreen = ScreenAIConfig
			return a, nil
		case "refresh":
			a.Loading = true
			a.LoadingMsg = "Refreshing repository..."
			return a, a.checkAndRefreshCmd()
		case "exit":
			return a, tea.Quit
		}
		return a, cmd

	case ScreenPush:
		action, cmd := a.pushScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, a.checkAndRefreshCmd()
		} else if strings.HasPrefix(action, "do_push:") {
			parts := strings.Split(action, ":")
			target := parts[1]
			a.Loading = true
			a.LoadingMsg = "Pushing commits..."
			switch target {
			case "current":
				return a, a.pushCmd("", a.Status.Branch, false, false)
			case "upstream":
				return a, a.pushCmd("origin", a.Status.Branch, true, false)
			case "branch":
				bName := parts[2]
				return a, a.pushCmd("origin", bName, false, false)
			case "force":
				bName := parts[2]
				return a, a.pushCmd("origin", bName, false, true)
			}
		}
		return a, cmd

	case ScreenPull:
		action, cmd := a.pullScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, a.checkAndRefreshCmd()
		} else if strings.HasPrefix(action, "do_pull:") {
			parts := strings.Split(action, ":")
			target := parts[1]
			a.Loading = true
			a.LoadingMsg = "Pulling changes..."
			switch target {
			case "current":
				return a, a.pullCmd("", "")
			case "switch_and_pull":
				bName := parts[2]
				return a, a.switchAndPullCmd(bName)
			case "pull_into_current":
				bName := parts[2]
				return a, a.pullCmd("origin", bName)
			}
		}
		return a, cmd

	case ScreenCommit:
		action, cmd := a.commitScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, a.checkAndRefreshCmd()
		} else if action == "do_ai_generate" {
			a.Loading = true
			a.LoadingMsg = "Generating commit message with AI..."
			return a, a.generateAICmd()
		} else if strings.HasPrefix(action, "do_commit:") {
			message := strings.TrimPrefix(action, "do_commit:")
			a.Loading = true
			a.LoadingMsg = "Creating commit..."
			return a, a.commitCmd(message)
		}
		return a, cmd

	case ScreenBranch:
		action, cmd := a.branchScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, a.checkAndRefreshCmd()
		} else if strings.HasPrefix(action, "do_branch:") {
			parts := strings.Split(action, ":")
			actionType := parts[1]
			bName := parts[2]
			switch actionType {
			case "create_switch":
				a.Loading = true
				a.LoadingMsg = "Creating and switching branch..."
				return a, a.branchCreateCmd(bName, true)
			case "create_only":
				a.Loading = true
				a.LoadingMsg = "Creating branch..."
				return a, a.branchCreateCmd(bName, false)
			case "switch":
				a.Loading = true
				a.LoadingMsg = "Switching branch..."
				return a, a.branchSwitchCmd(bName)
			case "delete_safe":
				a.Loading = true
				a.LoadingMsg = "Deleting branch..."
				return a, a.branchDeleteCmd(bName, false)
			case "delete_force":
				a.Loading = true
				a.LoadingMsg = "Force deleting branch..."
				return a, a.branchDeleteCmd(bName, true)
			}
		}
		return a, cmd

	case ScreenChanges:
		action, cmd := a.changesScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, a.checkAndRefreshCmd()
		} else if action == "commit" {
			diffText, _ := a.GitExec.GetDiff(a.WorkDir, false, "")
			a.commitScreen.Reset(a.Status, diffText, a.AppConfig.AI.Enabled)
			a.ActiveScreen = ScreenCommit
			return a, nil
		}
		return a, cmd

	case ScreenHistory:
		action, cmd := a.historyScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, a.checkAndRefreshCmd()
		} else if strings.HasPrefix(action, "view_commit_detail:") {
			hash := strings.TrimPrefix(action, "view_commit_detail:")
			a.Loading = true
			a.LoadingMsg = "Loading commit details..."
			return a, a.loadCommitDetailCmd(hash)
		} else if strings.HasPrefix(action, "view_commit_diff:") {
			hash := strings.TrimPrefix(action, "view_commit_diff:")
			a.Loading = true
			a.LoadingMsg = "Loading commit diff..."
			return a, a.loadCommitDiffCmd(hash)
		} else if strings.HasPrefix(action, "do_revert:") {
			hash := strings.TrimPrefix(action, "do_revert:")
			a.Loading = true
			a.LoadingMsg = "Reverting commit..."
			return a, a.revertCmd(hash)
		} else if action == "refresh" {
			commits, _ := a.GitExec.GetCommitHistory(a.WorkDir, 50)
			a.historyScreen.Reset(a.Status, commits)
			return a, nil
		}
		return a, cmd

	case ScreenAIConfig:
		action, cmd := a.aiConfigScreen.Update(msg)
		if action == "main_menu" {
			a.ActiveScreen = ScreenMainMenu
			return a, nil
		} else if action == "ai_test_connection" {
			a.Loading = true
			a.LoadingMsg = "Testing AI connection..."
			return a, a.testAICmd(a.aiConfigScreen.Config)
		} else if action == "ai_save_config" {
			a.AppConfig.AI = a.aiConfigScreen.Config
			_ = config.Save(a.AppConfig)
			return a, nil
		}
		return a, cmd
	}

	return a, nil
}

// Render renders the application view.
func (a *App) View() string {
	if a.Loading {
		boxWidth := a.Width - 4
		if boxWidth < 40 {
			boxWidth = 40
		}
		header := components.NewHeader(a.Theme, a.Version).Render(a.RepoInfo, a.Status, a.Width)
		loadingText := a.Theme.AIText.Render("◌ " + a.LoadingMsg)
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", loadingText)
		return lipgloss.Place(a.Width, a.Height, lipgloss.Center, lipgloss.Center, a.Theme.Card.Width(boxWidth).Render(content))
	}

	var viewContent string
	switch a.ActiveScreen {
	case ScreenNotGit:
		viewContent = a.notGitScreen.Render(a.Width)
	case ScreenMainMenu:
		viewContent = a.mainMenuScreen.Render(a.RepoInfo, a.Status, a.Width)
	case ScreenPush:
		viewContent = a.pushScreen.Render(a.RepoInfo, a.Width)
	case ScreenPull:
		viewContent = a.pullScreen.Render(a.RepoInfo, a.Width)
	case ScreenCommit:
		viewContent = a.commitScreen.Render(a.RepoInfo, a.Width)
	case ScreenBranch:
		viewContent = a.branchScreen.Render(a.RepoInfo, a.Width)
	case ScreenChanges:
		viewContent = a.changesScreen.Render(a.RepoInfo, a.Width)
	case ScreenHistory:
		viewContent = a.historyScreen.Render(a.RepoInfo, a.Width)
	case ScreenAIConfig:
		viewContent = a.aiConfigScreen.Render(a.RepoInfo, a.Status, a.Width)
	}

	return lipgloss.Place(a.Width, a.Height, lipgloss.Center, lipgloss.Center, viewContent)
}

// --- Commands ---

func (a *App) checkAndRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		repoInfo, err := a.GitExec.DetectRepository(a.WorkDir)
		if err != nil || !repoInfo.IsInside {
			return repoStateMsg{RepoInfo: repoInfo, Err: err}
		}

		status, _ := a.GitExec.GetStatus(a.WorkDir)
		branches, _ := a.GitExec.ListBranches(a.WorkDir)

		return repoStateMsg{
			RepoInfo: repoInfo,
			Status:   status,
			Branches: branches,
		}
	}
}

func (a *App) refreshSilentCmd() tea.Cmd {
	return func() tea.Msg {
		repoInfo, _ := a.GitExec.DetectRepository(a.WorkDir)
		if repoInfo != nil && repoInfo.IsInside {
			status, _ := a.GitExec.GetStatus(a.WorkDir)
			branches, _ := a.GitExec.ListBranches(a.WorkDir)
			return repoStateMsg{
				RepoInfo: repoInfo,
				Status:   status,
				Branches: branches,
			}
		}
		return nil
	}
}

func (a *App) initRepoCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.InitRepository(dir)
		if err != nil {
			return asyncResultMsg{Action: "init", Success: false, Message: err.Error(), Err: err}
		}
		return asyncResultMsg{Action: "init", Success: true, Message: "Repository initialized successfully"}
	}
}

func (a *App) pushCmd(remote, branch string, setUpstream, force bool) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.Push(a.WorkDir, remote, branch, setUpstream, force)
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{
				Action:  "push",
				Success: false,
				Title:   title,
				Message: exp,
				Err:     err,
			}
		}
		target := branch
		if remote != "" {
			target = fmt.Sprintf("%s → %s/%s", branch, remote, branch)
		}
		return asyncResultMsg{
			Action:  "push",
			Success: true,
			Message: target,
		}
	}
}

func (a *App) pullCmd(remote, branch string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if remote == "" && branch == "" {
			err = a.GitExec.PullCurrent(a.WorkDir)
		} else {
			err = a.GitExec.PullBranchIntoCurrent(a.WorkDir, remote, branch)
		}
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{
				Action:  "pull",
				Success: false,
				Title:   title,
				Message: exp,
				Err:     err,
			}
		}
		return asyncResultMsg{
			Action:  "pull",
			Success: true,
			Message: "Successfully pulled newest changes into current branch",
		}
	}
}

func (a *App) switchAndPullCmd(branch string) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.SwitchAndPull(a.WorkDir, branch)
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{
				Action:  "pull",
				Success: false,
				Title:   title,
				Message: exp,
				Err:     err,
			}
		}
		return asyncResultMsg{
			Action:  "pull",
			Success: true,
			Message: fmt.Sprintf("Switched to %s and pulled latest changes", branch),
		}
	}
}

func (a *App) commitCmd(message string) tea.Cmd {
	return func() tea.Msg {
		res, err := a.GitExec.Commit(a.WorkDir, message, true)
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{
				Action:  "commit",
				Success: false,
				Title:   title,
				Message: exp,
				Err:     err,
			}
		}
		return asyncResultMsg{
			Action:    "commit",
			Success:   true,
			ExtraHash: res.Hash,
			Message:   res.Subject,
		}
	}
}

func (a *App) generateAICmd() tea.Cmd {
	return func() tea.Msg {
		diff, err := a.GitExec.GetComprehensiveDiff(a.WorkDir)
		if err != nil {
			return aiMessageGeneratedMsg{Err: fmt.Errorf("could not retrieve diff: %w", err)}
		}

		provider := ai.NewProvider(a.AppConfig.AI)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		msg, err := provider.GenerateCommitMessage(ctx, diff)
		return aiMessageGeneratedMsg{Message: msg, Err: err}
	}
}

func (a *App) testAICmd(cfg ai.Config) tea.Cmd {
	return func() tea.Msg {
		provider := ai.NewProvider(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := provider.TestConnection(ctx)
		if err != nil {
			return aiTestMsg{Success: false, Message: err.Error()}
		}
		return aiTestMsg{Success: true, Message: "Connected"}
	}
}

func (a *App) branchCreateCmd(name string, switchAfter bool) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.CreateBranch(a.WorkDir, name, switchAfter)
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{Action: "branch_create", Success: false, Title: title, Message: exp, Err: err}
		}
		actionDesc := "created"
		if switchAfter {
			actionDesc = "created and checked out"
		}
		return asyncResultMsg{Action: "branch_create", Success: true, Message: fmt.Sprintf("Branch '%s' %s", name, actionDesc)}
	}
}

func (a *App) branchSwitchCmd(name string) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.SwitchBranch(a.WorkDir, name)
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{Action: "branch_switch", Success: false, Title: title, Message: exp, Err: err}
		}
		return asyncResultMsg{Action: "branch_switch", Success: true, Message: fmt.Sprintf("Switched to branch '%s'", name)}
	}
}

func (a *App) branchDeleteCmd(name string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.DeleteBranch(a.WorkDir, name, force)
		action := "branch_delete_safe"
		if force {
			action = "branch_delete_force"
		}
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{Action: action, Success: false, Title: title, Message: exp, ExtraHash: name, Err: err}
		}
		return asyncResultMsg{Action: action, Success: true, Message: fmt.Sprintf("Branch '%s' deleted", name)}
	}
}

func (a *App) loadCommitDetailCmd(hash string) tea.Cmd {
	return func() tea.Msg {
		detail, err := a.GitExec.GetCommitDetails(a.WorkDir, hash)
		return commitDetailLoadedMsg{Detail: detail, Err: err}
	}
}

func (a *App) loadCommitDiffCmd(hash string) tea.Cmd {
	return func() tea.Msg {
		diff, err := a.GitExec.GetCommitDiff(a.WorkDir, hash)
		return commitDiffLoadedMsg{Diff: diff, Err: err}
	}
}

func (a *App) revertCmd(hash string) tea.Cmd {
	return func() tea.Msg {
		err := a.GitExec.RevertCommit(a.WorkDir, hash)
		if err != nil {
			title, exp := getFriendlyError(err)
			return asyncResultMsg{Action: "revert", Success: false, Title: title, Message: exp, Err: err}
		}
		return asyncResultMsg{Action: "revert", Success: true, Message: fmt.Sprintf("Reverted commit %s", hash[:7])}
	}
}

func (a *App) checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		rel, hasUpdate, err := updater.CheckLatestVersion(a.Version)
		if err != nil || rel == nil || !hasUpdate {
			return updateAvailableMsg{HasUpdate: false}
		}
		return updateAvailableMsg{
			HasUpdate: true,
			Version:   rel.TagName,
		}
	}
}

func getFriendlyError(err error) (string, string) {
	if gitErr, ok := err.(*git.GitError); ok {
		return gitErr.UserFriendlyError()
	}
	return "Operation Failed", err.Error()
}
