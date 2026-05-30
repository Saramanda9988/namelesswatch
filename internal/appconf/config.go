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

	DefaultAIContextRecentTurns    = 12
	DefaultAIContextCompactTurns   = 24
	DefaultAIContextSoftBudget     = 60000
	DefaultAIContextHardBudget     = 120000
	DefaultAIChoicePrefetchEnabled = false
	DefaultAIChoicePrefetchGlobal  = 2
	DefaultAIChoicePrefetchSession = 2
	DefaultAIChoicePrefetchTTLMS   = 120000
	DefaultAIChoicePrefetchWaitMS  = 1200
)

type AppConfig struct {
	AIProvider                         string `json:"ai_provider"`
	AIBaseURL                          string `json:"ai_base_url"`
	AIModel                            string `json:"ai_model"`
	AIToken                            string `json:"ai_token,omitempty"`
	AIContextRecentTurns               int    `json:"ai_context_recent_turns"`
	AIContextCompactTurns              int    `json:"ai_context_compact_turns"`
	AIContextSoftBudget                int    `json:"ai_context_soft_budget"`
	AIContextHardBudget                int    `json:"ai_context_hard_budget"`
	AIChoicePrefetchEnabled            bool   `json:"ai_choice_prefetch_enabled"`
	AIChoicePrefetchGlobalConcurrency  int    `json:"ai_choice_prefetch_global_concurrency"`
	AIChoicePrefetchSessionConcurrency int    `json:"ai_choice_prefetch_session_concurrency"`
	AIChoicePrefetchTTLMS              int    `json:"ai_choice_prefetch_ttl_ms"`
	AIChoicePrefetchWaitMS             int    `json:"ai_choice_prefetch_wait_ms"`
}

func DefaultConfig() *AppConfig {
	baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	if baseURL == "" {
		baseURL = DefaultAIBaseURL
	}

	config := &AppConfig{
		AIProvider:                         strings.TrimSpace(os.Getenv("AI_PROVIDER")),
		AIBaseURL:                          baseURL,
		AIModel:                            strings.TrimSpace(os.Getenv("OPENAI_MODEL")),
		AIToken:                            strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		AIContextRecentTurns:               DefaultAIContextRecentTurns,
		AIContextCompactTurns:              DefaultAIContextCompactTurns,
		AIContextSoftBudget:                DefaultAIContextSoftBudget,
		AIContextHardBudget:                DefaultAIContextHardBudget,
		AIChoicePrefetchEnabled:            DefaultAIChoicePrefetchEnabled,
		AIChoicePrefetchGlobalConcurrency:  DefaultAIChoicePrefetchGlobal,
		AIChoicePrefetchSessionConcurrency: DefaultAIChoicePrefetchSession,
		AIChoicePrefetchTTLMS:              DefaultAIChoicePrefetchTTLMS,
		AIChoicePrefetchWaitMS:             DefaultAIChoicePrefetchWaitMS,
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
	config.AIContextRecentTurns = normalizeInt(config.AIContextRecentTurns, DefaultAIContextRecentTurns, 4, 24)
	config.AIContextCompactTurns = normalizeInt(config.AIContextCompactTurns, DefaultAIContextCompactTurns, config.AIContextRecentTurns+1, 96)
	config.AIContextSoftBudget = normalizeInt(config.AIContextSoftBudget, DefaultAIContextSoftBudget, 12000, 300000)
	config.AIContextHardBudget = normalizeInt(config.AIContextHardBudget, DefaultAIContextHardBudget, config.AIContextSoftBudget, 500000)
	config.AIChoicePrefetchGlobalConcurrency = normalizeInt(config.AIChoicePrefetchGlobalConcurrency, DefaultAIChoicePrefetchGlobal, 1, 8)
	config.AIChoicePrefetchSessionConcurrency = normalizeInt(config.AIChoicePrefetchSessionConcurrency, DefaultAIChoicePrefetchSession, 1, config.AIChoicePrefetchGlobalConcurrency)
	config.AIChoicePrefetchTTLMS = normalizeInt(config.AIChoicePrefetchTTLMS, DefaultAIChoicePrefetchTTLMS, 10000, 600000)
	config.AIChoicePrefetchWaitMS = normalizeInt(config.AIChoicePrefetchWaitMS, DefaultAIChoicePrefetchWaitMS, 0, 10000)
}

func normalizeInt(value, fallback, minValue, maxValue int) int {
	if value == 0 {
		value = fallback
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
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
