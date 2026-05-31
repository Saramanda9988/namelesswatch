# NamelessWatch

NamelessWatch 是一个本地运行的 AI 文字冒险游戏。它使用 Wails 将 Go 后端和 React 前端打包成桌面应用，让 AI 读取剧情包中的场景、规则、隐藏真相、记忆和结局设定，并以规则怪谈主持人的方式逐回合推进游戏。

项目当前围绕“剧情包驱动的 AI 角色扮演”展开：剧情作者不需要硬编码完整分支树，而是用一组规范文档约束世界观、规则、素材和结局，AI 在每一回合输出结构化 JSON，前端据此展示逐句叙事、行动选项、场景切换、BGM 和结局。

## 功能

- 剧情包导入：从本地文件夹导入 `metadata.json`、`scene.md`、`rule.md`、`true.md`、`memory.md`、`endings.md` 等剧情文件。
- AI 主持游玩：兼容 OpenAI Chat Completions API，支持自定义 Base URL、模型和 Token。
- 逐句叙事：AI 输出的 `payload` 按句子数组展示，前端按阅读节奏逐句推进。
- 行动选择：每回合提供 2-4 个行动选项，也支持玩家输入自定义行动。
- 场景与素材：支持场景图、地图、场景坐标和游戏结束图。
- BGM 控制：剧情包可声明 BGM 资源与场景默认曲目，AI 可在回合中请求播放或停止。
- 存档与快照：会话自动保存，可手动保存快照，并从快照分叉继续游玩。
- 成就系统：支持 AI 触发成就和规则判定成就。
- 剧情包 CLI：提供初始化和校验剧情包的命令行工具。

## 技术栈

- 桌面框架：Wails v2
- 后端：Go 1.23
- 前端：React 19、TypeScript、Vite 8、Tailwind CSS v4、shadcn/radix、zustand、TanStack Router/Query
- 包管理：pnpm 10.32.0
- 持久化：本机 JSON 文件，无数据库

## 目录结构

```text
.
├── app.go                 # Wails 暴露给前端的应用方法
├── main.go                # Wails 启动入口
├── cli/                   # 剧情包脚手架与校验 CLI
├── internal/
│   ├── appconf/           # 本机配置读写
│   ├── roleplay/          # AI 回合、提示词、响应校验、共享类型
│   ├── service/           # 游戏库、会话、存档、成就等业务服务
│   └── storypack/         # 剧情包脚手架与校验逻辑
├── frontend/
│   ├── src/               # React 前端源码
│   └── wailsjs/           # Wails 自动生成绑定，请勿手改
└── docs/
    └── example/           # 示例剧情包
```

## 环境要求

- Go 1.23+
- Wails CLI v2
- Node.js 与 pnpm 10.32.0

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

安装前端依赖：

```bash
cd frontend
pnpm install
```

## 本地开发

从仓库根目录启动桌面开发模式：

```bash
wails dev
```

Wails 会启动 Go 后端和 Vite 前端，并在后端绑定变化时重新生成前端调用代码。改动 `app.go` 中暴露给前端的方法后，可以手动重新生成绑定：

```bash
wails generate module
```

## AI 配置

应用首次启动会在本机配置目录创建 `appconf.json`，也可以在应用内“设置”窗口填写：

- Provider
- Base URL
- 模型名称
- Token
- 上下文预算
- 选项预生成配置

默认配置会读取以下环境变量：

```bash
OPENAI_BASE_URL
OPENAI_MODEL
OPENAI_API_KEY
AI_PROVIDER
```

接口需兼容 OpenAI Chat Completions，后端请求路径为：

```text
{AI_BASE_URL}/chat/completions
```

本机数据默认保存在：

```text
{os.UserConfigDir()}/namelesswatch/
```

常见文件包括：

- `appconf.json`：AI 与上下文配置
- `library.json`：已导入的游戏库
- `sessions/`：游玩会话与快照
- `story-assets/`：导入后物化的本地素材
- `achievements.json`：成就解锁记录

## 剧情包格式

一个可导入的剧情包至少需要包含：

```text
metadata.json
scene.md
rule.md
true.md
memory.md
endings.md
```

推荐同时提供：

```text
briefing.json
achievements.json
photo/metadata.json
photo/*.png
bgm/metadata.json
bgm/*.mp3
```

`docs/example/` 提供了可参考的完整示例。

### metadata.json

```json
{
  "title": "未命名规则怪谈",
  "initialScene": "entrance",
  "scenePositions": {
    "entrance": [85, 80],
    "kitchen": [80, 25]
  }
}
```

### photo/metadata.json

将场景 ID 映射到 `photo/` 目录下的图片文件：

```json
{
  "entrance": "玄关背景.png",
  "kitchen": "厨房背景.png",
  "game_over": "game_over.png"
}
```

### bgm/metadata.json

```json
{
  "tracks": {
    "daily": {
      "name": "日常氛围",
      "file": "daily.mp3"
    },
    "tension": {
      "name": "紧张氛围",
      "file": "tension.mp3"
    }
  },
  "sceneDefaults": {
    "entrance": "daily",
    "kitchen": "tension"
  }
}
```

支持的图片格式：`.png`、`.jpg`、`.jpeg`、`.webp`、`.gif`、`.bmp`

支持的音频格式：`.mp3`、`.ogg`、`.wav`、`.m4a`、`.webm`

## 剧情包 CLI

初始化剧情包：

```bash
go run ./cli init --title "我的规则怪谈" --initial-scene entrance ./packs/my-story
```

校验剧情包：

```bash
go run ./cli validate ./packs/my-story
```

覆盖已有脚手架文件：

```bash
go run ./cli init --force ./packs/my-story
```

`init` 会生成剧情文档、玩家开局须知、成就配置，以及 `photo/`、`bgm/` 的素材元数据文件。

## 常用命令

后端构建与测试：

```bash
go build ./...
go vet ./...
go test ./...
```

前端类型检查与构建：

```bash
cd frontend
pnpm exec tsc --noEmit
pnpm build
```

生产构建：

```bash
wails build
```