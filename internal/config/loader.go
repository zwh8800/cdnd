package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	// 全局配置实例
	cfg *Config

	// 配置文件名
	configName = "config"

	// 配置文件类型
	configType = "yaml"

	// 配置目录名
	configDir = ".cdnd"
)

// Init 初始化配置系统。
func Init() error {
	// 设置默认值
	cfg = DefaultConfig()

	// 配置 Viper
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, configDir)

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)

	// 启用环境变量覆盖
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults(cfg)

	// 如果配置目录不存在则创建
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	// 如果配置文件存在则读取
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		// 配置文件未找到，将使用默认值
	}

	// 反序列化配置
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	return nil
}

// setDefaults 在 Viper 中设置默认值。
func setDefaults(c *Config) {
	viper.SetDefault("llm.default_provider", c.LLM.DefaultProvider)

	for name, provider := range c.LLM.Providers {
		viper.SetDefault("llm.providers."+name+".model", provider.Model)
		viper.SetDefault("llm.providers."+name+".base_url", provider.BaseURL)
		viper.SetDefault("llm.providers."+name+".max_tokens", provider.MaxTokens)
		viper.SetDefault("llm.providers."+name+".temperature", provider.Temperature)
	}

	viper.SetDefault("game.autosave", c.Game.Autosave)
	viper.SetDefault("game.autosave_interval", c.Game.AutosaveInterval.String())
	viper.SetDefault("game.max_history_turns", c.Game.MaxHistoryTurns)
	viper.SetDefault("game.language", c.Game.Language)

	viper.SetDefault("display.typewriter_effect", c.Display.TypewriterEffect)
	viper.SetDefault("display.typing_speed", c.Display.TypingSpeed.String())
	viper.SetDefault("display.color_output", c.Display.ColorOutput)
	viper.SetDefault("display.show_tokens", c.Display.ShowTokens)

	viper.SetDefault("advanced.cache_enabled", c.Advanced.CacheEnabled)
	viper.SetDefault("advanced.cache_ttl", c.Advanced.CacheTTL.String())
	viper.SetDefault("advanced.log_level", c.Advanced.LogLevel)
}

// Get 返回全局配置。
func Get() *Config {
	return cfg
}

// Set 更新全局配置。
func Set(c *Config) {
	cfg = c
}

// Save 将配置保存到文件。
func Save() error {
	return viper.WriteConfig()
}

// SaveAs 将配置保存到指定文件。
func SaveAs(path string) error {
	return viper.WriteConfigAs(path)
}

// GetConfigPath 返回配置文件的路径。
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configName+"."+configType), nil
}

// GetDataDir 返回数据目录的路径。
func GetDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir), nil
}

// InitConfigFile 创建带有默认值的配置文件。
func InitConfigFile() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// 检查文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return nil // 文件已存在
	}

	// 创建配置文件
	return viper.SafeWriteConfigAs(configPath)
}
