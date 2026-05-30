package service

import (
	"context"
	"fmt"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type chatClientFactory func(appconf.AppConfig) roleplay.ChatCompleter

type prefetchConfig struct {
	Enabled            bool
	GlobalConcurrency  int
	SessionConcurrency int
	TTL                time.Duration
	SubmitWait         time.Duration
	ContextBudget      roleplay.ContextBudget
}

type prefetchKey struct {
	sessionID  string
	baseTurnID string
	choiceID   string
}

type prefetchJob struct {
	key       prefetchKey
	cancel    context.CancelFunc
	done      chan struct{}
	createdAt time.Time

	running   bool
	completed bool
	discarded bool
	workspace string
	branch    *roleplay.GameSession
	result    roleplay.GameTurnResult
	err       error
}

type prefetchOutcome struct {
	workspace string
	branch    *roleplay.GameSession
	result    roleplay.GameTurnResult
}

type prefetchManager struct {
	mu             sync.Mutex
	jobs           map[prefetchKey]*prefetchJob
	running        int
	sessionRunning map[string]int
}

func newPrefetchManager() *prefetchManager {
	return &prefetchManager{
		jobs:           make(map[prefetchKey]*prefetchJob),
		sessionRunning: make(map[string]int),
	}
}

func prefetchConfigFromAppConfig(config appconf.AppConfig) prefetchConfig {
	appconf.Normalize(&config)
	return prefetchConfig{
		Enabled:            config.AIChoicePrefetchEnabled,
		GlobalConcurrency:  config.AIChoicePrefetchGlobalConcurrency,
		SessionConcurrency: config.AIChoicePrefetchSessionConcurrency,
		TTL:                time.Duration(config.AIChoicePrefetchTTLMS) * time.Millisecond,
		SubmitWait:         time.Duration(config.AIChoicePrefetchWaitMS) * time.Millisecond,
		ContextBudget:      roleplay.ContextBudgetFromConfig(config),
	}
}

func (m *prefetchManager) startChoices(ctx context.Context, config prefetchConfig, modelConfig appconf.AppConfig, clientFactory chatClientFactory, pack roleplay.StoryPack, base roleplay.GameSession, baseTurn roleplay.GameTurn, options []roleplay.ChoiceOption, logf roleplay.TurnLogger) {
	if m == nil || !config.Enabled || len(options) == 0 || base.State == roleplay.SessionStateEnded {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	m.cleanupExpired(config.TTL)
	m.cancelSessionExceptBase(base.ID, baseTurn.ID)

	for _, option := range options {
		key := prefetchKey{sessionID: base.ID, baseTurnID: baseTurn.ID, choiceID: option.ID}
		jobCtx, cancel := context.WithTimeout(ctx, config.TTL)
		job := &prefetchJob{
			key:       key,
			cancel:    cancel,
			done:      make(chan struct{}),
			createdAt: time.Now(),
			running:   true,
		}

		m.mu.Lock()
		if _, exists := m.jobs[key]; exists {
			m.mu.Unlock()
			cancel()
			continue
		}
		if m.running >= config.GlobalConcurrency || m.sessionRunning[base.ID] >= config.SessionConcurrency {
			m.mu.Unlock()
			cancel()
			logPrefetch(logf, "choice_prefetch skipped session=%s base_turn=%s choice=%s reason=concurrency_limit", base.ID, baseTurn.ID, option.ID)
			continue
		}
		m.jobs[key] = job
		m.running++
		m.sessionRunning[base.ID]++
		m.mu.Unlock()

		baseCopy := base.Clone()
		go m.runJob(jobCtx, job, modelConfig, clientFactory, pack, baseCopy, option, config.ContextBudget, logf)
	}
}

func (m *prefetchManager) runJob(ctx context.Context, job *prefetchJob, modelConfig appconf.AppConfig, clientFactory chatClientFactory, pack roleplay.StoryPack, base roleplay.GameSession, option roleplay.ChoiceOption, budget roleplay.ContextBudget, logf roleplay.TurnLogger) {
	var (
		branch    *roleplay.GameSession
		workspace string
		result    roleplay.GameTurnResult
		err       error
	)
	defer func() {
		m.finishJob(job, workspace, branch, result, err)
	}()

	branch, workspace, err = createPrefetchBranch(base, option)
	if err != nil {
		return
	}
	client := clientFactory(modelConfig)
	result, err = roleplay.RunAITurnWithOptions(ctx, client, pack, branch, roleplay.AITurnOptions{ContextBudget: budget}, logf)
	if err != nil {
		err = fmt.Errorf("prefetch choice %q: %w", option.ID, err)
		return
	}
	logPrefetch(logf, "choice_prefetch done session=%s base_turn=%s choice=%s state=%s", job.key.sessionID, job.key.baseTurnID, option.ID, result.State)
}

func (m *prefetchManager) finishJob(job *prefetchJob, workspace string, branch *roleplay.GameSession, result roleplay.GameTurnResult, err error) {
	m.mu.Lock()
	if job.running {
		job.running = false
		m.running--
		m.sessionRunning[job.key.sessionID]--
		if m.sessionRunning[job.key.sessionID] <= 0 {
			delete(m.sessionRunning, job.key.sessionID)
		}
	}
	job.workspace = workspace
	job.branch = branch
	job.result = result
	job.err = err
	job.completed = true
	discarded := job.discarded
	close(job.done)
	m.mu.Unlock()

	if err != nil || discarded {
		_ = os.RemoveAll(workspace)
	}
}

func createPrefetchBranch(base roleplay.GameSession, option roleplay.ChoiceOption) (*roleplay.GameSession, string, error) {
	root := filepath.Join(filepath.Dir(base.WorkspacePath), ".prefetch")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, "", fmt.Errorf("create prefetch root: %w", err)
	}
	workspace, err := os.MkdirTemp(root, safePrefetchSegment(option.ID)+"-*")
	if err != nil {
		return nil, "", fmt.Errorf("create prefetch workspace: %w", err)
	}
	if err := copyDir(base.WorkspacePath, workspace); err != nil {
		_ = os.RemoveAll(workspace)
		return nil, "", fmt.Errorf("copy prefetch workspace: %w", err)
	}

	branch := base.Clone()
	branch.WorkspacePath = workspace
	branch.MemoryPath = filepath.Join(workspace, "memory.md")
	label := branch.ChoiceLabel(option.ID)
	if label == option.ID && option.Label != "" {
		label = option.Label
	}
	branch.AppendTurn(roleplay.GameTurn{
		ID:                  roleplay.NewID("turn"),
		Role:                roleplay.TurnRoleUser,
		Payload:             []string{label},
		SelectedChoiceID:    option.ID,
		SelectedChoiceLabel: label,
		CreatedAt:           roleplay.NowISO(),
	})
	return &branch, workspace, nil
}

func (m *prefetchManager) waitForResult(ctx context.Context, config prefetchConfig, sessionID, baseTurnID, choiceID string) (*prefetchOutcome, bool) {
	if m == nil || !config.Enabled {
		return nil, false
	}
	if ctx == nil {
		ctx = context.Background()
	}
	m.cleanupExpired(config.TTL)

	key := prefetchKey{sessionID: sessionID, baseTurnID: baseTurnID, choiceID: choiceID}
	m.mu.Lock()
	job, ok := m.jobs[key]
	m.mu.Unlock()
	if !ok {
		return nil, false
	}

	if !waitJob(ctx, job.done, config.SubmitWait) {
		return nil, false
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.jobs, key)
	if job.err != nil || job.branch == nil {
		job.discarded = true
		_ = os.RemoveAll(job.workspace)
		return nil, false
	}
	return &prefetchOutcome{
		workspace: job.workspace,
		branch:    job.branch,
		result:    job.result,
	}, true
}

func waitJob(ctx context.Context, done <-chan struct{}, wait time.Duration) bool {
	if wait <= 0 {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	case <-timer.C:
		return false
	}
}

func (m *prefetchManager) cancelSession(sessionID string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	jobs := make([]*prefetchJob, 0)
	for key, job := range m.jobs {
		if key.sessionID == sessionID {
			delete(m.jobs, key)
			job.discarded = true
			jobs = append(jobs, job)
		}
	}
	m.mu.Unlock()
	discardPrefetchJobs(jobs)
}

func (m *prefetchManager) cancelSessionExceptBase(sessionID string, baseTurnID string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	jobs := make([]*prefetchJob, 0)
	for key, job := range m.jobs {
		if key.sessionID == sessionID && key.baseTurnID != baseTurnID {
			delete(m.jobs, key)
			job.discarded = true
			jobs = append(jobs, job)
		}
	}
	m.mu.Unlock()
	discardPrefetchJobs(jobs)
}

func (m *prefetchManager) cancelBaseExceptChoice(sessionID string, baseTurnID string, choiceID string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	jobs := make([]*prefetchJob, 0)
	for key, job := range m.jobs {
		if key.sessionID == sessionID && key.baseTurnID == baseTurnID && key.choiceID != choiceID {
			delete(m.jobs, key)
			job.discarded = true
			jobs = append(jobs, job)
		}
	}
	m.mu.Unlock()
	discardPrefetchJobs(jobs)
}

func (m *prefetchManager) cleanupExpired(ttl time.Duration) {
	if m == nil || ttl <= 0 {
		return
	}
	cutoff := time.Now().Add(-ttl)
	m.mu.Lock()
	jobs := make([]*prefetchJob, 0)
	for key, job := range m.jobs {
		if job.createdAt.Before(cutoff) {
			delete(m.jobs, key)
			job.discarded = true
			jobs = append(jobs, job)
		}
	}
	m.mu.Unlock()
	discardPrefetchJobs(jobs)
}

func discardPrefetchJobs(jobs []*prefetchJob) {
	for _, job := range jobs {
		job.cancel()
		if job.completed {
			_ = os.RemoveAll(job.workspace)
		}
	}
}

func promotePrefetchBranch(current *roleplay.GameSession, outcome *prefetchOutcome) (roleplay.GameTurnResult, error) {
	if current == nil || outcome == nil || outcome.branch == nil {
		return roleplay.GameTurnResult{}, fmt.Errorf("prefetch result is empty")
	}
	original := current.Clone()
	promoted := outcome.branch.Clone()
	promoted.ID = original.ID
	promoted.GameID = original.GameID
	promoted.Label = original.Label
	promoted.IsSnapshot = original.IsSnapshot
	promoted.ParentID = original.ParentID
	promoted.CreatedAt = original.CreatedAt
	promoted.WorkspacePath = outcome.workspace
	promoted.MemoryPath = filepath.Join(outcome.workspace, "memory.md")
	*current = promoted

	result := outcome.result
	result.SessionID = current.ID
	result.GameID = current.GameID
	result.State = current.State
	result.CurrentBGMID = current.CurrentBGMID
	return result, nil
}

func discardPrefetchOutcome(outcome *prefetchOutcome) {
	if outcome != nil {
		_ = os.RemoveAll(outcome.workspace)
	}
}

func logPrefetch(logf roleplay.TurnLogger, format string, args ...interface{}) {
	if logf != nil {
		logf(format, args...)
	}
}

func safePrefetchSegment(value string) string {
	if value == "" {
		return "choice"
	}
	clean := make([]rune, 0, len(value))
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			clean = append(clean, r)
		case r >= 'A' && r <= 'Z':
			clean = append(clean, r)
		case r >= '0' && r <= '9':
			clean = append(clean, r)
		case r == '-' || r == '_':
			clean = append(clean, r)
		default:
			clean = append(clean, '_')
		}
	}
	if len(clean) == 0 {
		return "choice"
	}
	return string(clean)
}

func replaceDir(src, dst string) error {
	parent := filepath.Dir(dst)
	temp, err := os.MkdirTemp(parent, ".replace-*")
	if err != nil {
		return fmt.Errorf("create replacement directory: %w", err)
	}
	if err := copyDir(src, temp); err != nil {
		_ = os.RemoveAll(temp)
		return fmt.Errorf("copy replacement directory: %w", err)
	}
	if err := os.RemoveAll(dst); err != nil {
		_ = os.RemoveAll(temp)
		return fmt.Errorf("remove old workspace: %w", err)
	}
	if err := os.Rename(temp, dst); err != nil {
		_ = os.RemoveAll(temp)
		return fmt.Errorf("replace workspace: %w", err)
	}
	return nil
}
