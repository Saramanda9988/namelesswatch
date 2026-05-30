package service

import (
	"context"
	"errors"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"sync"
)

type GameService struct {
	ctx      context.Context
	mu       sync.Mutex
	config   appconf.AppConfig
	packs    map[string]roleplay.StoryPack
	sessions map[string]*roleplay.GameSession
	games    map[string]roleplay.LibraryGame
	gameIDs  []string
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

	pack, err := roleplay.NewStoryPack(gameID, files)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.packs[gameID] = pack
	return nil
}

func (s *GameService) ImportGamePack(files map[string]string) (roleplay.ImportGameResult, error) {
	game, result, err := roleplay.NewLibraryGame(files)
	if err != nil {
		return result, err
	}
	if result.Game == nil {
		return result, nil
	}

	pack, err := roleplay.NewStoryPack(game.ID, game.Files)
	if err != nil {
		return roleplay.ImportGameResult{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[game.ID] = game
	s.gameIDs = append([]string{game.ID}, s.gameIDs...)
	s.packs[game.ID] = pack

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

func (s *GameService) StartGame(gameID string) (roleplay.GameTurnResult, error) {
	s.mu.Lock()
	pack, ok := s.packs[gameID]
	s.mu.Unlock()
	if !ok {
		return roleplay.GameTurnResult{}, errors.New("story pack is not registered in backend")
	}

	session, err := roleplay.NewGameSession(gameID, pack)
	if err != nil {
		return roleplay.GameTurnResult{}, err
	}

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

func (s *GameService) SubmitChoice(sessionID string, choiceID string) (roleplay.GameTurnResult, error) {
	s.mu.Lock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return roleplay.GameTurnResult{}, errors.New("session not found")
	}
	if session.State == roleplay.SessionStateEnded {
		result := roleplay.ResultFromSession(session)
		s.mu.Unlock()
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
		return roleplay.GameTurnResult{}, errors.New("session not found")
	}
	pack := s.packs[session.GameID]
	config := s.config
	s.mu.Unlock()

	client := roleplay.NewOpenAIClient(config)
	result, err := roleplay.RunAITurn(s.ctx, client, pack, session)

	s.mu.Lock()
	defer s.mu.Unlock()
	if current, ok := s.sessions[sessionID]; ok {
		*current = *session
	}

	return result, err
}
