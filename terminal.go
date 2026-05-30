package main

import (
	"bytes"
	"context"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const terminalTimeout = 5 * time.Second
const terminalOutputLimit = 6000

func ExecuteTerminalRequest(session *GameSession, request AgentTerminalRequest) []TerminalExecution {
	results := make([]TerminalExecution, 0, len(request.Commands))
	for _, command := range request.Commands {
		results = append(results, ExecuteTerminalCommand(session.WorkspacePath, command.Command))
	}
	return results
}

func ExecuteTerminalCommand(workdir string, command string) TerminalExecution {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), terminalTimeout)
	defer cancel()

	cmd := shellCommand(ctx, command)
	cmd.Dir = workdir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	stdoutText, stdoutTruncated := truncateOutput(stdout.String())
	stderrText, stderrTruncated := truncateOutput(stderr.String())
	result := TerminalExecution{
		Command:     command,
		Stdout:      stdoutText,
		Stderr:      stderrText,
		ExitCode:    exitCode,
		DurationMS:  duration.Milliseconds(),
		TimedOut:    ctx.Err() == context.DeadlineExceeded,
		Truncated:   stdoutTruncated || stderrTruncated,
		CompletedAt: NowISO(),
	}
	if err != nil {
		result.Error = err.Error()
	}

	log.Printf("agent_terminal command=%q exit=%d duration=%s stdout=%q stderr=%q", result.Command, result.ExitCode, duration.String(), result.Stdout, result.Stderr)
	return result
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func truncateOutput(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= terminalOutputLimit {
		return value, false
	}
	runes := []rune(value)
	return string(runes[:terminalOutputLimit]) + "\n[truncated]", true
}
