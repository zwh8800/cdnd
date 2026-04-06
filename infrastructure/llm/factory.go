package llm

import (
	"fmt"

	llmd "github.com/zwh8800/cdnd/domain/llm"
	"github.com/zwh8800/cdnd/infrastructure/config"
)

// NewProvider 根据配置创建提供者。
// 它使用配置中设置的默认提供者。
func NewProvider(cfg *config.Config) (llmd.Provider, error) {
	if cfg == nil || cfg.LLM.DefaultProvider == "" {
		return nil, fmt.Errorf("no default provider configured")
	}

	providerCfg, exists := cfg.LLM.Providers[cfg.LLM.DefaultProvider]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in config", cfg.LLM.DefaultProvider)
	}

	// 将 config.ProviderConfig 转换为 llm.ProviderConfig
	llmCfg := llmd.ProviderConfig{
		APIKey:      providerCfg.APIKey,
		Model:       providerCfg.Model,
		BaseURL:     providerCfg.BaseURL,
		MaxTokens:   providerCfg.MaxTokens,
		Temperature: providerCfg.Temperature,
	}

	// 根据类型创建提供者
	switch cfg.LLM.DefaultProvider {
	case "openai":
		return NewOpenAIProvider(llmCfg), nil
	case "anthropic":
		return NewAnthropicProvider(llmCfg), nil
	case "ollama":
		return NewOllamaProvider(llmCfg), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.LLM.DefaultProvider)
	}
}

// NewProviderByName 根据名称和给定配置创建提供者。
func NewProviderByName(name string, cfg *config.Config) (Provider, error) {
	providerCfg, exists := cfg.LLM.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in config", name)
	}

	llmCfg := llmd.ProviderConfig{
		APIKey:      providerCfg.APIKey,
		Model:       providerCfg.Model,
		BaseURL:     providerCfg.BaseURL,
		MaxTokens:   providerCfg.MaxTokens,
		Temperature: providerCfg.Temperature,
	}

	switch name {
	case "openai":
		return NewOpenAIProvider(llmCfg), nil
	case "anthropic":
		return NewAnthropicProvider(llmCfg), nil
	case "ollama":
		return NewOllamaProvider(llmCfg), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
