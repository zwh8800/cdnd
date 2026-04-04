// Package config 提供 cdnd 的配置管理。
package config

import (
	"time"
)

// Config 表示应用程序的配置结构。
type Config struct {
	LLM      LLMConfig      `mapstructure:"llm"`
	Game     GameConfig     `mapstructure:"game"`
	Display  DisplayConfig  `mapstructure:"display"`
	Advanced AdvancedConfig `mapstructure:"advanced"`
}

// LLMConfig 包含 LLM 提供者的配置。
type LLMConfig struct {
	DefaultProvider string                    `mapstructure:"default_provider"`
	Providers       map[string]ProviderConfig `mapstructure:"providers"`
}

// ProviderConfig 表示单个 LLM 提供者的配置。
type ProviderConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	BaseURL     string  `mapstructure:"base_url"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// GameConfig 包含游戏相关的设置。
type GameConfig struct {
	Autosave         bool          `mapstructure:"autosave"`
	AutosaveInterval time.Duration `mapstructure:"autosave_interval"`
	MaxHistoryTurns  int           `mapstructure:"max_history_turns"`
	Language         string        `mapstructure:"language"`
}

// DisplayConfig 包含显示相关的设置。
type DisplayConfig struct {
	TypewriterEffect bool          `mapstructure:"typewriter_effect"`
	TypingSpeed      time.Duration `mapstructure:"typing_speed"`
	ColorOutput      bool          `mapstructure:"color_output"`
	ShowTokens       bool          `mapstructure:"show_tokens"`
}

// AdvancedConfig 包含高级设置。
type AdvancedConfig struct {
	CacheEnabled bool          `mapstructure:"cache_enabled"`
	CacheTTL     time.Duration `mapstructure:"cache_ttl"`
	LogLevel     string        `mapstructure:"log_level"`
	LogFile      string        `mapstructure:"log_file"`
}
