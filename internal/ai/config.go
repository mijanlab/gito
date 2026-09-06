package ai

// Config holds settings for AI commit message generation.
type Config struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

// DefaultConfig returns default sensible values for local Ollama / OpenAI-compatible provider.
func DefaultConfig() Config {
	return Config{
		Enabled:  true,
		Provider: "ollama",
		BaseURL:  "http://localhost:11434/v1",
		APIKey:   "ollama",
		Model:    "qwen2.5-coder:7b",
	}
}
