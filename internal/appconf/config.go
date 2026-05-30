package appconf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAIProvider = "openai"
	DefaultAIBaseURL  = "https://api.openai.com/v1"
)

type AppConfig struct {
	AIProvider string `json:"ai_provider"`
	AIBaseURL  string `json:"ai_base_url"`
	AIModel    string `json:"ai_model"`
	AIToken    string `json:"ai_token,omitempty"`
}

func DefaultConfig() *AppConfig {
	baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	if baseURL == "" {
		baseURL = DefaultAIBaseURL
	}

	config := &AppConfig{
		AIProvider: strings.TrimSpace(os.Getenv("AI_PROVIDER")),
		AIBaseURL:  baseURL,
		AIModel:    strings.TrimSpace(os.Getenv("OPENAI_MODEL")),
		AIToken:    strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
	}
	Normalize(config)
	return config
}

func LoadConfig() (*AppConfig, error) {
	config := DefaultConfig()

	configPath, err := getConfigPath()
	if err != nil {
		return config, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := SaveConfig(config); err != nil {
			return config, err
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return config, fmt.Errorf("parse config: %w", err)
	}
	Normalize(config)
	return config, nil
}

func SaveConfig(config *AppConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	Normalize(config)

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	return os.WriteFile(configPath, data, 0o600)
}

func Normalize(config *AppConfig) {
	config.AIProvider = strings.TrimSpace(config.AIProvider)
	if config.AIProvider == "" {
		config.AIProvider = DefaultAIProvider
	}
	config.AIBaseURL = strings.TrimRight(strings.TrimSpace(config.AIBaseURL), "/")
	if config.AIBaseURL == "" {
		config.AIBaseURL = DefaultAIBaseURL
	}
	config.AIModel = strings.TrimSpace(config.AIModel)
	config.AIToken = strings.TrimSpace(config.AIToken)
}

func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config directory: %w", err)
	}
	return filepath.Join(configDir, "namelesswatch"), nil
}

func GetDataDir() (string, error) {
	dataDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("create data directory: %w", err)
	}
	return dataDir, nil
}

func GetSubDir(subPath string) (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(dataDir, subPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create sub directory %q: %w", subPath, err)
	}
	return dir, nil
}

func getConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "appconf.json"), nil
}
