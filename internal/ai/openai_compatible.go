package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatible implements Provider for any OpenAI-compatible API endpoint.
type OpenAICompatible struct {
	cfg        Config
	httpClient *http.Client
}

// NewOpenAICompatible creates a new provider instance.
func NewOpenAICompatible(cfg Config) *OpenAICompatible {
	return &OpenAICompatible{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

const systemPrompt = `You are an expert Git commit message generator.
Your instructions:
- Analyze ONLY the supplied git diff.
- Never invent changes not present in the diff.
- Generate EXACTLY ONE concise commit message on a single line.
- Use the imperative mood (e.g., "Add Stripe payment integration" or "fix: resolve navbar overflow").
- Do NOT wrap in markdown code blocks or quotes.
- Do NOT include any explanations, greetings, or alternative options.
- Keep the message under 72 characters if possible.`

// GenerateCommitMessage sends the diff to the OpenAI-compatible endpoint and returns a clean commit message.
func (p *OpenAICompatible) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	if !p.cfg.Enabled {
		return "", fmt.Errorf("AI is currently disabled in settings")
	}

	diff = strings.TrimSpace(diff)
	if diff == "" {
		return "", fmt.Errorf("diff is empty, cannot generate commit message")
	}

	// Limit diff size to ~12KB to avoid exceeding context limits on smaller local LLMs
	if len(diff) > 12000 {
		diff = diff[:12000] + "\n\n... (diff truncated)"
	}

	endpoint := strings.TrimRight(p.cfg.BaseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}

	reqBody := chatRequest{
		Model: p.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Here is the git diff:\n\n%s", diff)},
		},
		Temperature: 0.2,
		MaxTokens:   128,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to AI provider (%s): %w", p.cfg.BaseURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI provider returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w", err)
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("AI provider error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no commit message returned by AI provider")
	}

	rawMsg := chatResp.Choices[0].Message.Content
	cleanMsg := CleanCommitMessage(rawMsg)
	if cleanMsg == "" {
		return "", fmt.Errorf("AI generated an empty commit message")
	}

	return cleanMsg, nil
}

// TestConnection verifies that the API endpoint is reachable and responsive.
func (p *OpenAICompatible) TestConnection(ctx context.Context) error {
	endpoint := strings.TrimRight(p.cfg.BaseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}

	reqBody := chatRequest{
		Model: p.cfg.Model,
		Messages: []chatMessage{
			{Role: "user", Content: "Respond with the single word: OK"},
		},
		Temperature: 0.1,
		MaxTokens:   10,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	return nil
}

// CleanCommitMessage cleans model output to ensure a clean, single-line commit message.
func CleanCommitMessage(msg string) string {
	msg = strings.TrimSpace(msg)

	// Remove markdown codeblock backticks ``` ... ```
	if strings.HasPrefix(msg, "```") {
		lines := strings.Split(msg, "\n")
		var filtered []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if !strings.HasPrefix(trimmed, "```") {
				filtered = append(filtered, l)
			}
		}
		msg = strings.TrimSpace(strings.Join(filtered, "\n"))
	}

	// Remove single backticks around message
	msg = strings.Trim(msg, "`")

	// Split into lines, take first non-empty line
	lines := strings.Split(msg, "\n")
	firstLine := ""
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			firstLine = trimmed
			break
		}
	}

	// Strip common prefixes like "- ", "* ", "Commit message: ", etc.
	firstLine = strings.TrimPrefix(firstLine, "- ")
	firstLine = strings.TrimPrefix(firstLine, "* ")
	firstLine = strings.TrimPrefix(firstLine, "• ")
	if idx := strings.Index(strings.ToLower(firstLine), "commit message:"); idx != -1 {
		firstLine = strings.TrimSpace(firstLine[idx+len("commit message:"):])
	}

	// Strip wrapping quotation marks
	firstLine = strings.Trim(firstLine, `"'`)

	return strings.TrimSpace(firstLine)
}
