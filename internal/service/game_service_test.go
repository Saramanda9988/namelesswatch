package service

import (
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestNormalizeAndMaterializeLibraryGameBGMAssets(t *testing.T) {
	setTestUserConfigDir(t)

	files := testGameFiles(t)
	files["bgm/metadata.json"] = `{"tracks":{"home_ambient":{"name":"家中低频","file":"home.mp3"}}}`
	files["bgm/home.mp3"] = "data:audio/mpeg;base64,QUJD"

	game, pack, err := normalizeAndMaterializeLibraryGame(roleplay.LibraryGame{
		ID:    "bgm-game",
		Files: files,
	})
	if err != nil {
		t.Fatalf("normalize and materialize: %v", err)
	}
	if len(game.BGMs) != 1 || len(pack.BGMs) != 1 {
		t.Fatalf("expected materialized BGM in game and pack, game=%#v pack=%#v", game.BGMs, pack.BGMs)
	}
	if !strings.HasPrefix(game.BGMs[0].URL, "/local/story-assets/bgm-game/bgm/home.mp3") {
		t.Fatalf("unexpected BGM URL: %s", game.BGMs[0].URL)
	}
	if game.Files["bgm/home.mp3"] != game.BGMs[0].URL {
		t.Fatalf("expected game file to be replaced with local URL, file=%q bgm=%q", game.Files["bgm/home.mp3"], game.BGMs[0].URL)
	}

	assetRoot, err := appconf.GetSubDir(storyAssetsDirName)
	if err != nil {
		t.Fatalf("get story assets dir: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(assetRoot, "bgm-game", "bgm", "home.mp3"))
	if err != nil {
		t.Fatalf("read materialized BGM: %v", err)
	}
	if string(data) != "ABC" {
		t.Fatalf("expected decoded audio bytes, got %q", string(data))
	}
}

func testLibraryGame(t *testing.T, id string) roleplay.LibraryGame {
	t.Helper()

	return roleplay.LibraryGame{
		ID:    id,
		Files: testGameFiles(t),
	}
}

func setTestUserConfigDir(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", root)
	case "darwin":
		t.Setenv("HOME", root)
	default:
		t.Setenv("XDG_CONFIG_HOME", root)
	}
}
