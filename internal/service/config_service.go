package service

import (
	"context"
	"fmt"
	"namelesswatch/internal/appconf"
	"sync"
)

type ConfigService struct {
	ctx    context.Context
	mu     sync.Mutex
	config *appconf.AppConfig
}

func NewConfigService(config *appconf.AppConfig) *ConfigService {
	return &ConfigService{config: config}
}

func (s *ConfigService) Init(ctx context.Context) {
	s.ctx = ctx
}

func (s *ConfigService) GetAppConfig() (appconf.AppConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return appconf.AppConfig{}, fmt.Errorf("config is not initialized")
	}
	return *s.config, nil
}

func (s *ConfigService) UpdateAppConfig(newConfig appconf.AppConfig) error {
	appconf.Normalize(&newConfig)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := appconf.SaveConfig(&newConfig); err != nil {
		return err
	}
	if s.config != nil {
		*s.config = newConfig
	}
	return nil
}
