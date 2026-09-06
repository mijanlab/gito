package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCleanCommitMessage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "feat: add Stripe payment gateway",
			expected: "feat: add Stripe payment gateway",
		},
		{
			input:    `"Fix broken navigation bar in mobile view"`,
			expected: "Fix broken navigation bar in mobile view",
		},
		{
			input:    "```\nfeat: resolve memory leak in worker\n```",
			expected: "feat: resolve memory leak in worker",
		},
		{
			input:    "- Add unit test suite for git operations",
			expected: "Add unit test suite for git operations",
		},
		{
			input:    "Commit message: refactor diff viewer layout",
			expected: "refactor diff viewer layout",
		},
		{
			input:    "`fix(auth): handle expired token cleanly`",
			expected: "fix(auth): handle expired token cleanly",
		},
		{
			input:    "Update user authentication flow\n\nThis commit updates the login and signup routes.",
			expected: "Update user authentication flow",
		},
	}

	for _, tc := range tests {
		result := CleanCommitMessage(tc.input)
		if result != tc.expected {
			t.Errorf("CleanCommitMessage(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestOpenAICompatibleProvider(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid method", http.StatusBadRequest)
			return
		}

		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: "feat: add user authentication endpoint",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := Config{
		Enabled:  true,
		Provider: "openai_compatible",
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "test-model",
	}

	provider := NewOpenAICompatible(cfg)
	ctx := context.Background()

	// Test GenerateCommitMessage
	msg, err := provider.GenerateCommitMessage(ctx, "diff --git a/main.go b/main.go\n+func login() {}")
	if err != nil {
		t.Fatalf("GenerateCommitMessage failed: %v", err)
	}
	if msg != "feat: add user authentication endpoint" {
		t.Errorf("expected 'feat: add user authentication endpoint', got '%s'", msg)
	}

	// Test TestConnection
	err = provider.TestConnection(ctx)
	if err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}
