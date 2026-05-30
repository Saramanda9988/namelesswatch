package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"strings"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type GameService struct {
	ctx      context.Context
	mu       sync.Mutex
	config   appconf.AppConfig
	packs    map[string]roleplay.StoryPack
	sessions map[string]*roleplay.GameSession
	games    map[string]roleplay.LibraryGame
	gameIDs  []string
	repo     *gameRepository
}

func NewGameService(config *appconf.AppConfig) *GameService {
	initialConfig := appconf.AppConfig{}
	if config != nil {
		initialConfig = *config
	}
	return &GameService{
		config:   initialConfig,
		packs:    make(map[string]roleplay.StoryPack),
		sessions: make(map[string]*roleplay.GameSession),
		games:    make(map[string]roleplay.LibraryGame),
		gameIDs:  []string{},
	}
}

func (s *GameService) Init(ctx context.Context) {
	s.ctx = ctx
}

func (s *GameService) LoadLibrary() error {
	repo, err := newGameRepository()
	if err != nil {
		return err
	}
	games, err := repo.load()
	if err != nil {
		return err
	}

	nextGames := make(map[string]roleplay.LibraryGame, len(games))
	nextPacks := make(map[string]roleplay.StoryPack, len(games))
	nextGameIDs := make([]string, 0, len(games))
	for _, game := range games {
		normalized, pack, err := normalizeLibraryGame(game)
		if err != nil {
			return fmt.Errorf("load game %q: %w", game.ID, err)
		}
		if _, exists := nextGames[normalized.ID]; exists {
			continue
		}
		nextGames[normalized.ID] = normalized
		nextPacks[normalized.ID] = pack
		nextGameIDs = append(nextGameIDs, normalized.ID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.repo = repo
	s.games = nextGames
	s.packs = nextPacks
	s.gameIDs = nextGameIDs
	return nil
}

func (s *GameService) SetConfig(config appconf.AppConfig) {
	appconf.Normalize(&config)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

func (s *GameService) RegisterGamePack(gameID string, files map[string]string) error {
	if gameID == "" {
		return errors.New("game id is required")
	}
	s.logInfof("register_game_pack game=%s files=%d", gameID, len(files))

	pack, err := roleplay.NewStoryPack(gameID, files)
	if err != nil {
		s.logErrorf("register_game_pack failed game=%s error=%v", gameID, err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.packs[gameID] = pack
	return nil
}

func (s *GameService) ImportGamePack(files map[string]string) (roleplay.ImportGameResult, error) {
	s.logInfof("import_game_pack files=%d", len(files))
	game, result, err := roleplay.NewLibraryGame(files)
	if err != nil {
		s.logErrorf("import_game_pack failed missing=%v valid=%v error=%v", result.Missing, result.ValidFiles, err)
		return result, err
	}
	if result.Game == nil {
		s.logWarningf("import_game_pack incomplete missing=%v valid=%v warnings=%v", result.Missing, result.ValidFiles, result.Warnings)
		return result, nil
	}

	pack, err := roleplay.NewStoryPack(game.ID, game.Files)
	if err != nil {
		s.logErrorf("import_game_pack story_pack_failed game=%s error=%v", game.ID, err)
		return roleplay.ImportGameResult{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[game.ID] = game
	s.gameIDs = append([]string{game.ID}, s.gameIDs...)
	s.packs[game.ID] = pack
	if err := s.persistLocked(); err != nil {
		s.logErrorf("import_game_pack persist_failed game=%s error=%v", game.ID, err)
		return roleplay.ImportGameResult{}, err
	}

	s.logInfof("import_game_pack success game=%s title=%q", game.ID, game.Title)
	return result, nil
}

func (s *GameService) GetGames() ([]roleplay.LibraryGame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	games := make([]roleplay.LibraryGame, 0, len(s.gameIDs))
	for _, gameID := range s.gameIDs {
		if game, ok := s.games[gameID]; ok {
			games = append(games, cloneLibraryGame(game))
		}
	}
	return games, nil
}

func (s *GameService) GetGame(gameID string) (roleplay.LibraryGame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, ok := s.games[gameID]
	if !ok {
		return roleplay.LibraryGame{}, errors.New("game not found")
	}
	return cloneLibraryGame(game), nil
}

func (s *GameService) CreateGame(game roleplay.LibraryGame) (roleplay.LibraryGame, error) {
	normalized, pack, err := normalizeLibraryGame(game)
	if err != nil {
		return roleplay.LibraryGame{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.games[normalized.ID]; exists {
		return roleplay.LibraryGame{}, errors.New("game already exists")
	}
	s.games[normalized.ID] = normalized
	s.packs[normalized.ID] = pack
	s.gameIDs = append([]string{normalized.ID}, s.gameIDs...)
	if err := s.persistLocked(); err != nil {
		return roleplay.LibraryGame{}, err
	}
	return cloneLibraryGame(normalized), nil
}

func (s *GameService) UpdateGame(gameID string, game roleplay.LibraryGame) (roleplay.LibraryGame, error) {
	if strings.TrimSpace(gameID) == "" {
		return roleplay.LibraryGame{}, errors.New("game id is required")
	}
	game.ID = gameID
	normalized, pack, err := normalizeLibraryGame(game)
	if err != nil {
		return roleplay.LibraryGame{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.games[gameID]; !exists {
		return roleplay.LibraryGame{}, errors.New("game not found")
	}
	s.games[gameID] = normalized
	s.packs[gameID] = pack
	if err := s.persistLocked(); err != nil {
		return roleplay.LibraryGame{}, err
	}
	return cloneLibraryGame(normalized), nil
}

func (s *GameService) DeleteGame(gameID string) error {
	if strings.TrimSpace(gameID) == "" {
		return errors.New("game id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.games[gameID]; !exists {
		return errors.New("game not found")
	}
	delete(s.games, gameID)
	delete(s.packs, gameID)
	s.sessions = sessionsWithoutGameID(s.sessions, gameID)
	s.gameIDs = removeGameID(s.gameIDs, gameID)
	return s.persistLocked()
}

func (s *GameService) StartGame(gameID string) (roleplay.GameTurnResult, error) {
	s.logInfof("start_game requested game=%s", gameID)
	s.mu.Lock()
	pack, ok := s.packs[gameID]
	s.mu.Unlock()
	if !ok {
		s.logErrorf("start_game failed game=%s error=story pack is not registered in backend", gameID)
		return roleplay.GameTurnResult{}, errors.New("story pack is not registered in backend")
	}

	session, err := roleplay.NewGameSession(gameID, pack)
	if err != nil {
		s.logErrorf("start_game session_failed game=%s error=%v", gameID, err)
		return roleplay.GameTurnResult{}, err
	}
	s.logInfof("start_game session_created game=%s session=%s workspace=%s", gameID, session.ID, session.WorkspacePath)

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	return s.advanceSession(session.ID)
}

func cloneLibraryGame(game roleplay.LibraryGame) roleplay.LibraryGame {
	files := make(map[string]string, len(game.Files))
	for name, content := range game.Files {
		files[name] = content
	}
	game.Files = files
	game.PhotoURLs = append([]string{}, game.PhotoURLs...)
	game.MapURLs = append([]string{}, game.MapURLs...)
	return game
}

func normalizeLibraryGame(game roleplay.LibraryGame) (roleplay.LibraryGame, roleplay.StoryPack, error) {
	game.ID = strings.TrimSpace(game.ID)
	if game.ID == "" {
		game.ID = roleplay.NewID("game")
	}
	game.Title = strings.TrimSpace(game.Title)
	if game.Title == "" {
		var metadata roleplay.GameMetadata
		if err := json.Unmarshal([]byte(game.Files[roleplay.MetadataFileName]), &metadata); err == nil {
			game.Title = metadata.GameTitle()
		}
	}
	if game.Title == "" {
		return roleplay.LibraryGame{}, roleplay.StoryPack{}, errors.New("game title is required")
	}
	if game.ImportedAt == "" {
		game.ImportedAt = roleplay.NowISO()
	}
	if game.Files == nil {
		return roleplay.LibraryGame{}, roleplay.StoryPack{}, errors.New("game files are required")
	}

	pack, err := roleplay.NewStoryPack(game.ID, game.Files)
	if err != nil {
		return roleplay.LibraryGame{}, roleplay.StoryPack{}, err
	}
	game.Files = pack.Files
	game.PhotoURLs = append([]string{}, game.PhotoURLs...)
	game.MapURLs = append([]string{}, game.MapURLs...)
	return game, pack, nil
}

func sessionsWithoutGameID(sessions map[string]*roleplay.GameSession, gameID string) map[string]*roleplay.GameSession {
	nextSessions := make(map[string]*roleplay.GameSession, len(sessions))
	for sessionID, session := range sessions {
		if session.GameID != gameID {
			nextSessions[sessionID] = session
		}
	}
	return nextSessions
}

func removeGameID(gameIDs []string, gameID string) []string {
	nextGameIDs := gameIDs[:0]
	for _, currentGameID := range gameIDs {
		if currentGameID != gameID {
			nextGameIDs = append(nextGameIDs, currentGameID)
		}
	}
	return nextGameIDs
}

func (s *GameService) orderedGamesLocked() []roleplay.LibraryGame {
	games := make([]roleplay.LibraryGame, 0, len(s.gameIDs))
	for _, gameID := range s.gameIDs {
		if game, ok := s.games[gameID]; ok {
			games = append(games, cloneLibraryGame(game))
		}
	}
	return games
}

func (s *GameService) persistLocked() error {
	if s.repo == nil {
		return errors.New("game repository is not initialized")
	}
	return s.repo.save(s.orderedGamesLocked())
}

func (s *GameService) SubmitChoice(sessionID string, choiceID string) (roleplay.GameTurnResult, error) {
	s.logInfof("submit_choice requested session=%s choice=%s", sessionID, choiceID)
	s.mu.Lock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		s.logErrorf("submit_choice failed session=%s error=session not found", sessionID)
		return roleplay.GameTurnResult{}, errors.New("session not found")
	}
	if session.State == roleplay.SessionStateEnded {
		result := roleplay.ResultFromSession(session)
		s.mu.Unlock()
		s.logWarningf("submit_choice ignored session=%s choice=%s reason=session already ended", sessionID, choiceID)
		return result, nil
	}

	label := session.ChoiceLabel(choiceID)
	session.AppendTurn(roleplay.GameTurn{
		ID:                  roleplay.NewID("turn"),
		Role:                roleplay.TurnRoleUser,
		Payload:             []string{label},
		SelectedChoiceID:    choiceID,
		SelectedChoiceLabel: label,
		CreatedAt:           roleplay.NowISO(),
	})
	s.mu.Unlock()

	return s.advanceSession(sessionID)
}

func (s *GameService) GetSession(sessionID string) (roleplay.GameSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return roleplay.GameSession{}, errors.New("session not found")
	}

	return session.Clone(), nil
}

func (s *GameService) advanceSession(sessionID string) (roleplay.GameTurnResult, error) {
	s.mu.Lock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		s.logErrorf("advance_session failed session=%s error=session not found", sessionID)
		return roleplay.GameTurnResult{}, errors.New("session not found")
	}
	pack := s.packs[session.GameID]
	config := s.config
	s.mu.Unlock()

	s.logInfof("advance_session start game=%s session=%s state=%s turns=%d model=%s base_url=%s", session.GameID, session.ID, session.State, len(session.Turns), config.AIModel, config.AIBaseURL)
	client := roleplay.NewOpenAIClient(config)
	result, err := roleplay.RunAITurnWithLogger(s.ctx, client, pack, session, s.logInfof)

	s.mu.Lock()
	defer s.mu.Unlock()
	if current, ok := s.sessions[sessionID]; ok {
		*current = *session
	}
	if err != nil {
		s.logErrorf("advance_session failed game=%s session=%s error=%v", session.GameID, session.ID, err)
	} else {
		s.logInfof("advance_session done game=%s session=%s state=%s payload_lines=%d tools=%d ending=%t", session.GameID, session.ID, result.State, len(result.Payload), len(result.Tools), result.Ending != nil)
	}

	return result, err
}

func (s *GameService) logInfof(format string, args ...interface{}) {
	if s.ctx == nil {
		return
	}
	wailsruntime.LogInfof(s.ctx, format, args...)
}

func (s *GameService) logWarningf(format string, args ...interface{}) {
	if s.ctx == nil {
		return
	}
	wailsruntime.LogWarningf(s.ctx, format, args...)
}

func (s *GameService) logErrorf(format string, args ...interface{}) {
	if s.ctx == nil {
		return
	}
	wailsruntime.LogErrorf(s.ctx, format, args...)
}
