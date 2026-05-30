package roleplay

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

type fakeChatClient struct {
	responses []string
	calls     int
}

func (f *fakeChatClient) Chat(_ context.Context, _ []ChatMessage) (string, error) {
	if f.calls >= len(f.responses) {
		return "", nil
	}
	response := f.responses[f.calls]
	f.calls++
	return response, nil
}

func loadExamplePack(t *testing.T) StoryPack {
	t.Helper()

	files := make(map[string]string)
	for _, fileName := range RequiredStoryFiles {
		content, err := os.ReadFile(filepath.Join("..", "..", "docs", "example", fileName))
		if err != nil {
			t.Fatalf("read example %s: %v", fileName, err)
		}
		files[fileName] = string(content)
	}

	pack, err := NewStoryPack("example", files)
	if err != nil {
		t.Fatalf("new story pack: %v", err)
	}
	return pack
}

func TestNewStoryPackRequiresStandardFiles(t *testing.T) {
	pack := loadExamplePack(t)
	if pack.Files["scene.md"] == "" || pack.Files["true.md"] == "" {
		t.Fatal("expected standard story files to be loaded")
	}

	delete(pack.Files, "true.md")
	if _, err := NewStoryPack("missing", pack.Files); err == nil || !strings.Contains(err.Error(), "true.md") {
		t.Fatalf("expected missing true.md error, got %v", err)
	}
}

func TestNewLibraryGameRequiresMetadataTitle(t *testing.T) {
	pack := loadExamplePack(t)
	files := map[string]string{
		"metadata.json": `{"title":"示例规则怪谈"}`,
	}
	for name, content := range pack.Files {
		files[name] = content
	}

	game, report, err := NewLibraryGame(files)
	if err != nil {
		t.Fatalf("new library game: %v", err)
	}
	if report.Game == nil || game.Title != "示例规则怪谈" {
		t.Fatalf("expected metadata title, got game=%#v report=%#v", game, report)
	}

	delete(files, "metadata.json")
	_, report, err = NewLibraryGame(files)
	if err != nil {
		t.Fatalf("missing metadata should be reported, not returned as error: %v", err)
	}
	if !slices.Contains(report.Missing, "metadata.json") {
		t.Fatalf("expected missing metadata.json, got %#v", report.Missing)
	}
}

func TestNewLibraryGameSupportsLegacyTTitle(t *testing.T) {
	pack := loadExamplePack(t)
	files := map[string]string{
		"metadata.json": `{"ttitle":"旧格式标题"}`,
	}
	for name, content := range pack.Files {
		files[name] = content
	}

	game, report, err := NewLibraryGame(files)
	if err != nil {
		t.Fatalf("new library game: %v", err)
	}
	if report.Game == nil || game.Title != "旧格式标题" {
		t.Fatalf("expected legacy metadata title, got game=%#v report=%#v", game, report)
	}
}

func TestNewLibraryGameParsesPhotoScenes(t *testing.T) {
	pack := loadExamplePack(t)
	files := map[string]string{
		"metadata.json":       `{"title":"带场景的规则怪谈"}`,
		"photo/metadata.json": `{"living_room":"客厅背景.png","kitchen":"厨房背景.png"}`,
		"photo/客厅背景.png":      "data:image/png;base64,AAA",
		"photo/厨房背景.png":      "data:image/png;base64,BBB",
	}
	for name, content := range pack.Files {
		files[name] = content
	}

	game, report, err := NewLibraryGame(files)
	if err != nil {
		t.Fatalf("new library game: %v", err)
	}
	if report.Game == nil {
		t.Fatalf("expected imported game, got report=%#v", report)
	}
	if len(game.Scenes) != 2 {
		t.Fatalf("expected 2 scenes, got %#v", game.Scenes)
	}
	if game.Scenes[0].ID == "" || game.Scenes[0].URL == "" {
		t.Fatalf("expected populated scene asset, got %#v", game.Scenes[0])
	}
	if len(game.PhotoURLs) != 2 {
		t.Fatalf("expected 2 photo urls, got %#v", game.PhotoURLs)
	}
}

func TestNewGameSessionInDirUsesGameAndSessionPath(t *testing.T) {
	pack := loadExamplePack(t)
	dir := t.TempDir()

	session, err := NewGameSessionInDir("demo-game", pack, dir)
	if err != nil {
		t.Fatalf("new game session in dir: %v", err)
	}
	if !strings.Contains(session.WorkspacePath, filepath.Join("demo-game", session.ID)) {
		t.Fatalf("unexpected workspace path: %s", session.WorkspacePath)
	}
	if session.CurrentSceneID == "" && len(pack.Scenes) > 0 {
		t.Fatal("expected current scene id to default from pack")
	}
}

func TestGameTurnValidationAndFallback(t *testing.T) {
	valid := AITurnResponse{
		Type:    "game_turn",
		State:   "continue",
		Payload: []string{"你听见手表又响了一次。"},
		Tools: []ChoiceTool{
			{
				Type:   "choice",
				ID:     "main",
				Prompt: "怎么做？",
				Options: []ChoiceOption{
					{ID: "check_watch", Label: "查看手表"},
					{ID: "ignore", Label: "先去厨房"},
				},
			},
		},
	}
	if err := ValidateGameTurn(valid); err != nil {
		t.Fatalf("valid game_turn rejected: %v", err)
	}

	if _, _, err := ParseAIResponse(`{"type":"game_turn","state":"continue","payload":[],"tools":[]}`); err != nil {
		t.Fatalf("parse should succeed before validation: %v", err)
	}
	if err := ValidateGameTurn(FallbackTurn()); err != nil {
		t.Fatalf("fallback turn should validate: %v", err)
	}
	if _, _, err := ParseAIResponse(`not-json`); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestRunAITurnRepairsInvalidResponse(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	client := &fakeChatClient{
		responses: []string{
			`not-json`,
			`{"type":"game_turn","state":"continue","payload":["你重新看向手表。"],"tools":[{"type":"choice","id":"main","options":[{"id":"check","label":"查看手表"},{"id":"door","label":"检查门口"}]}]}`,
		},
	}

	result, err := RunAITurn(context.Background(), client, pack, session)
	if err != nil {
		t.Fatalf("run ai turn: %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("expected repair retry call, got %d calls", client.calls)
	}
	if result.Payload[0] != "你重新看向手表。" {
		t.Fatalf("unexpected repaired payload: %#v", result.Payload)
	}
}

func TestRunAITurnRepairsFirstTurnEnding(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	client := &fakeChatClient{
		responses: []string{
			`{"type":"game_turn","state":"ended","payload":["你直接进入了循环。"],"tools":[],"ending":{"id":"loop","title":"循环结局","kind":"loop"}}`,
			`{"type":"game_turn","state":"continue","payload":["手表突然响了起来，你站在玄关前。"],"tools":[{"type":"choice","id":"main","options":[{"id":"check_watch","label":"查看手表"},{"id":"feed_dog","label":"先给狗喂食"}]}]}`,
		},
	}

	result, err := RunAITurn(context.Background(), client, pack, session)
	if err != nil {
		t.Fatalf("run ai turn: %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("expected repair retry call, got %d calls", client.calls)
	}
	if result.State == SessionStateEnded || result.Ending != nil {
		t.Fatalf("first turn should not end, got %#v", result)
	}
	if len(result.Tools) != 1 {
		t.Fatalf("expected choice tool after repair, got %#v", result.Tools)
	}
}

func TestBuildMessagesRequiresFreshRuleReviewAfterChoice(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	updatedRule := "每次推进前必须重新检查这条最新规则。"
	if err := os.WriteFile(filepath.Join(session.WorkspacePath, "rule.md"), []byte(updatedRule), 0o600); err != nil {
		t.Fatalf("write session rule: %v", err)
	}
	session.AppendTurn(GameTurn{
		ID:      NewID("turn"),
		Role:    TurnRoleAI,
		Payload: []string{"你站在玄关。"},
		Tools: []ChoiceTool{
			{
				Type: "choice",
				ID:   "main",
				Options: []ChoiceOption{
					{ID: "check_watch", Label: "查看手表"},
					{ID: "open_door", Label: "打开门"},
				},
			},
		},
	})
	session.AppendTurn(GameTurn{
		ID:                  NewID("turn"),
		Role:                TurnRoleUser,
		Payload:             []string{"查看手表"},
		SelectedChoiceID:    "check_watch",
		SelectedChoiceLabel: "查看手表",
	})

	messages := BuildMessages(pack, session, nil, "")
	if len(messages) != 2 {
		t.Fatalf("expected system and user messages, got %#v", messages)
	}
	systemPrompt := messages[0].Content
	userPrompt := messages[1].Content
	if !strings.Contains(systemPrompt, "Mandatory Rule Review After User Choice") {
		t.Fatal("expected system prompt to require rule review workflow")
	}
	if !strings.Contains(userPrompt, "Mandatory Rule Review After User Choice") {
		t.Fatal("expected user prompt to include mandatory rule review section")
	}
	if !strings.Contains(userPrompt, updatedRule) {
		t.Fatal("expected prompt to include fresh rule.md from session workspace")
	}
}

func TestRunAITurnFallsBackAfterFailedRepair(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	client := &fakeChatClient{
		responses: []string{`not-json`, `{"type":"game_turn","state":"continue","payload":[],"tools":[]}`},
	}

	result, err := RunAITurn(context.Background(), client, pack, session)
	if err != nil {
		t.Fatalf("run ai turn: %v", err)
	}
	if result.Payload[0] != FallbackTurn().Payload[0] {
		t.Fatalf("expected fallback payload, got %#v", result.Payload)
	}
}

func TestTerminalCanUpdateSessionMemoryCopy(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	originalMemory := pack.Files["memory.md"]
	command := "printf '\\nterminal-memory-updated\\n' >> memory.md"
	if runtime.GOOS == "windows" {
		command = "Add-Content -Path memory.md -Value 'terminal-memory-updated'"
	}

	result := ExecuteTerminalCommand(session.WorkspacePath, command)
	if result.ExitCode != 0 {
		t.Fatalf("terminal command failed: %#v", result)
	}

	sessionMemory, err := os.ReadFile(session.MemoryPath)
	if err != nil {
		t.Fatalf("read session memory: %v", err)
	}
	if !strings.Contains(string(sessionMemory), "terminal-memory-updated") {
		t.Fatal("expected session memory to be updated")
	}
	if pack.Files["memory.md"] != originalMemory {
		t.Fatal("original story pack memory was modified")
	}
}
