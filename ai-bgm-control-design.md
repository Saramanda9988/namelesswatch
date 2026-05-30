# AI BGM 控制能力设计文档

## 背景

NamelessWatch 当前已经支持 AI 通过 `game_turn` 返回叙事文本、选择框、结局和场景切换。场景切换由 `scene` 字段承载，前端根据返回结果切换背景图和地图激活点。

现有 PRD 已提到“BGM 控制能力，AI 进行 BGM 控制切换，塞到 tools 中”。结合当前代码实现，BGM 更适合作为和 `scene` 平级的回合状态变更，而不是放入 `tools` 数组。

## 目标

- 剧情包可以声明和携带 BGM 音频资源。
- AI 可以根据场景或氛围变化请求播放、切换或停止 BGM。
- 后端只允许 AI 选择剧情包内声明过的曲目 id。
- 前端负责实际播放行为，包括循环、音量、静音、淡入淡出和页面退出停止。
- 存档和恢复后能够还原当前 BGM 状态。

## 非目标

- 不让 AI 直接提供任意音频 URL 或本地路径。
- 不让 AI 控制全局音量、用户静音开关或复杂混音参数。
- 第一版不做音效 SE、多轨混音、环境声分层。
- 第一版不引入额外音频库，优先使用浏览器 `HTMLAudioElement`。

## 现有约束

### `tools` 当前是 choice 专用

`internal/roleplay/ai_session.go` 中的 `ValidateGameTurn` 当前要求：

- `continue` 状态必须有且只有一个 `choice` 工具。
- `tools` 中出现非 `choice` 类型会被拒绝。
- 前端 `choiceToolFrom` 也只从 `tools` 中查找 `type === 'choice'`。

因此如果直接把 BGM 作为 `tools` 中的第二个工具，会破坏现有校验语义，并且需要把 `ChoiceTool` 改成多态工具模型，改动面更大。

### `scene` 已经提供了状态变更范式

当前 `scene` 字段是 `game_turn` 的平级字段：

```json
{
  "type": "game_turn",
  "state": "continue",
  "payload": ["..."],
  "scene": { "id": "kitchen", "reason": "用户进入厨房" },
  "tools": [{ "type": "choice", "id": "main", "options": [] }]
}
```

BGM 的性质和 `scene` 类似，都是“本回合导致的表现层状态变化”，不是用户交互控件。因此推荐沿用这个模式。

## 推荐方案

新增 `bgm` 作为 `game_turn` 平级字段。

```json
{
  "type": "game_turn",
  "state": "continue",
  "payload": [
    "你推开厨房门，冰箱的嗡鸣突然变得很低。"
  ],
  "scene": {
    "id": "kitchen",
    "reason": "用户进入厨房"
  },
  "bgm": {
    "action": "play",
    "id": "kitchen_tension",
    "reason": "氛围转为压迫"
  },
  "tools": [
    {
      "type": "choice",
      "id": "main",
      "prompt": "你要怎么做？",
      "options": [
        { "id": "open_fridge", "label": "打开冰箱" },
        { "id": "step_back", "label": "后退一步" }
      ]
    }
  ]
}
```

### BGM 语义

- 没有 `bgm` 字段：保持当前 BGM，不重启、不切歌。
- `{"action":"play","id":"track_id"}`：播放或切换到指定曲目。
- `{"action":"stop"}`：淡出并停止当前 BGM。
- `reason` 仅用于日志、调试和未来回放，不展示给用户。

### 播放策略

BGM 默认循环播放。AI 只负责切换意图，前端负责播放细节。

```text
AI 返回 bgm play A
        |
        v
前端播放 A，loop = true
        |
下一回合没有 bgm 字段
        |
        v
继续播放 A，不重启
        |
AI 返回 bgm play B
        |
        v
A 淡出，B 淡入并循环
```

## 剧情包格式

新增可选目录 `bgm/`。

推荐结构：

```text
story-pack/
  metadata.json
  scene.md
  rule.md
  true.md
  memory.md
  endings.md
  photo/
    metadata.json
    ...
  bgm/
    metadata.json
    home_ambient.mp3
    kitchen_tension.mp3
```

推荐 `bgm/metadata.json`：

```json
{
  "tracks": {
    "home_ambient": {
      "name": "家中低频",
      "file": "home_ambient.mp3"
    },
    "kitchen_tension": {
      "name": "厨房压迫",
      "file": "kitchen_tension.mp3"
    }
  },
  "sceneDefaults": {
    "entrance": "home_ambient",
    "kitchen": "kitchen_tension"
  }
}
```

### 字段说明

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `tracks` | 是 | 曲目表，key 是 AI 可引用的曲目 id |
| `tracks.*.name` | 否 | 给 prompt、调试日志和未来 UI 使用的人类可读名称 |
| `tracks.*.file` | 是 | `bgm/` 目录下的音频文件名 |
| `sceneDefaults` | 否 | 场景默认 BGM，可作为 AI 或后端兜底参考 |

第一版建议支持的音频扩展名：

- `.mp3`
- `.ogg`
- `.wav`
- `.m4a`
- `.webm`

## 后端类型设计

建议新增类型：

```go
type BGMAsset struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	FileName string `json:"fileName"`
	URL      string `json:"url"`
}

type BGMChange struct {
	Action string `json:"action"`
	ID     string `json:"id,omitempty"`
	Reason string `json:"reason,omitempty"`
}
```

建议修改共享结构：

```go
type StoryPack struct {
	ID      string            `json:"id"`
	Files   map[string]string `json:"files"`
	Scenes  []SceneAsset      `json:"scenes,omitempty"`
	BGMs    []BGMAsset        `json:"bgms,omitempty"`
	MapURLs []string          `json:"mapUrls,omitempty"`
}

type LibraryGame struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Files     map[string]string `json:"files"`
	PhotoURLs []string          `json:"photoUrls"`
	MapURLs   []string          `json:"mapUrls"`
	Scenes    []SceneAsset      `json:"scenes,omitempty"`
	BGMs      []BGMAsset        `json:"bgms,omitempty"`
}

type GameSession struct {
	CurrentSceneID string `json:"currentSceneId,omitempty"`
	CurrentBGMID   string `json:"currentBgmId,omitempty"`
}

type GameTurn struct {
	Scene *SceneChange `json:"scene,omitempty"`
	BGM   *BGMChange   `json:"bgm,omitempty"`
}

type GameTurnResult struct {
	Scene        *SceneChange `json:"scene,omitempty"`
	BGM          *BGMChange   `json:"bgm,omitempty"`
	CurrentBGMID string      `json:"currentBgmId,omitempty"`
}

type AITurnResponse struct {
	Scene *SceneChange `json:"scene,omitempty"`
	BGM   *BGMChange   `json:"bgm,omitempty"`
}
```

说明：

- `BGMChange` 是本回合事件。
- `CurrentBGMID` 是会话当前状态。
- 恢复存档时，前端应使用 `currentBgmId` 直接恢复当前曲目。
- `bgm` 为空时，不代表停止，只代表本回合没有 BGM 变更。

## 后端解析和持久化

### 资源解析

在 `internal/roleplay/types.go` 中新增 `parseBGMAssets`：

- 读取 `bgm/metadata.json`。
- 解析 `tracks`。
- 校验 id、file 非空。
- 根据 `bgm/<file>` 找到对应资源 URL。
- 忽略缺失文件或空 URL 的条目。

### 资源落盘

`internal/service/game_assets.go` 当前只处理 `photo/` 和 `map/` 图片资源。需要扩展：

- 将 `imageAssetBytes` 泛化为 `assetBytes`。
- 新增 `isPackAudioPath`。
- 对 `bgm/` 下音频文件执行和图片相同的 materialize 流程。
- 生成 `/local/story-assets/<gameID>/bgm/<file>` URL。

### 存档兼容

新增字段均使用 `omitempty`，旧存档没有 `currentBgmId` 时等价于无 BGM。无需迁移。

## AI Prompt 设计

`BuildMessages` 中新增：

```text
Available BGM:
- home_ambient => 家中低频
- kitchen_tension => 厨房压迫

Current BGM:
- home_ambient
```

`BuildSystemPrompt` 新增约束：

```text
如果场景或氛围明显变化，可以在 game_turn 中返回 bgm 字段。
bgm 字段只能是 {"action":"play","id":"...","reason":"..."} 或 {"action":"stop","reason":"..."}。
play 的 id 必须来自 Available BGM。
如果当前 BGM 已经合适，不要返回 bgm 字段，前端会继续循环播放当前曲目。
不要把 BGM 放入 tools；tools 只用于用户 choice。
不要在 payload 中说明“音乐切换了”，除非这属于用户能在剧情世界中听见的声音。
```

响应示例也应更新：

```text
game_turn:
{"type":"game_turn","state":"continue","payload":["..."],"scene":{"id":"kitchen","reason":"..."},"bgm":{"action":"play","id":"kitchen_tension","reason":"..."},"tools":[{"type":"choice","id":"main","prompt":"你要怎么做？","options":[{"id":"...","label":"..."}]}]}
```

## 校验规则

新增 `ValidateBGMChange`，由 `ValidateGameTurnForSession` 调用。

规则：

- `bgm == nil`：合法。
- `action == "play"`：必须提供非空 `id`，且 id 必须存在于 `pack.BGMs`。
- `action == "stop"`：合法，忽略 `id` 或要求 `id` 为空均可。建议第一版忽略 `id`，降低模型修复频率。
- 其他 action：非法。
- 没有可用 BGM 时返回 `play`：非法，触发现有修复重试。

append AI turn 时：

- 如果 `bgm.action == "play"`，更新 `session.CurrentBGMID = bgm.ID`。
- 如果 `bgm.action == "stop"`，清空 `session.CurrentBGMID`。
- 如果 `bgm == nil`，不改变 `session.CurrentBGMID`。

## 前端播放设计

### 状态来源

`PlayPage` 每次收到 `GameTurnResult`：

- 优先读取 `result.currentBgmId` 作为真实播放状态。
- 根据 `game.bgms` 找到曲目 URL。
- 曲目不存在时停止播放并记录日志。

### 播放 hook

建议新增 `frontend/src/hooks/use-bgm-player.ts`。

职责：

- 管理 `HTMLAudioElement`。
- 设置 `loop = true`。
- 根据 `enabled` 和 `volume` 控制播放。
- 曲目变化时执行淡出、换源、淡入。
- 页面卸载时 pause 并清空 src。
- 捕获 `audio.play()` rejection，交给 UI 显示“点击启用音乐”状态。

### UI 控制

`PlayPage` 左侧工具栏增加图标按钮：

- 静音/取消静音。
- 当前 BGM 被浏览器或 WebView 拦截时，点击按钮触发 `play()`。

设置页建议将现有“语音音量”改为更通用的“音量”，或者新增：

- `bgmEnabled: boolean`
- `bgmVolume: number`

第一版可以复用现有 `voiceVolume`，但命名会不准确。

## 流程图

```text
导入剧情包
    |
    v
读取 bgm/metadata.json
    |
    v
音频资源 materialize 到 /local/story-assets
    |
    v
StoryPack / LibraryGame 暴露 BGMs
    |
    v
BuildMessages 暴露 Available BGM / Current BGM
    |
    v
AI 返回 game_turn.bgm
    |
    v
后端校验 action/id
    |
    v
更新 GameSession.CurrentBGMID
    |
    v
返回 GameTurnResult.currentBgmId
    |
    v
前端循环播放、切歌或停止
```

## 场景默认 BGM 的使用方式

`sceneDefaults` 有两种可选接入方式。

### 方式 A：只给 AI 参考

后端在 prompt 中展示每个场景默认 BGM，让 AI 自己在需要时返回 `bgm`。

优点：

- 行为完全由 AI 决定。
- 不会出现后端静默切歌。

缺点：

- 模型可能忘记切歌。

### 方式 B：后端兜底

如果 AI 返回了 `scene` 变更，但没有返回 `bgm`，后端根据 `sceneDefaults` 自动切换。

优点：

- “场景变化调整 BGM”更稳定。
- 对较弱模型更友好。

缺点：

- AI 的叙事意图可能和默认曲目不完全一致。
- 需要在结果中标记这次 BGM 是自动兜底还是 AI 指令，便于调试。

推荐第一版采用方式 A，避免隐式行为。等实际体验发现模型经常漏切，再加方式 B。

## 实施步骤

1. 定义后端 BGM 类型和 JSON 字段。
2. 解析 `bgm/metadata.json` 并更新 `StoryPack` / `LibraryGame`。
3. 扩展资源导入和 materialize，支持音频文件。
4. 扩展 AI prompt，暴露 Available BGM 和 Current BGM。
5. 扩展 AI 响应结构、校验、修复重试和 `appendAITurn` 状态更新。
6. 更新 session 保存和恢复返回，确保 `currentBgmId` 可恢复。
7. 重新生成 Wails TS 绑定。
8. 前端导入流程允许音频文件。
9. 前端 `PlayPage` 增加 BGM 播放 hook 和控制按钮。
10. 补充 Go 单元测试和前端类型检查。

## 测试建议

后端：

- `NewLibraryGame` 能解析 BGM metadata。
- 音频资源可 materialize 为 `/local/story-assets/...`。
- 合法 `bgm.play` 通过校验。
- 未声明 id 的 `bgm.play` 被拒绝并触发修复。
- `bgm.stop` 会清空 `CurrentBGMID`。
- 无 `bgm` 字段时保持原 `CurrentBGMID`。
- session save/load 保留 `CurrentBGMID`。

前端：

- `pnpm exec tsc --noEmit` 通过。
- 无 BGM 的旧剧情包仍可正常游玩。
- 进入页面后首个 BGM 可播放或显示手动启用状态。
- 同一曲目连续回合不重启。
- 切换曲目时不会叠放多个音频。
- 离开游玩页后音乐停止。

## 风险和注意事项

- WebView 或浏览器可能拒绝非用户手势触发的自动播放，需要 UI 提供手动启用入口。
- 音频格式跨平台支持不完全一致，推荐剧情包优先使用 `.mp3`。
- 如果 AI 每回合都返回同一首 `play`，前端应识别同 id 并不重启，同时 prompt 要要求“无变化时省略 bgm”。
- 如果未来要支持音效，建议新增 `sfx` 字段，不要复用 BGM 字段。
- 如果未来 `tools` 需要支持多态，应单独重构 `ChoiceTool`，不要为了 BGM 提前扩大 `tools` 语义。

## 结论

第一版建议采用：

- 剧情包声明受控 BGM 曲库。
- AI 在 `game_turn` 平级 `bgm` 字段中返回播放、切换或停止意图。
- 后端校验曲目 id 并维护 `CurrentBGMID`。
- 前端按当前 BGM 状态循环播放，并负责切歌淡入淡出、音量和用户开关。

这个方案贴合现有 `scene` 状态变更模式，避免破坏 `tools` 的 choice 专用语义，改动面可控，并且便于后续扩展场景默认曲目和音效系统。
