package roleplay

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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
