package main

import (
	"context"
	"log"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"namelesswatch/internal/service"
)

type App struct {
	config        *appconf.AppConfig
	configService *service.ConfigService
	gameService   *service.GameService
}

func NewApp() *App {
	config, err := appconf.LoadConfig()
	if err != nil {
		log.Printf("failed to load config, using defaults: %v", err)
		config = appconf.DefaultConfig()
	}

	return &App{
		config:        config,
		configService: service.NewConfigService(config),
		gameService:   service.NewGameService(config),
	}
}

func (a *App) startup(ctx context.Context) {
	a.configService.Init(ctx)
	a.gameService.Init(ctx)
}

func (a *App) GetAppConfig() (appconf.AppConfig, error) {
	return a.configService.GetAppConfig()
}

func (a *App) UpdateAppConfig(config appconf.AppConfig) error {
	appconf.Normalize(&config)
	if err := a.configService.UpdateAppConfig(config); err != nil {
		return err
	}
	a.gameService.SetConfig(config)
	return nil
}

func (a *App) RegisterGamePack(gameID string, files map[string]string) error {
	return a.gameService.RegisterGamePack(gameID, files)
}

func (a *App) StartGame(gameID string) (roleplay.GameTurnResult, error) {
	return a.gameService.StartGame(gameID)
}

func (a *App) SubmitChoice(sessionID string, choiceID string) (roleplay.GameTurnResult, error) {
	return a.gameService.SubmitChoice(sessionID, choiceID)
}

func (a *App) GetSession(sessionID string) (roleplay.GameSession, error) {
	return a.gameService.GetSession(sessionID)
}
