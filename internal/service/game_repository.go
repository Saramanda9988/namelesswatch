package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
)

const gameLibraryVersion = 1

type gameLibraryFile struct {
	Version int                    `json:"version"`
	Games   []roleplay.LibraryGame `json:"games"`
}

type gameRepository struct {
	path string
}

func newGameRepository() (*gameRepository, error) {
	configDir, err := appconf.GetConfigDir()
	if err != nil {
		return nil, err
	}
	return &gameRepository{path: filepath.Join(configDir, "library.json")}, nil
}

func (r *gameRepository) load() ([]roleplay.LibraryGame, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []roleplay.LibraryGame{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read game library: %w", err)
	}

	var library gameLibraryFile
	if err := json.Unmarshal(data, &library); err != nil {
		return nil, fmt.Errorf("parse game library: %w", err)
	}
	if library.Games == nil {
		return []roleplay.LibraryGame{}, nil
	}
	return library.Games, nil
}

func (r *gameRepository) save(games []roleplay.LibraryGame) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("create game library directory: %w", err)
	}

	data, err := json.MarshalIndent(gameLibraryFile{
		Version: gameLibraryVersion,
		Games:   games,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode game library: %w", err)
	}

	tempPath := r.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("write game library temp file: %w", err)
	}
	if err := os.Rename(tempPath, r.path); err != nil {
		if removeErr := os.Remove(r.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			_ = os.Remove(tempPath)
			return fmt.Errorf("replace game library: %w", err)
		}
		if renameErr := os.Rename(tempPath, r.path); renameErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("rename game library temp file: %w", renameErr)
		}
	}
	return nil
}
