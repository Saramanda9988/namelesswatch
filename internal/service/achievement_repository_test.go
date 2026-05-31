package service

import (
	"namelesswatch/internal/roleplay"
	"path/filepath"
	"testing"
)

func TestAchievementRepositoryLoadUpsertListAndDelete(t *testing.T) {
	repo := &achievementRepository{path: filepath.Join(t.TempDir(), "achievements.json")}

	unlocks, err := repo.load()
	if err != nil {
		t.Fatalf("load empty repository: %v", err)
	}
	if len(unlocks) != 0 {
		t.Fatalf("expected empty repository, got %#v", unlocks)
	}

	first := roleplay.AchievementUnlock{
		GameID:        "game-a",
		AchievementID: "watch_revenge",
		Title:         "手表的复仇",
		SessionID:     "session-a",
		EndingID:      "watch_revenge",
		UnlockedAt:    "2026-01-01T00:00:00Z",
	}
	saved, newlyUnlocked, err := repo.upsert(first)
	if err != nil {
		t.Fatalf("upsert first unlock: %v", err)
	}
	if !newlyUnlocked || saved.UnlockedAt != first.UnlockedAt {
		t.Fatalf("expected first unlock to be new and preserved, saved=%#v new=%t", saved, newlyUnlocked)
	}

	duplicate := first
	duplicate.SessionID = "session-b"
	duplicate.UnlockedAt = "2026-02-01T00:00:00Z"
	saved, newlyUnlocked, err = repo.upsert(duplicate)
	if err != nil {
		t.Fatalf("upsert duplicate unlock: %v", err)
	}
	if newlyUnlocked {
		t.Fatal("duplicate unlock should not be marked new")
	}
	if saved.SessionID != first.SessionID || saved.UnlockedAt != first.UnlockedAt {
		t.Fatalf("duplicate should preserve first unlock, got %#v", saved)
	}

	second := roleplay.AchievementUnlock{
		GameID:        "game-b",
		AchievementID: "two_friends",
		Title:         "两个好朋友",
		SessionID:     "session-c",
	}
	if _, _, err := repo.upsert(second); err != nil {
		t.Fatalf("upsert second game unlock: %v", err)
	}

	gameUnlocks, err := repo.list("game-a")
	if err != nil {
		t.Fatalf("list game unlocks: %v", err)
	}
	if len(gameUnlocks) != 1 || gameUnlocks[0].AchievementID != first.AchievementID {
		t.Fatalf("expected one game-a unlock, got %#v", gameUnlocks)
	}

	if err := repo.deleteByGame("game-a"); err != nil {
		t.Fatalf("delete game unlocks: %v", err)
	}
	gameUnlocks, err = repo.list("game-a")
	if err != nil {
		t.Fatalf("list deleted game unlocks: %v", err)
	}
	if len(gameUnlocks) != 0 {
		t.Fatalf("expected game-a unlocks deleted, got %#v", gameUnlocks)
	}
	remaining, err := repo.list("")
	if err != nil {
		t.Fatalf("list remaining unlocks: %v", err)
	}
	if len(remaining) != 1 || remaining[0].GameID != "game-b" {
		t.Fatalf("expected other games to remain, got %#v", remaining)
	}
}
