package main

import (
	"context"
	"errors"
	"log"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"namelesswatch/internal/service"
	storypack "namelesswatch/internal/storypack"
	"os"
	"path/filepath"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	config        *appconf.AppConfig
	configService *service.ConfigService
	gameService   *service.GameService
}

type StoryTemplateResult struct {
	Root    string   `json:"root"`
	Written []string `json:"written"`
}

func NewApp() *App {
	config, err := appconf.LoadConfig()
	if err != nil {
		log.Printf("failed to load config, using defaults: %v", err)
		config = appconf.DefaultConfig()
	}

	gameService := service.NewGameService(config)
	if err := gameService.LoadLibrary(); err != nil {
		log.Printf("failed to load game library: %v", err)
	}

	return &App{
		config:        config,
		configService: service.NewConfigService(config),
		gameService:   gameService,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

func (a *App) ImportGamePack(files map[string]string) (roleplay.ImportGameResult, error) {
	return a.gameService.ImportGamePack(files)
}

func (a *App) SelectStoryTemplateDirectory() (string, error) {
	if a.ctx == nil {
		return "", errors.New("app is not ready")
	}
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:                "选择模板生成位置",
		CanCreateDirectories: true,
	})
}

func (a *App) CreateStoryTemplate(parentPath string, folderName string, title string, initialScene string, force bool) (StoryTemplateResult, error) {
	parentPath = strings.TrimSpace(parentPath)
	if parentPath == "" {
		return StoryTemplateResult{}, errors.New("target path is required")
	}
	parentInfo, err := os.Stat(parentPath)
	if err != nil {
		return StoryTemplateResult{}, err
	}
	if !parentInfo.IsDir() {
		return StoryTemplateResult{}, errors.New("target path is not a directory")
	}

	targetPath := filepath.Join(parentPath, storypack.SafeFolderName(folderName))
	written, err := storypack.ScaffoldPack(targetPath, storypack.ScaffoldOptions{
		Title:        title,
		InitialScene: initialScene,
		Force:        force,
	})
	if err != nil {
		return StoryTemplateResult{}, err
	}

	absRoot, err := filepath.Abs(targetPath)
	if err != nil {
		return StoryTemplateResult{}, err
	}
	return StoryTemplateResult{
		Root:    absRoot,
		Written: written,
	}, nil
}

func (a *App) GetGames() ([]roleplay.LibraryGame, error) {
	return a.gameService.GetGames()
}

func (a *App) GetGame(gameID string) (roleplay.LibraryGame, error) {
	return a.gameService.GetGame(gameID)
}

func (a *App) CreateGame(game roleplay.LibraryGame) (roleplay.LibraryGame, error) {
	return a.gameService.CreateGame(game)
}

func (a *App) UpdateGame(gameID string, game roleplay.LibraryGame) (roleplay.LibraryGame, error) {
	return a.gameService.UpdateGame(gameID, game)
}

func (a *App) DeleteGame(gameID string) error {
	return a.gameService.DeleteGame(gameID)
}

func (a *App) StartGame(gameID string) (roleplay.GameTurnResult, error) {
	return a.gameService.StartGame(gameID)
}

func (a *App) SubmitChoice(sessionID string, choiceID string) (roleplay.GameTurnResult, error) {
	return a.gameService.SubmitChoice(sessionID, choiceID)
}

func (a *App) SubmitCustomChoice(sessionID string, reply string) (roleplay.GameTurnResult, error) {
	return a.gameService.SubmitCustomChoice(sessionID, reply)
}

func (a *App) GetSession(sessionID string) (roleplay.GameSession, error) {
	return a.gameService.GetSession(sessionID)
}

func (a *App) ListSessions(gameID string) ([]service.SessionSummary, error) {
	return a.gameService.ListSessions(gameID)
}

func (a *App) ResumeSession(sessionID string) (roleplay.GameTurnResult, error) {
	return a.gameService.ResumeSession(sessionID)
}

func (a *App) SaveSnapshot(sessionID string, label string) (service.SessionSummary, error) {
	return a.gameService.SaveSnapshot(sessionID, label)
}

func (a *App) DeleteSession(sessionID string) error {
	return a.gameService.DeleteSession(sessionID)
}

func (a *App) ListUnlockedAchievements(gameID string) ([]roleplay.AchievementUnlock, error) {
	return a.gameService.ListUnlockedAchievements(gameID)
}
