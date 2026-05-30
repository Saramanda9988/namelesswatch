package roleplay

import (
	"context"
	"errors"
	"fmt"
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

type failingChatClient struct {
	err error
}

func (f failingChatClient) Chat(_ context.Context, _ []ChatMessage) (string, error) {
	return "", f.err
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

func TestNewLibraryGameRecognizesPlayerBriefing(t *testing.T) {
	pack := loadExamplePack(t)
	files := map[string]string{
		"metadata.json":        `{"title":"带开局规则的规则怪谈"}`,
		PlayerBriefingFileName: `{"title":"你需要记住的规则","items":[{"id":"feed-dog","text":"记得给狗喂食"}]}`,
	}
	for name, content := range pack.Files {
		files[name] = content
	}

	game, report, err := NewLibraryGame(files)
	if err != nil {
		t.Fatalf("new library game: %v", err)
	}
	if report.Game == nil || game.Files[PlayerBriefingFileName] == "" {
		t.Fatalf("expected imported game with briefing, got game=%#v report=%#v", game, report)
	}
	if !slices.Contains(report.ValidFiles, PlayerBriefingFileName) {
		t.Fatalf("expected briefing valid file, got %#v", report.ValidFiles)
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

func TestNewLibraryGameParsesBGMTracks(t *testing.T) {
	pack := loadExamplePack(t)
	files := map[string]string{
		"metadata.json":     `{"title":"带 BGM 的规则怪谈"}`,
		"bgm/metadata.json": `{"tracks":{"home_ambient":{"name":"家中低频","file":"home.mp3"},"missing":{"name":"缺失","file":"missing.mp3"},"note":{"name":"文本","file":"note.txt"}},"sceneDefaults":{"entrance":"home_ambient","kitchen":"missing"}}`,
		"bgm/home.mp3":      "data:audio/mpeg;base64,QUJD",
		"bgm/note.txt":      "not audio",
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
	if len(game.BGMs) != 1 {
		t.Fatalf("expected only valid BGM track, got %#v", game.BGMs)
	}
	if game.BGMs[0].ID != "home_ambient" || game.BGMs[0].Name != "家中低频" || game.BGMs[0].URL == "" {
		t.Fatalf("unexpected BGM asset: %#v", game.BGMs[0])
	}

	storyPack, err := NewStoryPack(game.ID, game.Files)
	if err != nil {
		t.Fatalf("new story pack: %v", err)
	}
	if len(storyPack.BGMs) != 1 {
		t.Fatalf("expected story pack BGM, got %#v", storyPack.BGMs)
	}
	if storyPack.BGMSceneDefaults["entrance"] != "home_ambient" {
		t.Fatalf("expected valid scene default, got %#v", storyPack.BGMSceneDefaults)
	}
	if _, ok := storyPack.BGMSceneDefaults["kitchen"]; ok {
		t.Fatalf("invalid scene default should be ignored: %#v", storyPack.BGMSceneDefaults)
	}
}

func TestNewLibraryGameAllowsNoBGM(t *testing.T) {
	pack := loadExamplePack(t)
	files := map[string]string{
		"metadata.json": `{"title":"无 BGM 的规则怪谈"}`,
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
	if len(game.BGMs) != 0 {
		t.Fatalf("expected no BGM tracks, got %#v", game.BGMs)
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

func TestNewGameSessionInitializesContextSummary(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSessionInDir("demo-game", pack, t.TempDir())
	if err != nil {
		t.Fatalf("new game session in dir: %v", err)
	}

	summary, err := ReadContextSummary(session)
	if err != nil {
		t.Fatalf("read context summary: %v", err)
	}
	for _, heading := range []string{"## 当前阶段", "## 关键事实", "## 用户选择", "## 规则后果", "## 未解决线索", "## 结局倾向"} {
		if !strings.Contains(summary, heading) {
			t.Fatalf("expected summary heading %q, got:\n%s", heading, summary)
		}
	}
	if _, err := os.Stat(ContextSummaryPath(session)); err != nil {
		t.Fatalf("context summary file missing: %v", err)
	}
}

func TestReadContextSummaryTreatsMissingFileAsDefault(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSessionInDir("demo-game", pack, t.TempDir())
	if err != nil {
		t.Fatalf("new game session in dir: %v", err)
	}
	if err := os.Remove(ContextSummaryPath(session)); err != nil {
		t.Fatalf("remove context summary: %v", err)
	}

	summary, err := ReadContextSummary(session)
	if err != nil {
		t.Fatalf("missing context summary should be readable: %v", err)
	}
	if summary != DefaultContextSummary() {
		t.Fatalf("expected default summary, got:\n%s", summary)
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

func TestValidateBGMChange(t *testing.T) {
	pack := StoryPack{
		BGMs: []BGMAsset{
			{ID: "home_ambient", Name: "家中低频", FileName: "home.mp3", URL: "/local/story-assets/game/bgm/home.mp3"},
		},
	}

	if err := ValidateBGMChange(&BGMChange{Action: "play", ID: "home_ambient"}, pack); err != nil {
		t.Fatalf("valid BGM play rejected: %v", err)
	}
	if err := ValidateBGMChange(&BGMChange{Action: "stop"}, StoryPack{}); err != nil {
		t.Fatalf("BGM stop should be accepted without tracks: %v", err)
	}
	if err := ValidateBGMChange(&BGMChange{Action: "play", ID: "missing"}, pack); err == nil {
		t.Fatal("expected unknown BGM id to be rejected")
	}
	if err := ValidateBGMChange(&BGMChange{Action: "fade", ID: "home_ambient"}, pack); err == nil {
		t.Fatal("expected invalid BGM action to be rejected")
	}
}

func TestAppendAITurnUpdatesCurrentBGM(t *testing.T) {
	session := &GameSession{
		ID:        "session-a",
		GameID:    "game-a",
		State:     SessionStatePlaying,
		Turns:     []GameTurn{},
		CreatedAt: NowISO(),
		UpdatedAt: NowISO(),
	}

	result := appendAITurn(session, AITurnResponse{
		Type:    "game_turn",
		State:   "continue",
		Payload: []string{"你听见墙内传来低频震动。"},
		BGM:     &BGMChange{Action: "play", ID: "home_ambient"},
		Tools: []ChoiceTool{{
			Type:    "choice",
			ID:      "main",
			Options: []ChoiceOption{{ID: "listen", Label: "继续听"}},
		}},
	})
	if session.CurrentBGMID != "home_ambient" || result.CurrentBGMID != "home_ambient" {
		t.Fatalf("expected current BGM to update, session=%q result=%q", session.CurrentBGMID, result.CurrentBGMID)
	}
	if result.Turn.BGM == nil || result.Turn.BGM.ID != "home_ambient" {
		t.Fatalf("expected BGM change on turn, got %#v", result.Turn.BGM)
	}

	result = appendAITurn(session, AITurnResponse{
		Type:    "game_turn",
		State:   "continue",
		Payload: []string{"那声音仍然贴着地板流动。"},
		Tools: []ChoiceTool{{
			Type:    "choice",
			ID:      "main",
			Options: []ChoiceOption{{ID: "step", Label: "向前一步"}},
		}},
	})
	if session.CurrentBGMID != "home_ambient" || result.CurrentBGMID != "home_ambient" {
		t.Fatalf("expected missing BGM field to keep current BGM, session=%q result=%q", session.CurrentBGMID, result.CurrentBGMID)
	}

	result = appendAITurn(session, AITurnResponse{
		Type:    "game_turn",
		State:   "continue",
		Payload: []string{"一切突然安静了。"},
		BGM:     &BGMChange{Action: "stop"},
		Tools: []ChoiceTool{{
			Type:    "choice",
			ID:      "main",
			Options: []ChoiceOption{{ID: "wait", Label: "站在原地"}},
		}},
	})
	if session.CurrentBGMID != "" || result.CurrentBGMID != "" {
		t.Fatalf("expected BGM stop to clear current BGM, session=%q result=%q", session.CurrentBGMID, result.CurrentBGMID)
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

func TestBuildMessagesMarksCustomInput(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.AppendTurn(GameTurn{
		ID:      NewID("turn"),
		Role:    TurnRoleAI,
		Payload: []string{"你站在玄关。"},
		Tools: []ChoiceTool{{
			Type:    "choice",
			ID:      "main",
			Options: []ChoiceOption{{ID: "check_watch", Label: "查看手表"}},
		}},
		CreatedAt: NowISO(),
	})
	session.AppendTurn(GameTurn{
		ID:                  NewID("turn"),
		Role:                TurnRoleUser,
		Payload:             []string{"敲三下门"},
		SelectedChoiceID:    "custom-test",
		SelectedChoiceLabel: "敲三下门",
		CustomInput:         true,
		CreatedAt:           NowISO(),
	})

	messages := BuildMessages(pack, session, nil, "")
	userPrompt := messages[1].Content
	if !strings.Contains(userPrompt, "(custom_input: 敲三下门)") {
		t.Fatalf("expected prompt to mark custom input, got:\n%s", userPrompt)
	}
	if strings.Contains(userPrompt, "(choice: custom-test") {
		t.Fatalf("custom input should not be formatted as a fixed option choice, got:\n%s", userPrompt)
	}
}

func TestBuildMessagesIncludesPlayerVisibleBriefing(t *testing.T) {
	pack := loadExamplePack(t)
	pack.Files[PlayerBriefingFileName] = `{"title":"你需要记住的规则","items":[{"id":"feed-dog","text":"记得给狗喂食"}]}`
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	messages := BuildMessages(pack, session, nil, "")
	userPrompt := messages[1].Content
	if !strings.Contains(userPrompt, "player-visible briefing.json") {
		t.Fatal("expected prompt to include player-visible briefing section")
	}
	if !strings.Contains(userPrompt, "记得给狗喂食") {
		t.Fatal("expected prompt to include briefing content")
	}
	if !strings.Contains(userPrompt, "属于用户已知信息") {
		t.Fatal("expected prompt to mark briefing as player-visible")
	}
}

func TestBuildMessagesIncludesContextSummaryAndLimitsRecentTurns(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := WriteContextSummary(session, "## 当前阶段\n- 已进入走廊\n\n## 关键事实\n- 手表响过三次"); err != nil {
		t.Fatalf("write context summary: %v", err)
	}
	for i := 0; i < 5; i++ {
		session.AppendTurn(GameTurn{
			ID:        NewID("turn"),
			Role:      TurnRoleAI,
			Payload:   []string{fmt.Sprintf("历史回合 %d", i)},
			CreatedAt: NowISO(),
		})
	}

	messages := BuildMessagesWithBudget(pack, session, nil, "", ContextBudget{RecentTurnLimit: 2})
	userPrompt := messages[1].Content
	if !strings.Contains(userPrompt, "--- context_summary.md ---") || !strings.Contains(userPrompt, "已进入走廊") {
		t.Fatalf("expected context summary in prompt, got:\n%s", userPrompt)
	}
	if strings.Contains(userPrompt, "历史回合 0") || strings.Contains(userPrompt, "历史回合 1") || strings.Contains(userPrompt, "历史回合 2") {
		t.Fatalf("expected old turns to be omitted from recent window, got:\n%s", userPrompt)
	}
	if !strings.Contains(userPrompt, "历史回合 3") || !strings.Contains(userPrompt, "历史回合 4") {
		t.Fatalf("expected newest turns to remain, got:\n%s", userPrompt)
	}
}

func TestBuildMessagesAppliesLowPriorityBudgets(t *testing.T) {
	pack := loadExamplePack(t)
	pack.Files["scene.md"] = "SCENE-LONG-CONTENT-123456789"
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := os.WriteFile(session.MemoryPath, []byte("memory-123456789"), 0o600); err != nil {
		t.Fatalf("write memory: %v", err)
	}
	if err := WriteContextSummary(session, "summary-123456789"); err != nil {
		t.Fatalf("write context summary: %v", err)
	}

	messages := BuildMessagesWithBudget(
		pack,
		session,
		[]TerminalExecution{{Command: "cmd", Stdout: "stdout-123456789", Stderr: "stderr-123456789"}},
		"repair-123456789",
		ContextBudget{
			RecentTurnLimit:        12,
			StoryFileRuneBudget:    8,
			SummaryRuneBudget:      8,
			MemoryRuneBudget:       8,
			TerminalResultRunes:    8,
			RepairInstructionRunes: 8,
		},
	)
	userPrompt := messages[1].Content
	for _, unexpected := range []string{"SCENE-LONG-CONTENT-123456789", "memory-123456789", "summary-123456789", "stdout-123456789", "stderr-123456789", "repair-123456789"} {
		if strings.Contains(userPrompt, unexpected) {
			t.Fatalf("expected %q to be truncated, got:\n%s", unexpected, userPrompt)
		}
	}
	if !strings.Contains(userPrompt, "[truncated]") {
		t.Fatalf("expected truncation marker, got:\n%s", userPrompt)
	}
}

func TestCompactSessionContextFailureKeepsExistingSummary(t *testing.T) {
	pack := loadExamplePack(t)
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	original := "## 当前阶段\n- 旧摘要仍然有效\n"
	if err := WriteContextSummary(session, original); err != nil {
		t.Fatalf("write context summary: %v", err)
	}
	for i := 0; i < 4; i++ {
		session.AppendTurn(GameTurn{
			ID:        NewID("turn"),
			Role:      TurnRoleAI,
			Payload:   []string{fmt.Sprintf("待压缩回合 %d", i)},
			CreatedAt: NowISO(),
		})
	}

	err = CompactSessionContext(context.Background(), failingChatClient{err: errors.New("model unavailable")}, session, ContextBudget{RecentTurnLimit: 1}, nil)
	if err == nil {
		t.Fatal("expected compaction error")
	}
	summary, readErr := ReadContextSummary(session)
	if readErr != nil {
		t.Fatalf("read context summary: %v", readErr)
	}
	if summary != original {
		t.Fatalf("expected old summary to remain, got:\n%s", summary)
	}
}

func TestBuildMessagesIncludesBGMContext(t *testing.T) {
	pack := loadExamplePack(t)
	pack.BGMs = []BGMAsset{
		{ID: "home_ambient", Name: "家中低频", FileName: "home.mp3", URL: "/local/story-assets/game/bgm/home.mp3"},
	}
	pack.BGMSceneDefaults = map[string]string{"entrance": "home_ambient"}
	session, err := NewGameSession("example", pack)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.CurrentBGMID = "home_ambient"

	messages := BuildMessages(pack, session, nil, "")
	if len(messages) != 2 {
		t.Fatalf("expected system and user messages, got %#v", messages)
	}
	userPrompt := messages[1].Content
	for _, expected := range []string{"Available BGM:", "home_ambient => 家中低频", "Current BGM:", "Scene Default BGM:", "entrance => home_ambient"} {
		if !strings.Contains(userPrompt, expected) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", expected, userPrompt)
		}
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
