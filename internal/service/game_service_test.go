package service

import (
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"testing"
)

func testGameFiles(t *testing.T) map[string]string {
	t.Helper()

	files := map[string]string{
		"metadata.json": `{"ttitle":"持久化测试"}`,
	}
	for _, fileName := range []string{"scene.md", "rule.md", "true.md", "memory.md", "endings.md"} {
		content, err := os.ReadFile(filepath.Join("..", "..", "docs", "example", fileName))
		if err != nil {
			t.Fatalf("read example %s: %v", fileName, err)
		}
		files[fileName] = string(content)
	}
	return files
}

func TestGameServiceCRUDPersistsJSON(t *testing.T) {
	service := NewGameService(nil)
	service.repo = &gameRepository{path: filepath.Join(t.TempDir(), "library.json")}

	created, err := service.CreateGame(testLibraryGame(t, ""))
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	if created.ID == "" || created.Title != "持久化测试" {
		t.Fatalf("unexpected created game: %#v", created)
	}

	loadedGames, err := service.repo.load()
	if err != nil {
		t.Fatalf("load persisted games: %v", err)
	}
	if len(loadedGames) != 1 || loadedGames[0].ID != created.ID {
		t.Fatalf("expected persisted game, got %#v", loadedGames)
	}

	created.Title = "已更新标题"
	updated, err := service.UpdateGame(created.ID, created)
	if err != nil {
		t.Fatalf("update game: %v", err)
	}
	if updated.Title != "已更新标题" {
		t.Fatalf("expected updated title, got %#v", updated)
	}

	if _, err := service.GetGame(created.ID); err != nil {
		t.Fatalf("get game: %v", err)
	}

	if err := service.DeleteGame(created.ID); err != nil {
		t.Fatalf("delete game: %v", err)
	}
	games, err := service.GetGames()
	if err != nil {
		t.Fatalf("get games: %v", err)
	}
	if len(games) != 0 {
		t.Fatalf("expected empty games after delete, got %#v", games)
	}
}

func testLibraryGame(t *testing.T, id string) roleplay.LibraryGame {
	t.Helper()

	return roleplay.LibraryGame{
		ID:    id,
		Files: testGameFiles(t),
	}
}
