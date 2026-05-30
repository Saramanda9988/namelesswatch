# AI Roleplay MVP Scope

## 背景

本项目是一款 AI 规则怪谈角色扮演应用。用户导入一个剧情资源包后，AI 作为主持人读取资源包中的剧情资料、规则、真相、结局和记事本，并根据用户选择持续推进游戏，直到进入某个结局。

第一版目标不是实现完整规则引擎，而是建立一个可玩的 AI 主导闭环：

1. 导入标准剧情包。
2. 启动 AI 游戏会话。
3. AI 输出结构化 JSON。
4. 前端逐行展示文本。
5. 用户通过选择框做决定。
6. AI 根据选择继续推进，并可以通过 terminal 读写当前会话 workspace 中的文档。
7. AI 判断并返回游戏结局。

## 第一版业务范围

### 支持内容

- 使用 OpenAI-compatible 接口调用模型。
- 只支持选择框交互，不支持用户自由输入。
- 第一版由 AI 主导剧情判断、规则执行和结局判定。
- 后端负责组装 prompt、调用模型、校验 JSON、维护会话状态。
- 前端负责展示 AI 文本、渲染选择项、发送用户选择、展示结局。
- `memory.md` 是 agent 的长期记事本，属于剧情包内容的一部分。
- agent 可以通过 terminal 读取剧情包文档、读取 `memory.md`、修改 `memory.md`。
- terminal 能力只用于 MVP 试水和 agent 内部推理，不直接暴露给用户。

### 暂不支持内容

- 不做确定性规则引擎。
- 不支持用户自由文本输入。
- 不支持多人游戏。
- 不支持复杂存档管理。
- 不支持流式模型输出。
- 不支持自动生成图片、地图、角色立绘。
- 不强制实现 BGM、音效、背景切换等表现层工具。

## 剧情包标准

标准剧情包应包含以下 Markdown 文件，文件名不带 `@`：

```text
scene.md
rule.md
true.md
memory.md
endings.md
```

### scene.md

描述故事开头、初始场景、主角身份、当前时间、初始事件等。

用途：

- 作为游戏开局上下文。
- 帮助 AI 建立初始叙事语气和场景。

### rule.md

描述游戏规则，包括用户显式知道的规则、隐藏规则、阶段规则、死亡条件、循环条件等。

用途：

- 作为 AI 推进剧情时必须遵守的约束。
- 第一版不由后端解析为确定性规则，而是直接放入 prompt。

### true.md

描述故事真相。

用途：

- 只提供给 AI。
- 不应直接暴露给用户。
- AI 可以基于真相进行暗示、伏笔和结局判断。

### memory.md

agent 的记事本。

用途：

- 记录用户已经做过的选择。
- 记录当前阶段、重要道具、异常事件、可能结局走向。
- AI 可以通过 terminal 读取和修改当前会话的 `memory.md`。
- 第一版中，`memory.md` 是当前游戏会话的可变文本，不直接修改原始导入包。

启动游戏时，后端应将原始剧情包复制到当前 session workspace：

```text
session-workspace/
  scene.md       只读语义
  rule.md        只读语义
  true.md        只读语义，隐藏给用户
  endings.md     只读语义
  memory.md      可读写，会话副本
```

MVP 阶段不设计复杂 `MemoryAction` DSL。agent 如果需要整理、追加或重写记忆，可以直接通过 terminal 修改 `memory.md`。这会牺牲一部分边界清晰度，但能更快验证 agent 是否真的需要文档级操作能力。

### endings.md

描述所有可达结局。

用途：

- 供 AI 判断何时结束游戏。
- 供 AI 在结束时选择对应结局并生成结局文本。

## 会话模型

一次游戏会话由以下信息组成：

```ts
type GameSession = {
  id: string
  gameId: string
  state: 'idle' | 'playing' | 'ended'
  workspacePath: string
  memoryPath: string
  turns: GameTurn[]
  createdAt: string
  updatedAt: string
}
```

```ts
type GameTurn = {
  id: string
  role: 'ai' | 'user'
  payload: string[]
  selectedChoiceId?: string
  selectedChoiceLabel?: string
  tools?: GameTool[]
  createdAt: string
}
```

第一版可以只保存在前端内存或后端内存中，不要求落盘。后续再扩展为本地持久化存档。

完整 `GameSession` 不应直接进入模型推理上下文。`sessionId`、路径、时间戳、UI 状态等运行时字段只由后端使用。

每轮发给模型的应该是后端构造出的推理上下文：

```text
Reasoning Context
├─ 剧情包文档摘要或全文
│  ├─ scene.md
│  ├─ rule.md
│  ├─ true.md
│  └─ endings.md
├─ 当前 memory.md 内容
├─ 最近若干轮 AI 文本和用户选择
├─ 当前用户选择
└─ 输出协议与工具使用规则
```

历史 turn 不无限追加。第一版建议只保留最近 8 到 12 轮；长期状态依赖 `memory.md` 沉淀。

## AI 输出协议

模型必须返回严格 JSON，不允许返回 Markdown 包裹、不允许返回额外解释。

模型响应分为两类：

1. 面向用户的游戏回合响应。
2. 面向后端的 agent terminal 请求。

后端收到 terminal 请求后执行命令，将结果追加进推理上下文，再继续调用模型，直到得到游戏回合响应或达到最大工具轮数。

### 游戏回合响应

```ts
type AiTurnResponse = {
  type: 'game_turn'
  state: 'continue' | 'ended'
  payload: string[]
  tools: GameTool[]
  ending?: EndingResult
}
```

### payload

`payload` 是展示给用户的文本数组。

要求：

- 每个元素是一行或一段短文本。
- 使用第二人称“你”叙述。
- 不直接泄露 `true.md` 中的真相。
- 不展示系统提示、规则原文或 JSON 解释。

### UI tools

第一版必须至少支持选择框工具。

```ts
type GameTool = ChoiceTool

type ChoiceTool = {
  type: 'choice'
  id: string
  prompt?: string
  options: ChoiceOption[]
}

type ChoiceOption = {
  id: string
  label: string
}
```

约束：

- 当 `state` 为 `continue` 时，必须返回一个 `choice` 工具。
- 每次最多一个 `choice` 工具。
- 每个选择项应是可执行动作，不是解释性文字。
- 建议每次 2 到 4 个选项。

### Agent terminal 请求

```ts
type AgentTerminalRequest = {
  type: 'agent_terminal'
  reason: string
  commands: TerminalCommand[]
}

type TerminalCommand = {
  command: string
}
```

约束：

- terminal 请求不直接展示给用户。
- terminal 的工作目录固定为当前 session workspace。
- agent 可以用 terminal 读取 `scene.md`、`rule.md`、`true.md`、`endings.md`、`memory.md`。
- agent 可以修改 `memory.md`。
- agent 不应修改 `scene.md`、`rule.md`、`true.md`、`endings.md`。
- 后端应设置最大 terminal 轮数，建议每个用户回合最多 3 轮。
- 后端应记录 command、stdout、stderr，便于调试。

MVP 可以先使用 terminal 作为粗粒度能力，不提前设计文件工具 DSL。后续如果发现 terminal 太宽或不可控，再收敛为 `read_doc`、`append_memory`、`update_memory_section` 等受控工具。

### ending

当 `state` 为 `ended` 时，必须返回 `ending`。

```ts
type EndingResult = {
  id: string
  title: string
  kind: 'good' | 'bad' | 'loop' | 'neutral'
}
```

约束：

- `ending.id` 应对应 `endings.md` 中定义的某个结局。
- 结束回合可以不返回选择框。

## 后端能力

后端建议提供以下 Wails 方法：

```go
StartGame(gameID string) (GameTurnResult, error)
SubmitChoice(sessionID string, choiceID string) (GameTurnResult, error)
GetSession(sessionID string) (GameSession, error)
```

### StartGame

职责：

- 读取游戏包文档。
- 初始化 session。
- 创建当前 session workspace。
- 将剧情包文档复制到 session workspace。
- 调用 AI 生成第一回合。
- 返回文本和选择项。

### SubmitChoice

职责：

- 记录用户选择。
- 将剧情文档、当前 `memory.md`、历史回合、用户选择组装为 prompt。
- 调用 OpenAI-compatible 接口。
- 如果 AI 返回 terminal 请求，执行命令并继续推理。
- 如果 AI 返回游戏回合，校验 AI JSON。
- 返回下一回合。

### GetSession

职责：

- 返回当前会话状态。
- 供刷新、调试或后续存档功能使用。

## OpenAI-compatible 配置

第一版需要支持以下配置：

```text
OPENAI_BASE_URL
OPENAI_API_KEY
OPENAI_MODEL
```

建议从环境变量或本地配置文件读取，不在前端暴露 API key。

请求接口使用 OpenAI Chat Completions 兼容格式即可。第一版不依赖原生 function calling，所有工具都由模型写入 JSON 响应，再由后端解释执行。

## Prompt 结构

每次调用模型时，后端应组装以下上下文：

```text
System:
- 你是规则怪谈游戏主持人。
- 必须遵守剧情包规则。
- 只能输出严格 JSON。
- 不允许泄露隐藏真相。
- 当前只允许通过 choice 工具让用户行动。
- 如果需要读取或修改文档，可以返回 agent_terminal 请求。
- terminal 结果不会直接展示给用户。

Story Pack:
- scene.md
- rule.md
- true.md
- endings.md

Current Memory:
- 当前 session workspace 中 memory.md 的内容

Recent Turns:
- 最近若干轮 AI 文本和用户选择

User Action:
- 用户刚刚选择了什么
```

每轮 prompt 不直接包含完整 `GameSession`。后端只把本轮推理真正需要的信息发给模型。

## Agent terminal 执行策略

terminal 是第一版用于验证文档级 agent 能力的 MVP 方案。

基本流程：

```text
用户选择
  ↓
后端构造推理上下文
  ↓
模型返回 agent_terminal
  ↓
后端在 session workspace 执行命令
  ↓
后端把命令结果加入上下文
  ↓
模型返回 game_turn
  ↓
前端展示文本和选择框
```

执行边界：

- 工作目录必须固定为当前 session workspace。
- 不把真实绝对路径写入 prompt，避免模型依赖本机路径。
- 命令超时建议 3 到 5 秒。
- 命令输出需要做长度截断。
- 每个用户回合最多执行 3 轮 terminal 请求。
- terminal 执行失败时，把错误作为工具结果返回给模型，让模型自行修正一次。

MVP 阶段可以接受 terminal 能力偏宽，因为目标是验证 agent 是否能有效维护 `memory.md`。当验证通过后，再决定是否收敛为白名单文件工具。

## 前端交互

游玩页需要从静态脚本播放改为会话驱动：

1. 页面打开后调用 `StartGame(gameId)`。
2. 渲染 `payload` 文本。
3. 如果返回 `choice` 工具，展示选择框。
4. 用户点击选择后禁用按钮，调用 `SubmitChoice(sessionId, choiceId)`。
5. 返回下一回合后追加文本。
6. 如果 `state` 为 `ended`，展示结局状态，不再展示普通选择。

第一版不需要自由输入框。

## 校验与兜底

后端必须校验 AI 返回：

- 是否为合法 JSON。
- `state` 是否为允许值。
- `payload` 是否为非空字符串数组。
- `continue` 状态是否包含选择框。
- `ended` 状态是否包含 ending。
- choice id 是否唯一。
- 文本长度是否在合理范围内。
- terminal 请求是否超过最大轮数。
- terminal 命令输出是否超过长度限制。

如果模型返回非法结果，后端可以进行一次修复重试。重试仍失败时，返回一个固定错误回合，例如：

```json
{
  "state": "continue",
  "payload": ["手表屏幕闪烁了一下，刚才的记忆像被什么东西擦乱了。"],
  "tools": [
    {
      "type": "choice",
      "id": "retry",
      "options": [
        { "id": "continue", "label": "重新整理思绪" }
      ]
    }
  ]
}
```

## MVP 验收标准

- 可以导入包含 `scene.md`、`rule.md`、`true.md`、`memory.md`、`endings.md` 的剧情包。
- 可以从游戏库进入 AI 游玩页。
- 首回合由 AI 根据剧情包生成。
- 用户只能通过选择框推进。
- 每次选择后 AI 能继续返回结构化 JSON。
- AI 可以通过 terminal 读取剧情包文档。
- AI 可以通过 terminal 更新当前 session workspace 中的 `memory.md`。
- 游戏可以进入 `ended` 状态并展示结局。
- OpenAI-compatible API key 不出现在前端代码和浏览器运行时。

## 后续扩展方向

- 将关键规则从 `rule.md` 抽取为结构化规则引擎。
- 支持自由文本输入。
- 支持本地存档和读档。
- 支持 BGM、音效、背景图、角色立绘等工具。
- 支持流式输出和打字机效果。
- 支持多剧情包版本管理。
- 支持 prompt 调试面板。
- 将 MVP terminal 能力收敛为更安全的受控文件工具。
