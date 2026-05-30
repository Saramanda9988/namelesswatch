package service

import (
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestSession(t *testing.T, root, gameID string) *roleplay.GameSession {
	t.Helper()

	pack, err := roleplay.NewStoryPack(gameID, testGameFiles(t))
	if err != nil {
		t.Fatalf("new story pack: %v", err)
	}
	session, err := roleplay.NewGameSessionInDir(gameID, pack, root)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.AppendTurn(roleplay.GameTurn{
		ID:        roleplay.NewID("turn"),
		Role:      roleplay.TurnRoleAI,
		Payload:   []string{"开场叙事。"},
		CreatedAt: roleplay.NowISO(),
	})
	return session
}

func TestSessionRepositorySaveLoadRoundTrip(t *testing.T) {
	repo := &sessionRepository{root: t.TempDir()}
	session := newTestSession(t, repo.root, "game-a")

	if err := repo.save(session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	loaded, err := repo.load(session.ID)
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if loaded.ID != session.ID || loaded.GameID != "game-a" {
		t.Fatalf("unexpected loaded session: %#v", loaded)
	}
	if len(loaded.Turns) != 1 || loaded.Turns[0].Payload[0] != "开场叙事。" {
		t.Fatalf("turns not persisted: %#v", loaded.Turns)
	}
}

func TestSessionRepositoryListFiltersByGame(t *testing.T) {
	repo := &sessionRepository{root: t.TempDir()}
	a := newTestSession(t, repo.root, "game-a")
	b := newTestSession(t, repo.root, "game-b")
	if err := repo.save(a); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := repo.save(b); err != nil {
		t.Fatalf("save b: %v", err)
	}

	all, err := repo.list("")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(all))
	}

	onlyA, err := repo.list("game-a")
	if err != nil {
		t.Fatalf("list game-a: %v", err)
	}
	if len(onlyA) != 1 || onlyA[0].GameID != "game-a" {
		t.Fatalf("unexpected filtered list: %#v", onlyA)
	}
}

func TestSessionRepositorySnapshotIsIndependentCopy(t *testing.T) {
	repo := &sessionRepository{root: t.TempDir()}
	session := newTestSession(t, repo.root, "game-a")
	if err := roleplay.WriteContextSummary(session, "## 当前阶段\n- source summary"); err != nil {
		t.Fatalf("write source summary: %v", err)
	}
	if err := repo.save(session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	snap, err := repo.snapshot(session, "存档点")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if !snap.IsSnapshot || snap.Label != "存档点" || snap.ParentID != session.ID {
		t.Fatalf("unexpected snapshot meta: %#v", snap)
	}
	if snap.ID == session.ID {
		t.Fatalf("snapshot should get a new id")
	}

	// Snapshot must have its own workspace copy with memory.md.
	if _, err := os.Stat(filepath.Join(snap.WorkspacePath, "memory.md")); err != nil {
		t.Fatalf("snapshot workspace missing memory.md: %v", err)
	}
	snapSummary, err := roleplay.ReadContextSummary(snap)
	if err != nil {
		t.Fatalf("read snapshot context summary: %v", err)
	}
	if !strings.Contains(snapSummary, "source summary") {
		t.Fatalf("snapshot did not copy context summary, got:\n%s", snapSummary)
	}
	if snap.WorkspacePath == session.WorkspacePath {
		t.Fatalf("snapshot workspace must differ from source")
	}

	listed, err := repo.list("game-a")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected source + snapshot, got %d", len(listed))
	}
}

func TestSessionRepositoryForkCopiesContextSummary(t *testing.T) {
	repo := &sessionRepository{root: t.TempDir()}
	session := newTestSession(t, repo.root, "game-a")
	if err := roleplay.WriteContextSummary(session, "## 当前阶段\n- snapshot summary"); err != nil {
		t.Fatalf("write source summary: %v", err)
	}
	snap, err := repo.snapshot(session, "存档点")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	forked, err := repo.fork(snap)
	if err != nil {
		t.Fatalf("fork: %v", err)
	}
	if forked.IsSnapshot || forked.ParentID != snap.ID || forked.ID == snap.ID {
		t.Fatalf("unexpected fork metadata: %#v", forked)
	}
	forkSummary, err := roleplay.ReadContextSummary(forked)
	if err != nil {
		t.Fatalf("read fork context summary: %v", err)
	}
	if !strings.Contains(forkSummary, "snapshot summary") {
		t.Fatalf("fork did not copy context summary, got:\n%s", forkSummary)
	}
	if forked.WorkspacePath == snap.WorkspacePath {
		t.Fatal("fork workspace must differ from snapshot")
	}
}

func TestSessionRepositoryDeleteRemovesDir(t *testing.T) {
	repo := &sessionRepository{root: t.TempDir()}
	session := newTestSession(t, repo.root, "game-a")
	if err := repo.save(session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	if err := repo.delete(session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := repo.load(session.ID); !os.IsNotExist(err) {
		t.Fatalf("expected session to be gone, got err=%v", err)
	}
	if _, err := os.Stat(repo.sessionDir(session.ID)); !os.IsNotExist(err) {
		t.Fatalf("session directory should be removed")
	}
	if _, err := os.Stat(session.WorkspacePath); !os.IsNotExist(err) {
		t.Fatalf("session workspace should be removed")
	}
}
