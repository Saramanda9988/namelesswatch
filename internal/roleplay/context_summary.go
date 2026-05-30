package roleplay

import (
	"context"
	"errors"
	"fmt"
	"namelesswatch/internal/appconf"
	"os"
	"path/filepath"
	"strings"
)

const ContextSummaryFileName = "context_summary.md"

type ContextBudget struct {
	RecentTurnLimit        int
	CompactTurnThreshold   int
	SoftPromptRuneBudget   int
	HardPromptRuneBudget   int
	StoryFileRuneBudget    int
	SummaryRuneBudget      int
	MemoryRuneBudget       int
	TerminalResultRunes    int
	RepairInstructionRunes int
}

func DefaultContextBudget() ContextBudget {
	return ContextBudget{
		RecentTurnLimit:        appconf.DefaultAIContextRecentTurns,
		CompactTurnThreshold:   appconf.DefaultAIContextCompactTurns,
		SoftPromptRuneBudget:   appconf.DefaultAIContextSoftBudget,
		HardPromptRuneBudget:   appconf.DefaultAIContextHardBudget,
		StoryFileRuneBudget:    40000,
		SummaryRuneBudget:      12000,
		MemoryRuneBudget:       18000,
		TerminalResultRunes:    6000,
		RepairInstructionRunes: 8000,
	}
}

func ContextBudgetFromConfig(config appconf.AppConfig) ContextBudget {
	appconf.Normalize(&config)
	budget := DefaultContextBudget()
	budget.RecentTurnLimit = config.AIContextRecentTurns
	budget.CompactTurnThreshold = config.AIContextCompactTurns
	budget.SoftPromptRuneBudget = config.AIContextSoftBudget
	budget.HardPromptRuneBudget = config.AIContextHardBudget
	return normalizeContextBudget(budget)
}

func normalizeContextBudget(budget ContextBudget) ContextBudget {
	defaults := DefaultContextBudget()
	if budget.RecentTurnLimit <= 0 {
		budget.RecentTurnLimit = defaults.RecentTurnLimit
	}
	if budget.CompactTurnThreshold <= budget.RecentTurnLimit {
		budget.CompactTurnThreshold = budget.RecentTurnLimit * 2
	}
	if budget.SoftPromptRuneBudget <= 0 {
		budget.SoftPromptRuneBudget = defaults.SoftPromptRuneBudget
	}
	if budget.HardPromptRuneBudget < budget.SoftPromptRuneBudget {
		budget.HardPromptRuneBudget = budget.SoftPromptRuneBudget
	}
	if budget.StoryFileRuneBudget <= 0 {
		budget.StoryFileRuneBudget = defaults.StoryFileRuneBudget
	}
	if budget.SummaryRuneBudget <= 0 {
		budget.SummaryRuneBudget = defaults.SummaryRuneBudget
	}
	if budget.MemoryRuneBudget <= 0 {
		budget.MemoryRuneBudget = defaults.MemoryRuneBudget
	}
	if budget.TerminalResultRunes <= 0 {
		budget.TerminalResultRunes = defaults.TerminalResultRunes
	}
	if budget.RepairInstructionRunes <= 0 {
		budget.RepairInstructionRunes = defaults.RepairInstructionRunes
	}
	return budget
}

func ContextSummaryPath(session *GameSession) string {
	if session == nil || strings.TrimSpace(session.WorkspacePath) == "" {
		return ""
	}
	return filepath.Join(session.WorkspacePath, ContextSummaryFileName)
}

func DefaultContextSummary() string {
	return strings.TrimSpace(`## 当前阶段
- 未记录

## 关键事实
- 未记录

## 用户选择
- 未记录

## 规则后果
- 未记录

## 未解决线索
- 未记录

## 结局倾向
- 未记录`) + "\n"
}

func EnsureContextSummary(session *GameSession) error {
	path := ContextSummaryPath(session)
	if path == "" {
		return errors.New("session workspace is required")
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat context summary: %w", err)
	}
	if err := os.WriteFile(path, []byte(DefaultContextSummary()), 0o600); err != nil {
		return fmt.Errorf("initialize context summary: %w", err)
	}
	return nil
}

func ReadContextSummary(session *GameSession) (string, error) {
	path := ContextSummaryPath(session)
	if path == "" {
		return DefaultContextSummary(), errors.New("session workspace is required")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultContextSummary(), nil
	}
	if err != nil {
		return DefaultContextSummary(), fmt.Errorf("read context summary: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return DefaultContextSummary(), nil
	}
	return string(data), nil
}

func WriteContextSummary(session *GameSession, summary string) error {
	path := ContextSummaryPath(session)
	if path == "" {
		return errors.New("session workspace is required")
	}
	value := strings.TrimSpace(summary)
	if value == "" {
		value = DefaultContextSummary()
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(value+"\n"), 0o600); err != nil {
		return fmt.Errorf("write context summary temp file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace context summary: %w", err)
	}
	return nil
}

func ShouldCompactSessionContext(session *GameSession, budget ContextBudget) bool {
	if session == nil {
		return false
	}
	budget = normalizeContextBudget(budget)
	if len(session.Turns) > budget.CompactTurnThreshold {
		return true
	}
	if memory, err := ReadWorkspaceFile(session, "memory.md"); err == nil && runeLen(memory) > budget.MemoryRuneBudget {
		return true
	}
	if summary, err := ReadContextSummary(session); err == nil && runeLen(summary) > budget.SummaryRuneBudget*2 {
		return true
	}
	return false
}

func CompactSessionContext(ctx context.Context, client ChatCompleter, session *GameSession, budget ContextBudget, logf TurnLogger) error {
	if session == nil {
		return errors.New("session is required")
	}
	budget = normalizeContextBudget(budget)
	if err := EnsureContextSummary(session); err != nil {
		return err
	}

	turns := turnsToCompact(session.Turns, budget.RecentTurnLimit)
	if len(turns) == 0 {
		return nil
	}
	previousSummary, err := ReadContextSummary(session)
	if err != nil {
		logTurn(logf, "context_compaction read_summary_failed session=%s error=%v", session.ID, err)
		previousSummary = DefaultContextSummary()
	}
	content, err := client.Chat(ctx, buildContextCompactionMessages(previousSummary, turns, budget))
	if err != nil {
		return fmt.Errorf("compact context summary: %w", err)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("compact context summary returned empty content")
	}
	if err := WriteContextSummary(session, content); err != nil {
		return err
	}
	logTurn(logf, "context_compaction done session=%s compacted_turns=%d summary_runes=%d", session.ID, len(turns), runeLen(content))
	return nil
}

func buildContextCompactionMessages(previousSummary string, turns []GameTurn, budget ContextBudget) []ChatMessage {
	var builder strings.Builder
	builder.WriteString("请把旧摘要和旧回合合并为新的 context_summary.md。\n")
	builder.WriteString("只输出 Markdown 摘要，不要输出解释，不要泄露 true.md 原文。\n")
	builder.WriteString("必须保留以下标题：当前阶段、关键事实、用户选择、规则后果、未解决线索、结局倾向。\n\n")
	builder.WriteString("--- 旧摘要 ---\n")
	builder.WriteString(limitRunes(previousSummary, budget.SummaryRuneBudget))
	builder.WriteString("\n\n--- 待压缩旧回合 ---\n")
	for _, turn := range turns {
		builder.WriteString(formatTurnForPrompt(turn))
		builder.WriteString("\n")
	}

	return []ChatMessage{
		{Role: "system", Content: "你是游戏会话上下文压缩器，负责把旧回合压缩成稳定、简洁、可继续推理的中文摘要。"},
		{Role: "user", Content: builder.String()},
	}
}

func turnsToCompact(turns []GameTurn, recentLimit int) []GameTurn {
	if recentLimit <= 0 || len(turns) <= recentLimit {
		return nil
	}
	return turns[:len(turns)-recentLimit]
}

func EstimatePromptRunes(messages []ChatMessage) int {
	total := 0
	for _, message := range messages {
		total += runeLen(message.Role)
		total += runeLen(message.Content)
	}
	return total
}

func limitRunes(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "\n[truncated]"
}

func runeLen(value string) int {
	return len([]rune(value))
}
