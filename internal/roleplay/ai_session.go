package roleplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxTerminalRounds = 3

func RunAITurn(ctx context.Context, client ChatCompleter, pack StoryPack, session *GameSession) (GameTurnResult, error) {
	var terminalResults []TerminalExecution

	for terminalRound := 0; terminalRound <= maxTerminalRounds; terminalRound++ {
		content, err := client.Chat(ctx, BuildMessages(pack, session, terminalResults, ""))
		if err != nil {
			return GameTurnResult{}, err
		}

		response, terminalRequest, err := ParseAIResponse(content)
		if err != nil || (response != nil && ValidateGameTurn(*response) != nil) {
			validationErr := err
			if validationErr == nil {
				validationErr = ValidateGameTurn(*response)
			}
			repaired, repairErr := repairGameTurn(ctx, client, pack, session, terminalResults, content, validationErr)
			if repairErr == nil {
				return appendAITurn(session, *repaired), nil
			}
			return appendAITurn(session, FallbackTurn()), nil
		}

		if terminalRequest != nil {
			if terminalRound == maxTerminalRounds {
				return appendAITurn(session, FallbackTurn()), nil
			}
			terminalResults = append(terminalResults, ExecuteTerminalRequest(session, *terminalRequest)...)
			continue
		}

		if response == nil {
			return appendAITurn(session, FallbackTurn()), nil
		}
		return appendAITurn(session, *response), nil
	}

	return appendAITurn(session, FallbackTurn()), nil
}

func repairGameTurn(ctx context.Context, client ChatCompleter, pack StoryPack, session *GameSession, terminalResults []TerminalExecution, raw string, validationErr error) (*AITurnResponse, error) {
	repairInstruction := fmt.Sprintf("上一次模型响应不合规：%s\n原始响应：\n%s\n请只返回一个修正后的 game_turn JSON，不要返回 agent_terminal。", validationErr.Error(), raw)
	content, err := client.Chat(ctx, BuildMessages(pack, session, terminalResults, repairInstruction))
	if err != nil {
		return nil, err
	}
	response, terminalRequest, err := ParseAIResponse(content)
	if err != nil {
		return nil, err
	}
	if terminalRequest != nil {
		return nil, errors.New("repair response returned agent_terminal")
	}
	if response == nil {
		return nil, errors.New("repair response is empty")
	}
	if err := ValidateGameTurn(*response); err != nil {
		return nil, err
	}
	return response, nil
}

func BuildMessages(pack StoryPack, session *GameSession, terminalResults []TerminalExecution, repairInstruction string) []ChatMessage {
	memory, err := ReadWorkspaceFile(session, "memory.md")
	if err != nil {
		memory = pack.Files["memory.md"]
	}

	var builder strings.Builder
	builder.WriteString("你是规则怪谈游戏主持人。必须遵守剧情包规则，只能输出严格 JSON，不允许 Markdown 包裹或额外解释。\n")
	builder.WriteString("不要直接泄露 true.md 的隐藏真相；前端只会展示 game_turn.payload。\n")
	builder.WriteString("用户只能通过 choice 工具行动。continue 状态必须包含一个 choice 工具，选项 2 到 4 个。\n")
	builder.WriteString("如果需要读取剧情文档或更新 memory.md，可以返回 agent_terminal；terminal 工作目录已固定为当前会话 workspace，请使用相对路径。\n")
	builder.WriteString("agent_terminal 不会展示给用户。不要依赖或输出本机绝对路径。\n\n")
	builder.WriteString("输出协议：\n")
	builder.WriteString(`game_turn: {"type":"game_turn","state":"continue","payload":["..."],"tools":[{"type":"choice","id":"main","prompt":"你要怎么做？","options":[{"id":"...","label":"..."}]}]}` + "\n")
	builder.WriteString(`ended: {"type":"game_turn","state":"ended","payload":["..."],"tools":[],"ending":{"id":"...","title":"...","kind":"good|bad|loop|neutral"}}` + "\n")
	builder.WriteString(`agent_terminal: {"type":"agent_terminal","reason":"...","commands":[{"command":"..."}]}` + "\n\n")
	builder.WriteString("Story Pack:\n")
	for _, fileName := range []string{"scene.md", "rule.md", "true.md", "endings.md"} {
		builder.WriteString("\n--- " + fileName + " ---\n")
		builder.WriteString(pack.Files[fileName])
		builder.WriteString("\n")
	}
	builder.WriteString("\n--- current memory.md ---\n")
	builder.WriteString(memory)
	builder.WriteString("\n\nRecent Turns:\n")
	for _, turn := range recentTurns(session.Turns, 12) {
		builder.WriteString(fmt.Sprintf("- %s: %s", turn.Role, strings.Join(turn.Payload, " / ")))
		if turn.SelectedChoiceID != "" {
			builder.WriteString(fmt.Sprintf(" (choice: %s - %s)", turn.SelectedChoiceID, turn.SelectedChoiceLabel))
		}
		if turn.Ending != nil {
			builder.WriteString(fmt.Sprintf(" (ending: %s)", turn.Ending.Title))
		}
		builder.WriteString("\n")
	}

	if len(terminalResults) > 0 {
		builder.WriteString("\nTerminal Results:\n")
		for _, result := range terminalResults {
			encoded, _ := json.Marshal(result)
			builder.Write(encoded)
			builder.WriteString("\n")
		}
	}
	if repairInstruction != "" {
		builder.WriteString("\nRepair Instruction:\n")
		builder.WriteString(repairInstruction)
		builder.WriteString("\n")
	}

	return []ChatMessage{
		{Role: "system", Content: "你是一个严格遵循 JSON 输出协议的 AI 规则怪谈主持人。"},
		{Role: "user", Content: builder.String()},
	}
}

func recentTurns(turns []GameTurn, limit int) []GameTurn {
	if len(turns) <= limit {
		return turns
	}
	return turns[len(turns)-limit:]
}

type AITurnResponse struct {
	Type    string       `json:"type"`
	State   string       `json:"state"`
	Payload []string     `json:"payload"`
	Tools   []ChoiceTool `json:"tools"`
	Ending  *Ending      `json:"ending,omitempty"`
}

func ParseAIResponse(content string) (*AITurnResponse, *AgentTerminalRequest, error) {
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(content), &envelope); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	switch envelope.Type {
	case "game_turn":
		var response AITurnResponse
		if err := json.Unmarshal([]byte(content), &response); err != nil {
			return nil, nil, err
		}
		return &response, nil, nil
	case "agent_terminal":
		var request AgentTerminalRequest
		if err := json.Unmarshal([]byte(content), &request); err != nil {
			return nil, nil, err
		}
		if err := ValidateTerminalRequest(request); err != nil {
			return nil, nil, err
		}
		return nil, &request, nil
	default:
		return nil, nil, fmt.Errorf("unsupported response type: %q", envelope.Type)
	}
}

func ValidateGameTurn(response AITurnResponse) error {
	if response.Type != "game_turn" {
		return errors.New("response type must be game_turn")
	}
	if response.State != "continue" && response.State != "ended" {
		return errors.New("state must be continue or ended")
	}
	if len(response.Payload) == 0 {
		return errors.New("payload is required")
	}
	for _, line := range response.Payload {
		if strings.TrimSpace(line) == "" {
			return errors.New("payload cannot contain empty text")
		}
		if len([]rune(line)) > 2000 {
			return errors.New("payload text is too long")
		}
	}
	if response.State == "continue" {
		choiceTools := 0
		for _, tool := range response.Tools {
			if tool.Type != "choice" {
				return errors.New("only choice tools are supported")
			}
			choiceTools++
			if len(tool.Options) < 1 || len(tool.Options) > 4 {
				return errors.New("choice options must contain 1 to 4 items")
			}
			seen := map[string]bool{}
			for _, option := range tool.Options {
				if strings.TrimSpace(option.ID) == "" || strings.TrimSpace(option.Label) == "" {
					return errors.New("choice options require id and label")
				}
				if seen[option.ID] {
					return errors.New("choice option ids must be unique")
				}
				seen[option.ID] = true
			}
		}
		if choiceTools != 1 {
			return errors.New("continue state requires exactly one choice tool")
		}
	}
	if response.State == "ended" && response.Ending == nil {
		return errors.New("ended state requires ending")
	}
	return nil
}

func ValidateTerminalRequest(request AgentTerminalRequest) error {
	if request.Type != "agent_terminal" {
		return errors.New("terminal request type must be agent_terminal")
	}
	if len(request.Commands) == 0 {
		return errors.New("terminal request requires at least one command")
	}
	if len(request.Commands) > 3 {
		return errors.New("terminal request allows at most three commands")
	}
	for _, command := range request.Commands {
		if strings.TrimSpace(command.Command) == "" {
			return errors.New("terminal command cannot be empty")
		}
	}
	return nil
}

func FallbackTurn() AITurnResponse {
	return AITurnResponse{
		Type:    "game_turn",
		State:   "continue",
		Payload: []string{"手表屏幕闪烁了一下，刚才的记忆像被什么东西擦乱了。"},
		Tools: []ChoiceTool{
			{
				Type:   "choice",
				ID:     "fallback",
				Prompt: "接下来怎么做？",
				Options: []ChoiceOption{
					{ID: "retry", Label: "重新整理思绪"},
				},
			},
		},
	}
}

func appendAITurn(session *GameSession, response AITurnResponse) GameTurnResult {
	tools := response.Tools
	if tools == nil {
		tools = []ChoiceTool{}
	}
	turn := GameTurn{
		ID:        NewID("turn"),
		Role:      TurnRoleAI,
		Payload:   response.Payload,
		Tools:     tools,
		Ending:    response.Ending,
		CreatedAt: NowISO(),
	}
	session.AppendTurn(turn)
	if response.State == "ended" {
		session.State = SessionStateEnded
	}

	return GameTurnResult{
		SessionID: session.ID,
		GameID:    session.GameID,
		State:     session.State,
		Payload:   response.Payload,
		Tools:     tools,
		Ending:    response.Ending,
		Turn:      turn,
	}
}

func StoryFileContents(session *GameSession, fileName string) string {
	content, err := os.ReadFile(filepath.Join(session.WorkspacePath, fileName))
	if err != nil {
		return ""
	}
	return string(content)
}
