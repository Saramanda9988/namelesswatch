package roleplay

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

type StoryPack struct {
	ID    string            `json:"id"`
	Files map[string]string `json:"files"`
}

type GameMetadata struct {
	Title  string `json:"title"`
	TTitle string `json:"ttitle"`
}

func (m GameMetadata) GameTitle() string {
	if title := strings.TrimSpace(m.Title); title != "" {
		return title
	}
	return strings.TrimSpace(m.TTitle)
}

type LibraryGame struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	ImportedAt string            `json:"importedAt"`
	Files      map[string]string `json:"files"`
	PhotoURLs  []string          `json:"photoUrls"`
	MapURLs    []string          `json:"mapUrls"`
}

type ImportGameResult struct {
	Game       *LibraryGame `json:"game,omitempty"`
	Missing    []string     `json:"missing"`
	Warnings   []string     `json:"warnings"`
	ValidFiles []string     `json:"validFiles"`
}

type GameSession struct {
	ID            string     `json:"id"`
	GameID        string     `json:"gameId"`
	State         string     `json:"state"`
	WorkspacePath string     `json:"workspacePath"`
	MemoryPath    string     `json:"memoryPath"`
	Turns         []GameTurn `json:"turns"`
	CreatedAt     string     `json:"createdAt"`
	UpdatedAt     string     `json:"updatedAt"`
}

type GameTurn struct {
	ID                  string       `json:"id"`
	Role                string       `json:"role"`
	Payload             []string     `json:"payload"`
	SelectedChoiceID    string       `json:"selectedChoiceId,omitempty"`
	SelectedChoiceLabel string       `json:"selectedChoiceLabel,omitempty"`
	Tools               []ChoiceTool `json:"tools,omitempty"`
	Ending              *Ending      `json:"ending,omitempty"`
	CreatedAt           string       `json:"createdAt"`
}

type GameTurnResult struct {
	SessionID string       `json:"sessionId"`
	GameID    string       `json:"gameId"`
	State     string       `json:"state"`
	Payload   []string     `json:"payload"`
	Tools     []ChoiceTool `json:"tools"`
	Ending    *Ending      `json:"ending,omitempty"`
	Turn      GameTurn     `json:"turn"`
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

type Ending struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
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
		normalized[strings.ToLower(filepath.Base(strings.ReplaceAll(name, "\\", "/")))] = content
	}

	var missing []string
	for _, fileName := range RequiredStoryFiles {
		if _, ok := normalized[fileName]; !ok {
			missing = append(missing, fileName)
		}
	}
	if len(missing) > 0 {
		return StoryPack{}, fmt.Errorf("missing story pack files: %s", strings.Join(missing, ", "))
	}

	return StoryPack{
		ID:    gameID,
		Files: normalized,
	}, nil
}

func NewLibraryGame(files map[string]string) (LibraryGame, ImportGameResult, error) {
	normalized := normalizeFileContents(files)
	validFiles := make([]string, 0, len(RequiredStoryFiles)+1)
	for _, fileName := range append([]string{MetadataFileName}, RequiredStoryFiles...) {
		if _, ok := normalized[fileName]; ok {
			validFiles = append(validFiles, fileName)
		}
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
	if err := json.Unmarshal([]byte(normalized[MetadataFileName]), &metadata); err != nil {
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

	game := LibraryGame{
		ID:         NewID("game"),
		Title:      title,
		ImportedAt: NowISO(),
		Files:      normalized,
		PhotoURLs:  []string{},
		MapURLs:    []string{},
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
		normalized[strings.ToLower(filepath.Base(strings.ReplaceAll(name, "\\", "/")))] = content
	}
	return normalized
}

func requiredImportFilesMissing(files map[string]string) []string {
	var missing []string
	if _, ok := files[MetadataFileName]; !ok {
		missing = append(missing, MetadataFileName)
	}
	for _, fileName := range RequiredStoryFiles {
		if _, ok := files[fileName]; !ok {
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
	return NewGameSessionInWorkspace(gameID, pack, workspace)
}

func NewGameSessionInRoot(gameID string, pack StoryPack, sessionsRoot string) (*GameSession, error) {
	sessionID := NewID("session")
	workspace := filepath.Join(sessionsRoot, safePathSegment(gameID), sessionID)
	session, err := NewGameSessionInWorkspace(gameID, pack, workspace)
	if err != nil {
		return nil, err
	}
	session.ID = sessionID
	return session, nil
}

func NewGameSessionInWorkspace(gameID string, pack StoryPack, workspace string) (*GameSession, error) {
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return nil, fmt.Errorf("create session workspace: %w", err)
	}
	for _, fileName := range RequiredStoryFiles {
		if err := os.WriteFile(filepath.Join(workspace, fileName), []byte(pack.Files[fileName]), 0o600); err != nil {
			return nil, fmt.Errorf("copy %s to session workspace: %w", fileName, err)
		}
	}

	now := NowISO()
	return &GameSession{
		ID:            NewID("session"),
		GameID:        gameID,
		State:         SessionStatePlaying,
		WorkspacePath: workspace,
		MemoryPath:    filepath.Join(workspace, "memory.md"),
		Turns:         []GameTurn{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
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

func (s *GameSession) Clone() GameSession {
	turns := slices.Clone(s.Turns)
	return GameSession{
		ID:            s.ID,
		GameID:        s.GameID,
		State:         s.State,
		WorkspacePath: s.WorkspacePath,
		MemoryPath:    s.MemoryPath,
		Turns:         turns,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
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
		SessionID: session.ID,
		GameID:    session.GameID,
		State:     session.State,
		Payload:   last.Payload,
		Tools:     last.Tools,
		Ending:    last.Ending,
		Turn:      last,
	}
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
