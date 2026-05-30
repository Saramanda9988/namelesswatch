package roleplay

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
	return ExecuteTerminalRequestWithLogger(session, request, nil)
}

func ExecuteTerminalRequestWithLogger(session *GameSession, request AgentTerminalRequest, logf TurnLogger) []TerminalExecution {
	results := make([]TerminalExecution, 0, len(request.Commands))
	for _, command := range request.Commands {
		results = append(results, ExecuteTerminalCommandWithLogger(session.WorkspacePath, command.Command, logf))
	}
	return results
}

func ExecuteTerminalCommand(workdir string, command string) TerminalExecution {
	return ExecuteTerminalCommandWithLogger(workdir, command, nil)
}

func ExecuteTerminalCommandWithLogger(workdir string, command string, logf TurnLogger) TerminalExecution {
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
	logTurn(logf, "agent_terminal command=%q exit=%d timed_out=%t duration=%s stdout=%q stderr=%q", result.Command, result.ExitCode, result.TimedOut, duration.String(), truncateLogValue(result.Stdout, 2000), truncateLogValue(result.Stderr, 2000))
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
