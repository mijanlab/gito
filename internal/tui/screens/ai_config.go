package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gito/internal/ai"
	"gito/internal/git"
	"gito/internal/theme"
	"gito/internal/tui/components"
)

type aiConfigMode int

const (
	aiModeMain aiConfigMode = iota
	aiModeInputBaseURL
	aiModeInputModel
	aiModeInputAPIKey
	aiModeTesting
	aiModeTestResult
)

// AIConfigScreen manages AI provider settings and testing.
type AIConfigScreen struct {
	Theme       *theme.Theme
	Header      *components.Header
	Footer      *components.Footer
	Menu        *components.Menu
	Mode        aiConfigMode
	Config      ai.Config
	InputVal    string
	TestSuccess bool
	TestMsg     string
}

// NewAIConfigScreen creates a new AIConfigScreen.
func NewAIConfigScreen(theme *theme.Theme, version string) *AIConfigScreen {
	return &AIConfigScreen{
		Theme:  theme,
		Header: components.NewHeader(theme, version),
		Footer: components.NewFooter(theme),
		Mode:   aiModeMain,
	}
}

// Reset initializes AI configuration screen.
func (s *AIConfigScreen) Reset(cfg ai.Config) {
	s.Config = cfg
	s.Mode = aiModeMain
	s.TestMsg = ""

	toggleTitle := "Disable AI"
	if !s.Config.Enabled {
		toggleTitle = "Enable AI"
	}

	items := []components.MenuItem{
		{ID: "test_connection", Title: "Test connection", Icon: "✦"},
		{ID: "toggle_ai", Title: toggleTitle, Icon: "⏻"},
		{ID: "edit_base_url", Title: "Change Base URL", Icon: "✎"},
		{ID: "edit_model", Title: "Change Model", Icon: "✎"},
		{ID: "edit_api_key", Title: "Change API Key", Icon: "✎"},
		{ID: "reset_ollama", Title: "Reset to Ollama defaults", Icon: "↺"},
		{ID: "back", Title: "Back", Icon: "←"},
	}
	s.Menu = components.NewMenu(s.Theme, items)
}

// SetTestResult sets the connection test outcome.
func (s *AIConfigScreen) SetTestResult(success bool, msg string) {
	s.Mode = aiModeTestResult
	s.TestSuccess = success
	s.TestMsg = msg
}

// Update handles keyboard navigation for AIConfigScreen.
func (s *AIConfigScreen) Update(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.Mode {
		case aiModeMain:
			switch msg.String() {
			case "up", "k":
				s.Menu.MoveUp()
			case "down", "j":
				s.Menu.MoveDown()
			case "enter":
				selected := s.Menu.Selected()
				if selected != nil {
					switch selected.ID {
					case "test_connection":
						s.Mode = aiModeTesting
						return "ai_test_connection", nil
					case "toggle_ai":
						s.Config.Enabled = !s.Config.Enabled
						s.Reset(s.Config)
						return "ai_save_config", nil
					case "edit_base_url":
						s.InputVal = s.Config.BaseURL
						s.Mode = aiModeInputBaseURL
						return "", nil
					case "edit_model":
						s.InputVal = s.Config.Model
						s.Mode = aiModeInputModel
						return "", nil
					case "edit_api_key":
						s.InputVal = s.Config.APIKey
						s.Mode = aiModeInputAPIKey
						return "", nil
					case "reset_ollama":
						s.Config = ai.DefaultConfig()
						s.Reset(s.Config)
						return "ai_save_config", nil
					case "back":
						return "main_menu", nil
					}
				}
			case "esc", "q":
				return "main_menu", nil
			}

		case aiModeInputBaseURL:
			switch msg.String() {
			case "enter":
				s.Config.BaseURL = strings.TrimSpace(s.InputVal)
				s.Reset(s.Config)
				return "ai_save_config", nil
			case "esc":
				s.Mode = aiModeMain
				return "", nil
			case "backspace":
				if len(s.InputVal) > 0 {
					s.InputVal = s.InputVal[:len(s.InputVal)-1]
				}
			default:
				if len(msg.String()) == 1 {
					s.InputVal += msg.String()
				}
			}

		case aiModeInputModel:
			switch msg.String() {
			case "enter":
				s.Config.Model = strings.TrimSpace(s.InputVal)
				s.Reset(s.Config)
				return "ai_save_config", nil
			case "esc":
				s.Mode = aiModeMain
				return "", nil
			case "backspace":
				if len(s.InputVal) > 0 {
					s.InputVal = s.InputVal[:len(s.InputVal)-1]
				}
			default:
				if len(msg.String()) == 1 {
					s.InputVal += msg.String()
				}
			}

		case aiModeInputAPIKey:
			switch msg.String() {
			case "enter":
				s.Config.APIKey = strings.TrimSpace(s.InputVal)
				s.Reset(s.Config)
				return "ai_save_config", nil
			case "esc":
				s.Mode = aiModeMain
				return "", nil
			case "backspace":
				if len(s.InputVal) > 0 {
					s.InputVal = s.InputVal[:len(s.InputVal)-1]
				}
			default:
				if len(msg.String()) == 1 {
					s.InputVal += msg.String()
				}
			}

		case aiModeTesting:
			// Waiting for test command response

		case aiModeTestResult:
			switch msg.String() {
			case "enter", "esc", "q":
				s.Mode = aiModeMain
				return "", nil
			}
		}
	}
	return "", nil
}

// Render renders the AI settings screen.
func (s *AIConfigScreen) Render(repo *git.RepoInfo, status *git.StatusInfo, width int) string {
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	headerBox := s.Header.Render(repo, status, width)

	switch s.Mode {
	case aiModeInputBaseURL:
		title := s.Theme.SectionTitle.Render("Change AI Base URL")
		prompt := s.Theme.Label.Render("Enter OpenAI-compatible base URL:")
		inputBox := s.Theme.Value.Bold(true).Render("> " + s.InputVal + "█")
		hint := s.Theme.Label.Render("Example: http://localhost:11434/v1 or https://api.openai.com/v1")
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "Enter", Label: "Save"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", prompt, inputBox, "", hint, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case aiModeInputModel:
		title := s.Theme.SectionTitle.Render("Change AI Model")
		prompt := s.Theme.Label.Render("Enter model name:")
		inputBox := s.Theme.Value.Bold(true).Render("> " + s.InputVal + "█")
		hint := s.Theme.Label.Render("Example: qwen2.5-coder:7b or gpt-4o-mini")
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "Enter", Label: "Save"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", prompt, inputBox, "", hint, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case aiModeInputAPIKey:
		title := s.Theme.SectionTitle.Render("Change AI API Key")
		prompt := s.Theme.Label.Render("Enter API Key (use 'ollama' for local):")
		masked := strings.Repeat("*", len(s.InputVal))
		if len(s.InputVal) <= 4 {
			masked = s.InputVal
		}
		inputBox := s.Theme.Value.Bold(true).Render("> " + masked + "█")
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "Enter", Label: "Save"},
			{Key: "Esc", Label: "Cancel"},
		}, width)
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", prompt, inputBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case aiModeTesting:
		title := s.Theme.SectionTitle.Render("Testing AI Connection...")
		loading := s.Theme.AIText.Render("◌ Connecting to " + s.Config.BaseURL + "...")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", loading)
		return s.Theme.Card.Width(boxWidth).Render(content)

	case aiModeTestResult:
		var title, msg string
		if s.TestSuccess {
			title = s.Theme.SuccessText.Bold(true).Render("✓ AI Connection Successful")
			msg = s.Theme.Value.Render("Connected and responsive!\nModel: " + s.Config.Model)
		} else {
			title = s.Theme.ErrorText.Bold(true).Render("✕ AI Connection Failed")
			msg = s.Theme.Label.Render("Error details:\n  " + s.TestMsg)
		}
		hint := s.Theme.Label.Render("Press Enter to continue")
		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", msg, "", hint)
		return s.Theme.Card.Width(boxWidth).Render(content)

	default: // aiModeMain
		title := s.Theme.SectionTitle.Render("AI Settings")

		var statusStr string
		if s.Config.Enabled {
			statusStr = s.Theme.SuccessText.Render("● Enabled")
		} else {
			statusStr = s.Theme.Label.Render("○ Disabled")
		}

		keyMasked := strings.Repeat("*", len(s.Config.APIKey))
		if s.Config.APIKey == "ollama" || len(s.Config.APIKey) <= 4 {
			keyMasked = s.Config.APIKey
		}

		var infoLines []string
		infoLines = append(infoLines, fmt.Sprintf("%s  %s", s.Theme.Label.Render("Status:  "), statusStr))
		infoLines = append(infoLines, fmt.Sprintf("%s  %s", s.Theme.Label.Render("Base URL:"), s.Theme.Value.Render(s.Config.BaseURL)))
		infoLines = append(infoLines, fmt.Sprintf("%s  %s", s.Theme.Label.Render("Model:   "), s.Theme.Value.Render(s.Config.Model)))
		infoLines = append(infoLines, fmt.Sprintf("%s  %s", s.Theme.Label.Render("API Key: "), s.Theme.Label.Render(keyMasked)))
		infoBox := strings.Join(infoLines, "\n")

		menuBox := s.Menu.Render(width)
		footer := s.Footer.Render([]components.KeyHelp{
			{Key: "↑↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Back"},
		}, width)

		content := lipgloss.JoinVertical(lipgloss.Left, headerBox, "", title, "", infoBox, "", menuBox, "", footer)
		return s.Theme.Card.Width(boxWidth).Render(content)
	}
}
