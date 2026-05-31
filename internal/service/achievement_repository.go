package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const achievementFileVersion = 1

const achievementFileName = "achievements.json"

type achievementFile struct {
	Version int                          `json:"version"`
	Unlocks []roleplay.AchievementUnlock `json:"unlocks"`
}

type achievementRepository struct {
	path string
}

func newAchievementRepository() (*achievementRepository, error) {
	configDir, err := appconf.GetConfigDir()
	if err != nil {
		return nil, err
	}
	return &achievementRepository{path: filepath.Join(configDir, achievementFileName)}, nil
}

func (r *achievementRepository) load() ([]roleplay.AchievementUnlock, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []roleplay.AchievementUnlock{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read achievements: %w", err)
	}

	var file achievementFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse achievements: %w", err)
	}
	if file.Unlocks == nil {
		return []roleplay.AchievementUnlock{}, nil
	}
	sortAchievementUnlocks(file.Unlocks)
	return file.Unlocks, nil
}

func (r *achievementRepository) list(gameID string) ([]roleplay.AchievementUnlock, error) {
	unlocks, err := r.load()
	if err != nil {
		return nil, err
	}
	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		return unlocks, nil
	}
	filtered := make([]roleplay.AchievementUnlock, 0, len(unlocks))
	for _, unlock := range unlocks {
		if unlock.GameID == gameID {
			filtered = append(filtered, unlock)
		}
	}
	return filtered, nil
}

func (r *achievementRepository) upsert(unlock roleplay.AchievementUnlock) (roleplay.AchievementUnlock, bool, error) {
	unlock.GameID = strings.TrimSpace(unlock.GameID)
	unlock.AchievementID = strings.TrimSpace(unlock.AchievementID)
	unlock.Title = strings.TrimSpace(unlock.Title)
	unlock.SessionID = strings.TrimSpace(unlock.SessionID)
	unlock.EndingID = strings.TrimSpace(unlock.EndingID)
	unlock.UnlockedAt = strings.TrimSpace(unlock.UnlockedAt)
	if unlock.GameID == "" || unlock.AchievementID == "" {
		return roleplay.AchievementUnlock{}, false, errors.New("achievement unlock requires game id and achievement id")
	}
	if unlock.Title == "" {
		return roleplay.AchievementUnlock{}, false, errors.New("achievement unlock title is required")
	}
	if unlock.UnlockedAt == "" {
		unlock.UnlockedAt = roleplay.NowISO()
	}

	unlocks, err := r.load()
	if err != nil {
		return roleplay.AchievementUnlock{}, false, err
	}
	for _, existing := range unlocks {
		if existing.GameID == unlock.GameID && existing.AchievementID == unlock.AchievementID {
			return existing, false, nil
		}
	}

	unlocks = append(unlocks, unlock)
	sortAchievementUnlocks(unlocks)
	if err := r.save(unlocks); err != nil {
		return roleplay.AchievementUnlock{}, false, err
	}
	return unlock, true, nil
}

func (r *achievementRepository) deleteByGame(gameID string) error {
	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		return nil
	}
	unlocks, err := r.load()
	if err != nil {
		return err
	}
	filtered := unlocks[:0]
	for _, unlock := range unlocks {
		if unlock.GameID != gameID {
			filtered = append(filtered, unlock)
		}
	}
	return r.save(filtered)
}

func (r *achievementRepository) save(unlocks []roleplay.AchievementUnlock) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("create achievements directory: %w", err)
	}
	sortAchievementUnlocks(unlocks)
	data, err := json.MarshalIndent(achievementFile{
		Version: achievementFileVersion,
		Unlocks: unlocks,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode achievements: %w", err)
	}
	tempPath := r.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("write achievements temp file: %w", err)
	}
	if err := os.Rename(tempPath, r.path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace achievements file: %w", err)
	}
	return nil
}

func sortAchievementUnlocks(unlocks []roleplay.AchievementUnlock) {
	sort.SliceStable(unlocks, func(i, j int) bool {
		if unlocks[i].GameID != unlocks[j].GameID {
			return unlocks[i].GameID < unlocks[j].GameID
		}
		return unlocks[i].AchievementID < unlocks[j].AchievementID
	})
}
