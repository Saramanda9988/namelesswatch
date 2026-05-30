package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"sort"
)

const sessionFileVersion = 1

const sessionFileName = "session.json"

type sessionFile struct {
	Version int                  `json:"version"`
	Session roleplay.GameSession `json:"session"`
}

type sessionRepository struct {
	root string
}

func newSessionRepository() (*sessionRepository, error) {
	configDir, err := appconf.GetConfigDir()
	if err != nil {
		return nil, err
	}
	return &sessionRepository{root: filepath.Join(configDir, "sessions")}, nil
}

func (r *sessionRepository) sessionDir(sessionID string) string {
	return filepath.Join(r.root, sessionID)
}

func (r *sessionRepository) save(session *roleplay.GameSession) error {
	if session == nil || session.ID == "" {
		return errors.New("session id is required")
	}
	dir := r.sessionDir(session.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create session directory: %w", err)
	}

	data, err := json.MarshalIndent(sessionFile{
		Version: sessionFileVersion,
		Session: session.Clone(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}

	path := filepath.Join(dir, sessionFileName)
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("write session temp file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace session file: %w", err)
	}
	return nil
}

func (r *sessionRepository) load(sessionID string) (*roleplay.GameSession, error) {
	if sessionID == "" {
		return nil, errors.New("session id is required")
	}
	path := filepath.Join(r.sessionDir(sessionID), sessionFileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, os.ErrNotExist
	}
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}

	var file sessionFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	session := file.Session
	return &session, nil
}

// list returns all stored sessions, optionally filtered by gameID (empty = all),
// sorted by UpdatedAt descending (most recent first).
func (r *sessionRepository) list(gameID string) ([]roleplay.GameSession, error) {
	entries, err := os.ReadDir(r.root)
	if errors.Is(err, os.ErrNotExist) {
		return []roleplay.GameSession{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read sessions directory: %w", err)
	}

	sessions := make([]roleplay.GameSession, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		session, err := r.load(entry.Name())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if gameID != "" && session.GameID != gameID {
			continue
		}
		sessions = append(sessions, *session)
	}

	sort.SliceStable(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt > sessions[j].UpdatedAt
	})
	return sessions, nil
}

func (r *sessionRepository) delete(sessionID string) error {
	if sessionID == "" {
		return errors.New("session id is required")
	}
	if err := os.RemoveAll(r.sessionDir(sessionID)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// snapshot makes a frozen copy of src under a new session id, copying the
// workspace directory so the snapshot has its own memory.md.
func (r *sessionRepository) snapshot(src *roleplay.GameSession, label string) (*roleplay.GameSession, error) {
	if src == nil {
		return nil, errors.New("source session is required")
	}

	snap := src.Clone()
	snap.ID = roleplay.NewID("session")
	snap.IsSnapshot = true
	snap.ParentID = src.ID
	snap.Label = label
	snap.CreatedAt = roleplay.NowISO()
	snap.UpdatedAt = snap.CreatedAt

	dstWorkspace := filepath.Join(r.sessionDir(snap.ID), "workspace")
	if err := copyDir(src.WorkspacePath, dstWorkspace); err != nil {
		return nil, fmt.Errorf("copy snapshot workspace: %w", err)
	}
	snap.WorkspacePath = dstWorkspace
	snap.MemoryPath = filepath.Join(dstWorkspace, "memory.md")

	if err := r.save(&snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// fork creates a fresh playable session from src (typically a snapshot),
// copying the workspace into a new session dir and clearing snapshot flags.
func (r *sessionRepository) fork(src *roleplay.GameSession) (*roleplay.GameSession, error) {
	if src == nil {
		return nil, errors.New("source session is required")
	}

	forked := src.Clone()
	forked.ID = roleplay.NewID("session")
	forked.IsSnapshot = false
	forked.ParentID = src.ID
	forked.Label = ""
	forked.UpdatedAt = roleplay.NowISO()

	dstWorkspace := filepath.Join(r.sessionDir(forked.ID), "workspace")
	if err := copyDir(src.WorkspacePath, dstWorkspace); err != nil {
		return nil, fmt.Errorf("copy forked workspace: %w", err)
	}
	forked.WorkspacePath = dstWorkspace
	forked.MemoryPath = filepath.Join(dstWorkspace, "memory.md")

	if err := r.save(&forked); err != nil {
		return nil, err
	}
	return &forked, nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
