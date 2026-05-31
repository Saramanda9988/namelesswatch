package service

import (
	"context"
	"errors"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestGameServiceDeleteGameDeletesAchievementUnlocks(t *testing.T) {
	setTestUserConfigDir(t)
	service := NewGameService(nil)
	service.repo = &gameRepository{path: filepath.Join(t.TempDir(), "library.json")}
	service.achievementRepo = &achievementRepository{path: filepath.Join(t.TempDir(), "achievements.json")}

	created, err := service.CreateGame(testLibraryGame(t, ""))
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	if _, _, err := service.achievementRepo.upsert(roleplay.AchievementUnlock{
		GameID:        created.ID,
		AchievementID: "watch_revenge",
		Title:         "手表的复仇",
		SessionID:     "session-a",
	}); err != nil {
		t.Fatalf("upsert achievement: %v", err)
	}

	if err := service.DeleteGame(created.ID); err != nil {
		t.Fatalf("delete game: %v", err)
	}
	unlocks, err := service.achievementRepo.list(created.ID)
	if err != nil {
		t.Fatalf("list achievements: %v", err)
	}
	if len(unlocks) != 0 {
		t.Fatalf("expected deleted game achievements to be removed, got %#v", unlocks)
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

type scriptedChatFactory struct {
	mu      sync.Mutex
	clients []*scriptedChatClient
	calls   int
}

type scriptedChatClient struct {
	response string
	err      error
	delay    time.Duration
	started  chan struct{}
	once     sync.Once
}

func (f *scriptedChatFactory) factory(_ appconf.AppConfig) roleplay.ChatCompleter {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls >= len(f.clients) {
		f.calls++
		return &scriptedChatClient{response: endedTurnResponse("默认回合")}
	}
	client := f.clients[f.calls]
	f.calls++
	return client
}

func (f *scriptedChatFactory) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (c *scriptedChatClient) Chat(ctx context.Context, _ []roleplay.ChatMessage) (string, error) {
	if c.started != nil {
		c.once.Do(func() {
			close(c.started)
		})
	}
	if c.delay > 0 {
		timer := time.NewTimer(c.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timer.C:
		}
	}
	if c.err != nil {
		return "", c.err
	}
	return c.response, nil
}

func TestGameServiceSubmitChoiceUsesPrefetchHit(t *testing.T) {
	config := prefetchTestConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: endedTurnResponse("预生成命中")}},
	}
	service, session, pack := newPrefetchTestService(t, config, factory)
	baseTurn := session.Turns[len(session.Turns)-1]
	originalWorkspace := session.WorkspacePath

	service.maybeStartChoicePrefetch(session.Clone(), pack, config, roleplay.ResultFromSession(session))
	result, err := service.SubmitChoice(session.ID, "left")
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if got := strings.Join(result.Payload, ""); got != "预生成命中" {
		t.Fatalf("expected prefetch hit payload, got %#v", result.Payload)
	}
	if factory.callCount() != 1 {
		t.Fatalf("expected only prefetch model call, got %d", factory.callCount())
	}

	current := service.sessions[session.ID]
	if current.ID != session.ID || current.GameID != session.GameID || current.Label != session.Label || current.ParentID != session.ParentID {
		t.Fatalf("promoted branch changed external identity: %#v", current)
	}
	if len(current.Turns) != 3 || current.Turns[1].SelectedChoiceID != "left" || current.Turns[0].ID != baseTurn.ID {
		t.Fatalf("unexpected promoted turns: %#v", current.Turns)
	}
	if current.WorkspacePath == originalWorkspace || !strings.Contains(current.WorkspacePath, ".prefetch") {
		t.Fatalf("expected promoted branch workspace, got %q from %q", current.WorkspacePath, originalWorkspace)
	}
}

func TestGameServiceSubmitChoiceFallsBackWhenPrefetchMissing(t *testing.T) {
	config := prefetchTestConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: endedTurnResponse("同步回落")}},
	}
	service, session, _ := newPrefetchTestService(t, config, factory)

	result, err := service.SubmitChoice(session.ID, "left")
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if got := strings.Join(result.Payload, ""); got != "同步回落" {
		t.Fatalf("expected fallback payload, got %#v", result.Payload)
	}
	if factory.callCount() != 1 {
		t.Fatalf("expected one synchronous model call, got %d", factory.callCount())
	}
}

func TestGameServiceSubmitCustomChoiceRecordsUserInput(t *testing.T) {
	config := *appconf.DefaultConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: endedTurnResponse("自定义后果")}},
	}
	service, session, _ := newPrefetchTestService(t, config, factory)

	result, err := service.SubmitCustomChoice(session.ID, "  敲三下门  ")
	if err != nil {
		t.Fatalf("submit custom choice: %v", err)
	}
	if got := strings.Join(result.Payload, ""); got != "自定义后果" {
		t.Fatalf("expected custom payload, got %#v", result.Payload)
	}
	if factory.callCount() != 1 {
		t.Fatalf("expected one synchronous model call, got %d", factory.callCount())
	}

	current := service.sessions[session.ID]
	if len(current.Turns) != 3 {
		t.Fatalf("expected ai/user/ai turns, got %#v", current.Turns)
	}
	userTurn := current.Turns[1]
	if !userTurn.CustomInput || userTurn.SelectedChoiceLabel != "敲三下门" || strings.TrimSpace(userTurn.SelectedChoiceID) == "" {
		t.Fatalf("expected custom user turn, got %#v", userTurn)
	}
	if !strings.HasPrefix(userTurn.SelectedChoiceID, "custom-") {
		t.Fatalf("expected generated custom choice id, got %q", userTurn.SelectedChoiceID)
	}
}

func TestGameServiceSubmitCustomChoiceRecordsAchievementUnlock(t *testing.T) {
	config := *appconf.DefaultConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: achievementEndedResponse("手表从黑暗里回来了。", "watch_revenge", "手表的复仇", "bad")}},
	}
	service, session, _, repo := newAchievementTestService(t, config, factory, customInputAchievementJSON, "")

	result, err := service.SubmitCustomChoice(session.ID, "我把手表扔出窗外")
	if err != nil {
		t.Fatalf("submit custom choice: %v", err)
	}
	if result.Achievement == nil || result.Achievement.AchievementID != "watch_revenge" || !result.Achievement.New {
		t.Fatalf("expected new achievement unlock, got %#v", result.Achievement)
	}
	if result.Achievement.UnlockedAt == "" {
		t.Fatalf("expected unlock time, got %#v", result.Achievement)
	}

	unlocks, err := repo.list(session.GameID)
	if err != nil {
		t.Fatalf("list achievements: %v", err)
	}
	if len(unlocks) != 1 || unlocks[0].AchievementID != "watch_revenge" || unlocks[0].SessionID != session.ID {
		t.Fatalf("expected persisted unlock, got %#v", unlocks)
	}
	latest, ok := service.sessions[session.ID].LatestAITurn()
	if !ok || latest.Achievement == nil || latest.Achievement.ID != "watch_revenge" {
		t.Fatalf("expected session turn achievement, got %#v", latest)
	}
}

func TestGameServicePrefetchHitRecordsAchievementUnlock(t *testing.T) {
	config := prefetchTestConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: achievementEndedResponse("你和朋友一起被困进循环。", "two_friends", "两个好朋友", "loop")}},
	}
	service, session, pack, repo := newAchievementTestService(t, config, factory, choiceAchievementJSON, "")
	baseTurn := session.Turns[len(session.Turns)-1]

	service.maybeStartChoicePrefetch(session.Clone(), pack, config, roleplay.ResultFromSession(session))
	result, err := service.SubmitChoice(session.ID, baseTurn.Tools[0].Options[0].ID)
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if result.Achievement == nil || result.Achievement.AchievementID != "two_friends" || !result.Achievement.New {
		t.Fatalf("expected prefetch achievement unlock, got %#v", result.Achievement)
	}
	if factory.callCount() != 1 {
		t.Fatalf("expected only prefetch model call, got %d", factory.callCount())
	}
	unlocks, err := repo.list(session.GameID)
	if err != nil {
		t.Fatalf("list achievements: %v", err)
	}
	if len(unlocks) != 1 || unlocks[0].AchievementID != "two_friends" {
		t.Fatalf("expected persisted prefetch unlock, got %#v", unlocks)
	}
}

func TestGameServiceRuleBasedOneLifeUnlock(t *testing.T) {
	config := *appconf.DefaultConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: goodEndingResponse("你在朋友家安全熬到了天亮。")}},
	}
	service, session, _, repo := newAchievementTestService(t, config, factory, oneLifeAchievementJSON, "")

	result, err := service.SubmitChoice(session.ID, "left")
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if result.Achievement == nil || result.Achievement.AchievementID != "one_life_clear" || !result.Achievement.New {
		t.Fatalf("expected one-life achievement unlock, got %#v", result.Achievement)
	}
	latest, ok := service.sessions[session.ID].LatestAITurn()
	if !ok || latest.Achievement == nil || latest.Achievement.ID != "one_life_clear" {
		t.Fatalf("expected rule achievement on latest turn, got %#v", latest)
	}
	unlocks, err := repo.list(session.GameID)
	if err != nil {
		t.Fatalf("list achievements: %v", err)
	}
	if len(unlocks) != 1 || unlocks[0].AchievementID != "one_life_clear" {
		t.Fatalf("expected persisted one-life unlock, got %#v", unlocks)
	}
}

func TestGameServiceRuleBasedOneLifeSkipsSnapshotFork(t *testing.T) {
	config := *appconf.DefaultConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: goodEndingResponse("你在朋友家安全熬到了天亮。")}},
	}
	service, session, _, repo := newAchievementTestService(t, config, factory, oneLifeAchievementJSON, "snapshot-session")

	result, err := service.SubmitChoice(session.ID, "left")
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if result.Achievement != nil {
		t.Fatalf("snapshot fork should not unlock one-life achievement, got %#v", result.Achievement)
	}
	unlocks, err := repo.list(session.GameID)
	if err != nil {
		t.Fatalf("list achievements: %v", err)
	}
	if len(unlocks) != 0 {
		t.Fatalf("expected no persisted unlocks, got %#v", unlocks)
	}
}

func TestGameServiceSubmitChoiceFallsBackWhenPrefetchTimesOut(t *testing.T) {
	config := prefetchTestConfig()
	config.AIChoicePrefetchWaitMS = 10
	started := make(chan struct{})
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{
			{response: endedTurnResponse("迟到预生成"), delay: 200 * time.Millisecond, started: started},
			{response: endedTurnResponse("超时同步回落")},
		},
	}
	service, session, pack := newPrefetchTestService(t, config, factory)
	service.maybeStartChoicePrefetch(session.Clone(), pack, config, roleplay.ResultFromSession(session))
	waitForStarted(t, started)

	result, err := service.SubmitChoice(session.ID, "left")
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if got := strings.Join(result.Payload, ""); got != "超时同步回落" {
		t.Fatalf("expected timeout fallback payload, got %#v", result.Payload)
	}
	if factory.callCount() != 2 {
		t.Fatalf("expected prefetch plus fallback calls, got %d", factory.callCount())
	}
}

func TestGameServiceSubmitChoiceFallsBackWhenPrefetchFails(t *testing.T) {
	config := prefetchTestConfig()
	started := make(chan struct{})
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{
			{err: errors.New("prefetch failed"), started: started},
			{response: endedTurnResponse("失败同步回落")},
		},
	}
	service, session, pack := newPrefetchTestService(t, config, factory)
	service.maybeStartChoicePrefetch(session.Clone(), pack, config, roleplay.ResultFromSession(session))
	waitForStarted(t, started)

	result, err := service.SubmitChoice(session.ID, "left")
	if err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if got := strings.Join(result.Payload, ""); got != "失败同步回落" {
		t.Fatalf("expected failure fallback payload, got %#v", result.Payload)
	}
	if factory.callCount() != 2 {
		t.Fatalf("expected failed prefetch plus fallback calls, got %d", factory.callCount())
	}
}

func TestGameServiceDiscardStalePrefetchResult(t *testing.T) {
	config := prefetchTestConfig()
	factory := &scriptedChatFactory{
		clients: []*scriptedChatClient{{response: endedTurnResponse("过期预生成")}},
	}
	service, session, pack := newPrefetchTestService(t, config, factory)
	baseTurnID := session.Turns[len(session.Turns)-1].ID
	service.maybeStartChoicePrefetch(session.Clone(), pack, config, roleplay.ResultFromSession(session))
	session.AppendTurn(roleplay.GameTurn{
		ID:        "turn-new-base",
		Role:      roleplay.TurnRoleAI,
		Payload:   []string{"新的 base 回合。"},
		Tools:     testChoiceTools(),
		CreatedAt: roleplay.NowISO(),
	})

	_, used, err := service.promotePrefetchedChoice(session.ID, baseTurnID, "left", config)
	if err != nil {
		t.Fatalf("promote prefetch: %v", err)
	}
	if used {
		t.Fatal("stale prefetch result should not be promoted")
	}
	prefetchRoot := filepath.Join(filepath.Dir(session.WorkspacePath), ".prefetch")
	if entries, readErr := os.ReadDir(prefetchRoot); readErr == nil && len(entries) != 0 {
		t.Fatalf("expected stale prefetch workspace to be cleaned, got %d entries", len(entries))
	}
}

func TestCreatePrefetchBranchIsolatesWorkspace(t *testing.T) {
	pack, err := roleplay.NewStoryPack("game-a", testGameFiles(t))
	if err != nil {
		t.Fatalf("new story pack: %v", err)
	}
	session, err := roleplay.NewGameSessionInDir("game-a", pack, t.TempDir())
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := os.WriteFile(session.MemoryPath, []byte("source memory"), 0o600); err != nil {
		t.Fatalf("write source memory: %v", err)
	}
	if err := roleplay.WriteContextSummary(session, "## 当前阶段\n- source summary"); err != nil {
		t.Fatalf("write context summary: %v", err)
	}
	session.AppendTurn(roleplay.GameTurn{
		ID:        "turn-base",
		Role:      roleplay.TurnRoleAI,
		Payload:   []string{"你站在岔路口。"},
		Tools:     testChoiceTools(),
		CreatedAt: roleplay.NowISO(),
	})

	branch, workspace, err := createPrefetchBranch(session.Clone(), roleplay.ChoiceOption{ID: "left", Label: "向左"})
	if err != nil {
		t.Fatalf("create prefetch branch: %v", err)
	}
	defer os.RemoveAll(workspace)
	if branch.WorkspacePath == session.WorkspacePath || branch.MemoryPath == session.MemoryPath {
		t.Fatal("branch must use independent workspace paths")
	}
	if err := os.WriteFile(branch.MemoryPath, []byte("branch memory"), 0o600); err != nil {
		t.Fatalf("write branch memory: %v", err)
	}
	sourceMemory, err := os.ReadFile(session.MemoryPath)
	if err != nil {
		t.Fatalf("read source memory: %v", err)
	}
	if string(sourceMemory) != "source memory" {
		t.Fatalf("source memory was modified: %q", string(sourceMemory))
	}
	branchSummary, err := roleplay.ReadContextSummary(branch)
	if err != nil {
		t.Fatalf("read branch summary: %v", err)
	}
	if !strings.Contains(branchSummary, "source summary") {
		t.Fatalf("branch did not copy context summary, got:\n%s", branchSummary)
	}
}

func testLibraryGame(t *testing.T, id string) roleplay.LibraryGame {
	t.Helper()

	return roleplay.LibraryGame{
		ID:    id,
		Files: testGameFiles(t),
	}
}

func prefetchTestConfig() appconf.AppConfig {
	config := *appconf.DefaultConfig()
	config.AIChoicePrefetchEnabled = true
	config.AIChoicePrefetchGlobalConcurrency = 1
	config.AIChoicePrefetchSessionConcurrency = 1
	config.AIChoicePrefetchTTLMS = 1000
	config.AIChoicePrefetchWaitMS = 200
	return config
}

func newPrefetchTestService(t *testing.T, config appconf.AppConfig, factory *scriptedChatFactory) (*GameService, *roleplay.GameSession, roleplay.StoryPack) {
	t.Helper()

	pack, err := roleplay.NewStoryPack("game-a", testGameFiles(t))
	if err != nil {
		t.Fatalf("new story pack: %v", err)
	}
	session, err := roleplay.NewGameSessionInDir("game-a", pack, t.TempDir())
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.Label = "主线"
	session.ParentID = "parent-session"
	session.AppendTurn(roleplay.GameTurn{
		ID:        "turn-base",
		Role:      roleplay.TurnRoleAI,
		Payload:   []string{"你站在门前。"},
		Tools:     testChoiceTools(),
		CreatedAt: roleplay.NowISO(),
	})

	service := NewGameService(&config)
	service.packs[pack.ID] = pack
	service.sessions[session.ID] = session
	service.newChatClient = factory.factory
	return service, session, pack
}

func newAchievementTestService(t *testing.T, config appconf.AppConfig, factory *scriptedChatFactory, achievements string, parentID string) (*GameService, *roleplay.GameSession, roleplay.StoryPack, *achievementRepository) {
	t.Helper()

	files := testGameFiles(t)
	files[roleplay.AchievementsFileName] = achievements
	pack, err := roleplay.NewStoryPack("game-a", files)
	if err != nil {
		t.Fatalf("new story pack: %v", err)
	}
	session, err := roleplay.NewGameSessionInDir("game-a", pack, t.TempDir())
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.Label = "主线"
	session.ParentID = parentID
	session.AppendTurn(roleplay.GameTurn{
		ID:        "turn-base",
		Role:      roleplay.TurnRoleAI,
		Payload:   []string{"你站在门前。"},
		Tools:     testChoiceTools(),
		CreatedAt: roleplay.NowISO(),
	})

	repo := &achievementRepository{path: filepath.Join(t.TempDir(), "achievements.json")}
	service := NewGameService(&config)
	service.packs[pack.ID] = pack
	service.sessions[session.ID] = session
	service.achievementRepo = repo
	service.newChatClient = factory.factory
	return service, session, pack, repo
}

func testChoiceTools() []roleplay.ChoiceTool {
	return []roleplay.ChoiceTool{{
		Type: "choice",
		ID:   "main",
		Options: []roleplay.ChoiceOption{
			{ID: "left", Label: "向左"},
		},
	}}
}

func endedTurnResponse(payload string) string {
	return `{"type":"game_turn","state":"ended","payload":["` + payload + `"],"tools":[],"ending":{"id":"loop","title":"循环结局","kind":"loop"}}`
}

func achievementEndedResponse(payload string, achievementID string, title string, kind string) string {
	return `{"type":"game_turn","state":"ended","payload":["` + payload + `"],"tools":[],"ending":{"id":"` + achievementID + `","title":"` + title + `","kind":"` + kind + `"},"achievement":{"id":"` + achievementID + `","title":"` + title + `"}}`
}

func goodEndingResponse(payload string) string {
	return `{"type":"game_turn","state":"ended","payload":["` + payload + `"],"tools":[],"ending":{"id":"good-1","title":"好结局1","kind":"good"}}`
}

const customInputAchievementJSON = `[
  {
    "id": "watch_revenge",
    "title": "手表的复仇",
    "type": "ai_triggered",
    "trigger": "玩家明确把手表丢弃或扔掉。",
    "requiresCustomInput": true,
    "ending": {"id":"watch_revenge","title":"手表的复仇","kind":"bad"}
  }
]`

const choiceAchievementJSON = `[
  {
    "id": "two_friends",
    "title": "两个好朋友",
    "type": "ai_triggered",
    "trigger": "玩家把朋友拉到自己家并让双方都进入循环。",
    "ending": {"id":"two_friends","title":"两个好朋友","kind":"loop"}
  }
]`

const oneLifeAchievementJSON = `[
  {
    "id": "one_life_clear",
    "title": "殿堂级理解",
    "type": "rule_based",
    "ending": {"id":"hall_of_fame","title":"殿堂级理解","kind":"good"},
    "rule": {"kind":"one_life_completion","endingKind":"good","forbidSnapshotFork":true}
  }
]`

func waitForStarted(t *testing.T, started <-chan struct{}) {
	t.Helper()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for fake chat client to start")
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
