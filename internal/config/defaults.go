package config

import (
	"time"
)

// DefaultConfig 返回默认配置。
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			DefaultProvider: "openai",
			Providers: map[string]ProviderConfig{
				"openai": {
					Model:       "gpt-4-turbo-preview",
					BaseURL:     "https://api.openai.com/v1",
					MaxTokens:   4096,
					Temperature: 0.7,
				},
				"anthropic": {
					Model:       "claude-3-opus-20240229",
					MaxTokens:   4096,
					Temperature: 0.7,
				},
				"ollama": {
					BaseURL:     "http://localhost:11434",
					Model:       "llama2",
					MaxTokens:   4096,
					Temperature: 0.7,
				},
			},
		},
		Game: GameConfig{
			Autosave:         true,
			AutosaveInterval: 5 * time.Minute,
			MaxHistoryTurns:  100,
			Language:         "zh-CN",
		},
		Display: DisplayConfig{
			TypewriterEffect: true,
			TypingSpeed:      50 * time.Millisecond,
			ColorOutput:      true,
			ShowTokens:       false,
		},
		Advanced: AdvancedConfig{
			CacheEnabled: true,
			CacheTTL:     24 * time.Hour,
			LogLevel:     "info",
			LogFile:      "",
		},
	}
}
