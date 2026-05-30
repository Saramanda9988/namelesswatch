package main

import (
	"context"
	"errors"
	"sync"
)

// App struct
type App struct {
	ctx      context.Context
	mu       sync.Mutex
	packs    map[string]StoryPack
	sessions map[string]*GameSession
	client   *OpenAIClient
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		packs:    make(map[string]StoryPack),
		sessions: make(map[string]*GameSession),
		client:   NewOpenAIClientFromEnv(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// RegisterGamePack stores frontend-imported story pack documents for backend sessions.
func (a *App) RegisterGamePack(gameID string, files map[string]string) error {
	if gameID == "" {
		return errors.New("game id is required")
	}

	pack, err := NewStoryPack(gameID, files)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.packs[gameID] = pack
	return nil
}

// StartGame creates a new AI roleplay session and returns the first AI turn.
func (a *App) StartGame(gameID string) (GameTurnResult, error) {
	a.mu.Lock()
	pack, ok := a.packs[gameID]
	a.mu.Unlock()
	if !ok {
		return GameTurnResult{}, errors.New("story pack is not registered in backend")
	}

	session, err := NewGameSession(gameID, pack)
	if err != nil {
		return GameTurnResult{}, err
	}

	a.mu.Lock()
	a.sessions[session.ID] = session
	a.mu.Unlock()

	return a.advanceSession(session.ID)
}

// SubmitChoice records a user choice and advances the current AI session.
func (a *App) SubmitChoice(sessionID string, choiceID string) (GameTurnResult, error) {
	a.mu.Lock()
	session, ok := a.sessions[sessionID]
	if !ok {
		a.mu.Unlock()
		return GameTurnResult{}, errors.New("session not found")
	}
	if session.State == SessionStateEnded {
		result := ResultFromSession(session)
		a.mu.Unlock()
		return result, nil
	}

	label := session.ChoiceLabel(choiceID)
	session.AppendTurn(GameTurn{
		ID:                  NewID("turn"),
		Role:                TurnRoleUser,
		Payload:             []string{label},
		SelectedChoiceID:    choiceID,
		SelectedChoiceLabel: label,
		CreatedAt:           NowISO(),
	})
	a.mu.Unlock()

	return a.advanceSession(sessionID)
}

// GetSession returns current runtime session state.
func (a *App) GetSession(sessionID string) (GameSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, ok := a.sessions[sessionID]
	if !ok {
		return GameSession{}, errors.New("session not found")
	}

	return session.Clone(), nil
}

func (a *App) advanceSession(sessionID string) (GameTurnResult, error) {
	a.mu.Lock()
	session, ok := a.sessions[sessionID]
	if !ok {
		a.mu.Unlock()
		return GameTurnResult{}, errors.New("session not found")
	}
	pack := a.packs[session.GameID]
	a.mu.Unlock()

	result, err := RunAITurn(a.ctx, a.client, pack, session)

	a.mu.Lock()
	defer a.mu.Unlock()
	if current, ok := a.sessions[sessionID]; ok {
		*current = *session
	}

	return result, err
}
