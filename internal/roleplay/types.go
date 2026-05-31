package roleplay

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	SessionStateIdle    = "idle"
	SessionStatePlaying = "playing"
	SessionStateEnded   = "ended"
	TurnRoleAI          = "ai"
	TurnRoleUser        = "user"
)

var RequiredStoryFiles = []string{"scene.md", "rule.md", "true.md", "memory.md", "endings.md"}

const MetadataFileName = "metadata.json"
const PlayerBriefingFileName = "briefing.json"
const AchievementsFileName = "achievements.json"

const (
	AchievementTypeAITriggered = "ai_triggered"
	AchievementTypeRuleBased   = "rule_based"
	AchievementRuleOneLife     = "one_life_completion"
)

type StoryPack struct {
	ID               string                  `json:"id"`
	Files            map[string]string       `json:"files"`
	Scenes           []SceneAsset            `json:"scenes,omitempty"`
	BGMs             []BGMAsset              `json:"bgms,omitempty"`
	BGMSceneDefaults map[string]string       `json:"bgmSceneDefaults,omitempty"`
	MapURLs          []string                `json:"mapUrls,omitempty"`
	Achievements     []AchievementDefinition `json:"achievements,omitempty"`
}

type GameMetadata struct {
	Title          string               `json:"title"`
	TTitle         string               `json:"ttitle"`
	InitialScene   string               `json:"initialScene"`
	ScenePositions map[string][]float64 `json:"scenePositions"`
}

func (m GameMetadata) GameTitle() string {
	if title := strings.TrimSpace(m.Title); title != "" {
		return title
	}
	return strings.TrimSpace(m.TTitle)
}

type LibraryGame struct {
	ID           string                  `json:"id"`
	Title        string                  `json:"title"`
	ImportedAt   string                  `json:"importedAt"`
	Files        map[string]string       `json:"files"`
	PhotoURLs    []string                `json:"photoUrls"`
	MapURLs      []string                `json:"mapUrls"`
	Scenes       []SceneAsset            `json:"scenes,omitempty"`
	BGMs         []BGMAsset              `json:"bgms,omitempty"`
	Achievements []AchievementDefinition `json:"achievements,omitempty"`
}

type ImportGameResult struct {
	Game       *LibraryGame `json:"game,omitempty"`
	Missing    []string     `json:"missing"`
	Warnings   []string     `json:"warnings"`
	ValidFiles []string     `json:"validFiles"`
}

type GameSession struct {
	ID             string     `json:"id"`
	GameID         string     `json:"gameId"`
	State          string     `json:"state"`
	CurrentSceneID string     `json:"currentSceneId,omitempty"`
	CurrentBGMID   string     `json:"currentBgmId,omitempty"`
	WorkspacePath  string     `json:"workspacePath"`
	MemoryPath     string     `json:"memoryPath"`
	Turns          []GameTurn `json:"turns"`
	Label          string     `json:"label,omitempty"`
	IsSnapshot     bool       `json:"isSnapshot,omitempty"`
	ParentID       string     `json:"parentId,omitempty"`
	CreatedAt      string     `json:"createdAt"`
	UpdatedAt      string     `json:"updatedAt"`
}

type GameTurn struct {
	ID                  string                `json:"id"`
	Role                string                `json:"role"`
	Payload             []string              `json:"payload"`
	SelectedChoiceID    string                `json:"selectedChoiceId,omitempty"`
	SelectedChoiceLabel string                `json:"selectedChoiceLabel,omitempty"`
	CustomInput         bool                  `json:"customInput,omitempty"`
	Tools               []ChoiceTool          `json:"tools,omitempty"`
	Scene               *SceneChange          `json:"scene,omitempty"`
	BGM                 *BGMChange            `json:"bgm,omitempty"`
	Ending              *Ending               `json:"ending,omitempty"`
	Achievement         *AchievementReference `json:"achievement,omitempty"`
	CreatedAt           string                `json:"createdAt"`
}

type GameTurnResult struct {
	SessionID    string                   `json:"sessionId"`
	GameID       string                   `json:"gameId"`
	State        string                   `json:"state"`
	Payload      []string                 `json:"payload"`
	Tools        []ChoiceTool             `json:"tools"`
	Scene        *SceneChange             `json:"scene,omitempty"`
	BGM          *BGMChange               `json:"bgm,omitempty"`
	CurrentBGMID string                   `json:"currentBgmId,omitempty"`
	Ending       *Ending                  `json:"ending,omitempty"`
	Achievement  *AchievementUnlockResult `json:"achievement,omitempty"`
	Turn         GameTurn                 `json:"turn"`
}

type ChoiceTool struct {
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Prompt  string         `json:"prompt,omitempty"`
	Options []ChoiceOption `json:"options"`
}

type ChoiceOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type SceneAsset struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	FileName    string  `json:"fileName"`
	URL         string  `json:"url"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	HasPosition bool    `json:"hasPosition"`
}

type BGMAsset struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	FileName string `json:"fileName"`
	URL      string `json:"url"`
}

type SceneChange struct {
	ID     string `json:"id"`
	Reason string `json:"reason,omitempty"`
}

type BGMChange struct {
	Action string `json:"action"`
	ID     string `json:"id,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type Ending struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

type AchievementDefinition struct {
	ID                  string           `json:"id"`
	Title               string           `json:"title"`
	Type                string           `json:"type,omitempty"`
	Trigger             string           `json:"trigger,omitempty"`
	RequiresCustomInput bool             `json:"requiresCustomInput,omitempty"`
	Ending              Ending           `json:"ending"`
	Rule                *AchievementRule `json:"rule,omitempty"`
}

type AchievementRule struct {
	Kind               string `json:"kind"`
	EndingID           string `json:"endingId,omitempty"`
	EndingKind         string `json:"endingKind,omitempty"`
	ForbidSnapshotFork bool   `json:"forbidSnapshotFork,omitempty"`
}

type AchievementReference struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type AchievementUnlock struct {
	GameID        string `json:"gameId"`
	AchievementID string `json:"achievementId"`
	Title         string `json:"title"`
	SessionID     string `json:"sessionId"`
	EndingID      string `json:"endingId,omitempty"`
	UnlockedAt    string `json:"unlockedAt"`
}

type AchievementUnlockResult struct {
	GameID        string `json:"gameId"`
	AchievementID string `json:"achievementId"`
	Title         string `json:"title"`
	SessionID     string `json:"sessionId"`
	EndingID      string `json:"endingId,omitempty"`
	UnlockedAt    string `json:"unlockedAt,omitempty"`
	New           bool   `json:"new"`
}

type AgentTerminalRequest struct {
	Type     string            `json:"type"`
	Reason   string            `json:"reason"`
	Commands []TerminalCommand `json:"commands"`
}

type TerminalCommand struct {
	Command string `json:"command"`
}

type TerminalExecution struct {
	Command     string `json:"command"`
	Stdout      string `json:"stdout"`
	Stderr      string `json:"stderr"`
	ExitCode    int    `json:"exitCode"`
	DurationMS  int64  `json:"durationMs"`
	TimedOut    bool   `json:"timedOut"`
	Truncated   bool   `json:"truncated"`
	Error       string `json:"error,omitempty"`
	CompletedAt string `json:"completedAt"`
}

func NewStoryPack(gameID string, files map[string]string) (StoryPack, error) {
	normalized := make(map[string]string, len(files))
	for name, content := range files {
		normalized[normalizeRelativePath(name)] = content
	}

	var missing []string
	for _, fileName := range RequiredStoryFiles {
		if _, ok := normalized[strings.ToLower(fileName)]; !ok {
			missing = append(missing, fileName)
		}
	}
	if len(missing) > 0 {
		return StoryPack{}, fmt.Errorf("missing story pack files: %s", strings.Join(missing, ", "))
	}

	bgms := parseBGMAssets(normalized)
	achievements, err := parseAchievementDefinitions(normalized)
	if err != nil {
		return StoryPack{}, err
	}
	return StoryPack{
		ID:               gameID,
		Files:            normalized,
		Scenes:           orderScenesWithInitial(parseSceneAssets(normalized), parseInitialScene(normalized)),
		BGMs:             bgms,
		BGMSceneDefaults: parseBGMSceneDefaults(normalized, bgms),
		MapURLs:          parseMapURLs(normalized),
		Achievements:     achievements,
	}, nil
}

func NewLibraryGame(files map[string]string) (LibraryGame, ImportGameResult, error) {
	normalized := normalizeFileContents(files)
	validFiles := make([]string, 0, len(RequiredStoryFiles)+3)
	storyFileNames := append([]string{MetadataFileName}, RequiredStoryFiles...)
	storyFileNames = append(storyFileNames, PlayerBriefingFileName)
	storyFileNames = append(storyFileNames, AchievementsFileName)
	for _, fileName := range storyFileNames {
		if _, ok := normalized[strings.ToLower(fileName)]; ok {
			validFiles = append(validFiles, fileName)
		}
	}
	if _, ok := normalized[normalizeRelativePath("bgm/metadata.json")]; ok {
		validFiles = append(validFiles, "bgm/metadata.json")
	}

	missing := requiredImportFilesMissing(normalized)
	if len(missing) > 0 {
		return LibraryGame{}, ImportGameResult{
			Missing:    missing,
			Warnings:   []string{},
			ValidFiles: validFiles,
		}, nil
	}

	var metadata GameMetadata
	if err := json.Unmarshal([]byte(normalized[strings.ToLower(MetadataFileName)]), &metadata); err != nil {
		return LibraryGame{}, ImportGameResult{
			Missing:    []string{},
			Warnings:   []string{"metadata.json 解析失败"},
			ValidFiles: validFiles,
		}, fmt.Errorf("parse metadata.json: %w", err)
	}

	title := metadata.GameTitle()
	if title == "" {
		return LibraryGame{}, ImportGameResult{
			Missing:    []string{"metadata.json:title"},
			Warnings:   []string{},
			ValidFiles: validFiles,
		}, nil
	}

	achievements, err := parseAchievementDefinitions(normalized)
	if err != nil {
		return LibraryGame{}, ImportGameResult{
			Missing:    []string{},
			Warnings:   []string{"achievements.json 解析失败"},
			ValidFiles: validFiles,
		}, err
	}

	game := LibraryGame{
		ID:           NewID("game"),
		Title:        title,
		ImportedAt:   NowISO(),
		Files:        normalized,
		Scenes:       orderScenesWithInitial(parseSceneAssets(normalized), metadata.InitialScene),
		BGMs:         parseBGMAssets(normalized),
		MapURLs:      parseMapURLs(normalized),
		Achievements: achievements,
	}
	for _, scene := range game.Scenes {
		game.PhotoURLs = append(game.PhotoURLs, scene.URL)
	}

	return game, ImportGameResult{
		Game:       &game,
		Missing:    []string{},
		Warnings:   []string{},
		ValidFiles: validFiles,
	}, nil
}

func normalizeFileContents(files map[string]string) map[string]string {
	normalized := make(map[string]string, len(files))
	for name, content := range files {
		normalized[normalizeRelativePath(name)] = content
	}
	return normalized
}

func requiredImportFilesMissing(files map[string]string) []string {
	var missing []string
	if _, ok := files[strings.ToLower(MetadataFileName)]; !ok {
		missing = append(missing, MetadataFileName)
	}
	for _, fileName := range RequiredStoryFiles {
		if _, ok := files[strings.ToLower(fileName)]; !ok {
			missing = append(missing, fileName)
		}
	}
	return missing
}

func NewGameSession(gameID string, pack StoryPack) (*GameSession, error) {
	workspace, err := os.MkdirTemp("", "namelesswatch-session-*")
	if err != nil {
		return nil, fmt.Errorf("create session workspace: %w", err)
	}
	return newGameSessionWithWorkspace(NewID("session"), gameID, pack, workspace)
}

// NewGameSessionInDir creates a session whose workspace lives under a persistent
// base directory (baseDir/{gameID}/{sessionID}) instead of the system temp dir,
// so the session and its memory.md survive process restarts.
func NewGameSessionInDir(gameID string, pack StoryPack, baseDir string) (*GameSession, error) {
	sessionID := NewID("session")
	workspace := filepath.Join(baseDir, safePathSegment(gameID), sessionID)
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return nil, fmt.Errorf("create session workspace: %w", err)
	}
	return newGameSessionWithWorkspace(sessionID, gameID, pack, workspace)
}

func newGameSessionWithWorkspace(sessionID, gameID string, pack StoryPack, workspace string) (*GameSession, error) {
	for _, fileName := range RequiredStoryFiles {
		if err := os.WriteFile(filepath.Join(workspace, fileName), []byte(pack.Files[strings.ToLower(fileName)]), 0o600); err != nil {
			return nil, fmt.Errorf("copy %s to session workspace: %w", fileName, err)
		}
	}
	if err := EnsureContextSummary(&GameSession{WorkspacePath: workspace}); err != nil {
		return nil, err
	}

	now := NowISO()
	currentSceneID := ""
	if len(pack.Scenes) > 0 {
		currentSceneID = pack.Scenes[0].ID
	}
	currentBGMID := pack.BGMSceneDefaults[currentSceneID]
	return &GameSession{
		ID:             sessionID,
		GameID:         gameID,
		State:          SessionStatePlaying,
		CurrentSceneID: currentSceneID,
		CurrentBGMID:   currentBGMID,
		WorkspacePath:  workspace,
		MemoryPath:     filepath.Join(workspace, "memory.md"),
		Turns:          []GameTurn{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (s *GameSession) AppendTurn(turn GameTurn) {
	s.Turns = append(s.Turns, turn)
	s.UpdatedAt = NowISO()
	if turn.Role == TurnRoleAI && turn.Ending != nil {
		s.State = SessionStateEnded
	}
}

func (s *GameSession) ChoiceLabel(choiceID string) string {
	for i := len(s.Turns) - 1; i >= 0; i-- {
		turn := s.Turns[i]
		if turn.Role != TurnRoleAI {
			continue
		}
		for _, tool := range turn.Tools {
			for _, option := range tool.Options {
				if option.ID == choiceID {
					return option.Label
				}
			}
		}
		break
	}
	return choiceID
}

func (s *GameSession) LatestAITurn() (GameTurn, bool) {
	if s == nil {
		return GameTurn{}, false
	}
	for i := len(s.Turns) - 1; i >= 0; i-- {
		if s.Turns[i].Role == TurnRoleAI {
			return s.Turns[i], true
		}
	}
	return GameTurn{}, false
}

func (s *GameSession) LatestUserTurn() (GameTurn, bool) {
	if s == nil {
		return GameTurn{}, false
	}
	for i := len(s.Turns) - 1; i >= 0; i-- {
		if s.Turns[i].Role == TurnRoleUser {
			return s.Turns[i], true
		}
	}
	return GameTurn{}, false
}

func (s *GameSession) Clone() GameSession {
	turns := slices.Clone(s.Turns)
	return GameSession{
		ID:             s.ID,
		GameID:         s.GameID,
		State:          s.State,
		CurrentSceneID: s.CurrentSceneID,
		CurrentBGMID:   s.CurrentBGMID,
		WorkspacePath:  s.WorkspacePath,
		MemoryPath:     s.MemoryPath,
		Turns:          turns,
		Label:          s.Label,
		IsSnapshot:     s.IsSnapshot,
		ParentID:       s.ParentID,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func ResultFromSession(session *GameSession) GameTurnResult {
	var last GameTurn
	for i := len(session.Turns) - 1; i >= 0; i-- {
		if session.Turns[i].Role == TurnRoleAI {
			last = session.Turns[i]
			break
		}
	}

	return GameTurnResult{
		SessionID:    session.ID,
		GameID:       session.GameID,
		State:        session.State,
		Payload:      last.Payload,
		Tools:        last.Tools,
		Scene:        last.Scene,
		BGM:          last.BGM,
		CurrentBGMID: session.CurrentBGMID,
		Ending:       last.Ending,
		Achievement:  AchievementResultFromReference(session.GameID, session.ID, last.Ending, last.Achievement),
		Turn:         last,
	}
}

func AchievementResultFromReference(gameID, sessionID string, ending *Ending, achievement *AchievementReference) *AchievementUnlockResult {
	if achievement == nil {
		return nil
	}
	endingID := ""
	if ending != nil {
		endingID = ending.ID
	}
	return &AchievementUnlockResult{
		GameID:        gameID,
		AchievementID: achievement.ID,
		Title:         achievement.Title,
		SessionID:     sessionID,
		EndingID:      endingID,
		New:           false,
	}
}

func AchievementResultFromUnlock(unlock AchievementUnlock, newlyUnlocked bool) AchievementUnlockResult {
	return AchievementUnlockResult{
		GameID:        unlock.GameID,
		AchievementID: unlock.AchievementID,
		Title:         unlock.Title,
		SessionID:     unlock.SessionID,
		EndingID:      unlock.EndingID,
		UnlockedAt:    unlock.UnlockedAt,
		New:           newlyUnlocked,
	}
}

func FindAchievementDefinition(achievements []AchievementDefinition, achievementID string) (AchievementDefinition, bool) {
	achievementID = strings.TrimSpace(achievementID)
	if achievementID == "" {
		return AchievementDefinition{}, false
	}
	for _, achievement := range achievements {
		if achievement.ID == achievementID {
			return achievement, true
		}
	}
	return AchievementDefinition{}, false
}

func NewID(prefix string) string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(bytes[:])
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func ReadWorkspaceFile(session *GameSession, fileName string) (string, error) {
	if !slices.Contains(RequiredStoryFiles, fileName) {
		return "", errors.New("unsupported story file")
	}
	content, err := os.ReadFile(filepath.Join(session.WorkspacePath, fileName))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func normalizeRelativePath(name string) string {
	value := strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "/")
	value = path.Clean(value)
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "/")
	return strings.ToLower(value)
}

func safePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	value = replacer.Replace(value)
	if value == "." || value == ".." {
		return "unknown"
	}
	return value
}

func parseSceneAssets(files map[string]string) []SceneAsset {
	metadataKey := normalizeRelativePath("photo/metadata.json")
	metadataRaw, ok := files[metadataKey]
	if !ok || strings.TrimSpace(metadataRaw) == "" {
		return []SceneAsset{}
	}

	var mapping map[string]string
	if err := json.Unmarshal([]byte(metadataRaw), &mapping); err != nil {
		return []SceneAsset{}
	}

	positions := loadGameMetadata(files).ScenePositions
	scenes := make([]SceneAsset, 0, len(mapping))
	for sceneID, fileName := range mapping {
		id := strings.TrimSpace(sceneID)
		assetName := strings.TrimSpace(fileName)
		if id == "" || assetName == "" {
			continue
		}

		assetKey := normalizeRelativePath(filepath.Join("photo", assetName))
		url := files[assetKey]
		if strings.TrimSpace(url) == "" {
			continue
		}

		scene := SceneAsset{
			ID:       id,
			Name:     id,
			FileName: assetName,
			URL:      url,
		}
		if pos := positions[id]; len(pos) >= 2 {
			scene.X = pos[0]
			scene.Y = pos[1]
			scene.HasPosition = true
		}
		scenes = append(scenes, scene)
	}

	return scenes
}

type bgmMetadata struct {
	Tracks        map[string]bgmTrackMetadata `json:"tracks"`
	SceneDefaults map[string]string           `json:"sceneDefaults"`
}

type bgmTrackMetadata struct {
	Name string `json:"name"`
	File string `json:"file"`
}

func parseBGMMetadata(files map[string]string) bgmMetadata {
	metadataKey := normalizeRelativePath("bgm/metadata.json")
	metadataRaw, ok := files[metadataKey]
	if !ok || strings.TrimSpace(metadataRaw) == "" {
		return bgmMetadata{}
	}

	var metadata bgmMetadata
	if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {
		return bgmMetadata{}
	}
	return metadata
}

func parseBGMAssets(files map[string]string) []BGMAsset {
	metadata := parseBGMMetadata(files)
	if len(metadata.Tracks) == 0 {
		return []BGMAsset{}
	}

	bgms := make([]BGMAsset, 0, len(metadata.Tracks))
	for trackID, track := range metadata.Tracks {
		id := strings.TrimSpace(trackID)
		fileName := strings.TrimSpace(track.File)
		if id == "" || fileName == "" || !isSupportedBGMAudioFile(fileName) {
			continue
		}

		assetKey := normalizeRelativePath(path.Join("bgm", strings.ReplaceAll(fileName, "\\", "/")))
		if !strings.HasPrefix(assetKey, "bgm/") {
			continue
		}
		url := strings.TrimSpace(files[assetKey])
		if url == "" {
			continue
		}

		name := strings.TrimSpace(track.Name)
		if name == "" {
			name = id
		}
		bgms = append(bgms, BGMAsset{
			ID:       id,
			Name:     name,
			FileName: fileName,
			URL:      url,
		})
	}

	slices.SortFunc(bgms, func(a, b BGMAsset) int {
		return strings.Compare(a.ID, b.ID)
	})
	return bgms
}

func parseBGMSceneDefaults(files map[string]string, bgms []BGMAsset) map[string]string {
	metadata := parseBGMMetadata(files)
	if len(metadata.SceneDefaults) == 0 || len(bgms) == 0 {
		return nil
	}

	available := make(map[string]bool, len(bgms))
	for _, bgm := range bgms {
		available[bgm.ID] = true
	}

	defaults := make(map[string]string, len(metadata.SceneDefaults))
	for sceneID, bgmID := range metadata.SceneDefaults {
		sceneID = strings.TrimSpace(sceneID)
		bgmID = strings.TrimSpace(bgmID)
		if sceneID == "" || bgmID == "" || !available[bgmID] {
			continue
		}
		defaults[sceneID] = bgmID
	}
	if len(defaults) == 0 {
		return nil
	}
	return defaults
}

func isSupportedBGMAudioFile(fileName string) bool {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".mp3", ".ogg", ".wav", ".m4a", ".webm":
		return true
	default:
		return false
	}
}

// parseMapURLs locates the optional map background image bundled with a story pack.
func parseMapURLs(files map[string]string) []string {
	for _, candidate := range []string{"photo/map.png", "map.png", "map/map.png"} {
		if url := strings.TrimSpace(files[normalizeRelativePath(candidate)]); url != "" {
			return []string{url}
		}
	}
	return nil
}

func parseAchievementDefinitions(files map[string]string) ([]AchievementDefinition, error) {
	raw := strings.TrimSpace(files[normalizeRelativePath(AchievementsFileName)])
	if raw == "" {
		return nil, nil
	}

	var definitions []AchievementDefinition
	if err := json.Unmarshal([]byte(raw), &definitions); err != nil {
		var envelope struct {
			Achievements []AchievementDefinition `json:"achievements"`
		}
		if envelopeErr := json.Unmarshal([]byte(raw), &envelope); envelopeErr != nil {
			return nil, fmt.Errorf("parse %s: %w", AchievementsFileName, err)
		}
		definitions = envelope.Achievements
	}
	return normalizeAchievementDefinitions(definitions)
}

func normalizeAchievementDefinitions(definitions []AchievementDefinition) ([]AchievementDefinition, error) {
	if len(definitions) == 0 {
		return nil, nil
	}

	seen := make(map[string]bool, len(definitions))
	normalized := make([]AchievementDefinition, 0, len(definitions))
	for index, definition := range definitions {
		definition.ID = strings.TrimSpace(definition.ID)
		definition.Title = strings.TrimSpace(definition.Title)
		definition.Type = strings.TrimSpace(definition.Type)
		definition.Trigger = strings.TrimSpace(definition.Trigger)
		definition.Ending = normalizeEnding(definition.Ending)
		if definition.Rule != nil {
			rule := normalizeAchievementRule(*definition.Rule)
			definition.Rule = &rule
		}

		if definition.ID == "" {
			return nil, fmt.Errorf("achievement %d id is required", index+1)
		}
		if definition.Title == "" {
			return nil, fmt.Errorf("achievement %q title is required", definition.ID)
		}
		if seen[definition.ID] {
			return nil, fmt.Errorf("achievement %q is duplicated", definition.ID)
		}
		seen[definition.ID] = true

		if definition.Type == "" {
			if definition.Rule != nil {
				definition.Type = AchievementTypeRuleBased
			} else {
				definition.Type = AchievementTypeAITriggered
			}
		}
		if err := validateAchievementDefinition(definition); err != nil {
			return nil, err
		}
		normalized = append(normalized, definition)
	}
	return normalized, nil
}

func validateAchievementDefinition(definition AchievementDefinition) error {
	if err := validateAchievementEnding(definition.Ending); err != nil {
		return fmt.Errorf("achievement %q ending is invalid: %w", definition.ID, err)
	}

	switch definition.Type {
	case AchievementTypeAITriggered:
		if strings.TrimSpace(definition.Trigger) == "" {
			return fmt.Errorf("achievement %q trigger is required", definition.ID)
		}
	case AchievementTypeRuleBased:
		if definition.Rule == nil {
			return fmt.Errorf("achievement %q rule is required", definition.ID)
		}
		if err := validateAchievementRule(*definition.Rule); err != nil {
			return fmt.Errorf("achievement %q rule is invalid: %w", definition.ID, err)
		}
	default:
		return fmt.Errorf("achievement %q type %q is unsupported", definition.ID, definition.Type)
	}
	return nil
}

func normalizeEnding(ending Ending) Ending {
	ending.ID = strings.TrimSpace(ending.ID)
	ending.Title = strings.TrimSpace(ending.Title)
	ending.Kind = strings.TrimSpace(ending.Kind)
	return ending
}

func validateAchievementEnding(ending Ending) error {
	if ending.ID == "" {
		return errors.New("id is required")
	}
	if ending.Title == "" {
		return errors.New("title is required")
	}
	if !isValidEndingKind(ending.Kind) {
		return fmt.Errorf("kind %q is unsupported", ending.Kind)
	}
	return nil
}

func normalizeAchievementRule(rule AchievementRule) AchievementRule {
	rule.Kind = strings.TrimSpace(rule.Kind)
	rule.EndingID = strings.TrimSpace(rule.EndingID)
	rule.EndingKind = strings.TrimSpace(rule.EndingKind)
	return rule
}

func validateAchievementRule(rule AchievementRule) error {
	switch rule.Kind {
	case AchievementRuleOneLife:
		if rule.EndingID == "" && rule.EndingKind == "" {
			return errors.New("endingId or endingKind is required")
		}
		if rule.EndingKind != "" && !isValidEndingKind(rule.EndingKind) {
			return fmt.Errorf("endingKind %q is unsupported", rule.EndingKind)
		}
	default:
		return fmt.Errorf("kind %q is unsupported", rule.Kind)
	}
	return nil
}

func isValidEndingKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "good", "bad", "loop", "neutral":
		return true
	default:
		return false
	}
}

// loadGameMetadata decodes metadata.json from the (normalized) pack files. It returns a
// zero-value GameMetadata when the file is missing or cannot be parsed.
func loadGameMetadata(files map[string]string) GameMetadata {
	var metadata GameMetadata
	raw, ok := files[strings.ToLower(MetadataFileName)]
	if !ok || strings.TrimSpace(raw) == "" {
		return metadata
	}
	_ = json.Unmarshal([]byte(raw), &metadata)
	return metadata
}

// parseInitialScene reads metadata.json's initialScene field, which declares the scene
// the game should start in. Returns "" when absent or unparseable.
func parseInitialScene(files map[string]string) string {
	return strings.TrimSpace(loadGameMetadata(files).InitialScene)
}

// orderScenesWithInitial returns scenes in a deterministic order: the scene whose ID
// matches initialID (when present) comes first, the rest follow sorted by ID. This avoids
// depending on Go's randomized map iteration so the starting scene is stable across runs.
func orderScenesWithInitial(scenes []SceneAsset, initialID string) []SceneAsset {
	if len(scenes) == 0 {
		return scenes
	}

	ordered := make([]SceneAsset, len(scenes))
	copy(ordered, scenes)
	slices.SortFunc(ordered, func(a, b SceneAsset) int {
		return strings.Compare(a.ID, b.ID)
	})

	initialID = strings.TrimSpace(initialID)
	if initialID == "" {
		return ordered
	}
	for i, scene := range ordered {
		if scene.ID == initialID {
			rest := append(ordered[:i:i], ordered[i+1:]...)
			return append([]SceneAsset{scene}, rest...)
		}
	}
	return ordered
}
