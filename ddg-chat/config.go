package ddgchat

import (
	"fmt"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type Config struct {
	Port          int               `toml:"port"`
	Host          string            `toml:"host"`
	UserAgent     string            `toml:"user_agent"`
	Tokens        []string          `toml:"tokens"`
	DDGChatAPIURL string            `toml:"ddg_chat_api_url"`
	ModelMapping  map[string]string `toml:"model_mapping"`
}

func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Error("config file does not exist", zap.String("config_path", configPath))
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	config := &Config{}

	if _, err := toml.DecodeFile(configPath, config); err != nil {
		logger.Error("failed to decode config file", zap.String("config_path", configPath), zap.Error(err))
		return nil, fmt.Errorf("failed to decode config file: %s", err)
	}

	if err := ValidateConfig(config); err != nil {
		logger.Error("invalid config", zap.Error(err))
		return nil, err
	}

	return config, nil
}

func ValidateConfig(config *Config) error {
	if config.Port < 1 || config.Port > 65535 {
		logger.Error("invalid port", zap.Int("port", config.Port))
		return fmt.Errorf("invalid port: %d", config.Port)
	}

	if matched, _ := regexp.MatchString("^[a-zA-Z0-9.-]+$", config.Host); !matched {
		logger.Error("invalid host", zap.String("host", config.Host))
		return fmt.Errorf("invalid host: %s", config.Host)
	}

	if config.DDGChatAPIURL == "" {
		logger.Error("DDG Chat API URL is required")
		return fmt.Errorf("DDG Chat API URL is required")
	}

	return nil
}
