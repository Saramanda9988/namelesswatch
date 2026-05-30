# AGENTS.md — 编码约束

namelesswatch 是一个 Wails（Go 后端 + React 前端）的 AI 文字冒险游戏。本文件约束所有改动需遵守的工程规范。

## 技术栈

- 后端：Go 1.23 + Wails v2，代码在根目录 `app.go`、`main.go` 与 `internal/`。
- 前端：React 19 + TypeScript + Vite 8 + Tailwind v4 + shadcn(radix base) + zustand + TanStack Router/Query，目录 `frontend/`。
- 包管理：前端用 **pnpm**（见 `frontend/package.json` 的 `packageManager`，固定版本）；不要混用 npm/yarn。
- 持久化：纯 JSON 文件，落盘到 `{os.UserConfigDir()}/namelesswatch/`（如 `appconf.json`、`library.json`、`sessions/`）。无数据库。

## 常用命令

- 开发运行：`wails dev`（根目录；会自动重新生成前端绑定并热重载）。
- 重新生成 Wails TS 绑定：`wails generate module`（改动 `app.go` 暴露方法后必须执行）。
- 后端构建/测试：`go build ./...`、`go vet ./...`、`go test ./...`（根目录）。
- 前端类型检查/构建：在 `frontend/` 下 `pnpm exec tsc --noEmit`、`pnpm build`。
- 添加 shadcn 组件：在 `frontend/` 下用 `pnpm exec shadcn add <name>`（**不要**用 `pnpm dlx shadcn@latest`，当前有 zod/MCP 依赖冲突会报错）。

## 提交前必须通过

- 后端：`go build ./...` + `go test ./...`。
- 前端：`pnpm exec tsc --noEmit`（必须零错误）。
- 用 ReadLints 检查改动文件，不要引入新的 lint 报错；不强求修复未改动行的既有警告（如 Tailwind `bg-gradient-to-*` 提示）。

## 后端 Go 约束

- 包组织：业务逻辑放在 `internal/` 下的 `appconf`/`service`/`roleplay`，不要把逻辑写进 `app.go`；`app.go` 只做 Wails 绑定的薄封装。
- 并发：`GameService` 等共享状态的 struct 用 `sync.Mutex` 保护 map；持有锁期间不要做 IO/网络等阻塞调用（参考 `advanceSession` 先解锁再调 AI 再加锁写回的模式）。
- 错误处理：用 `fmt.Errorf("...: %w", err)` 包装并保留错误链；对外返回的错误信息保持简洁。
- 持久化：文件写入一律「写临时文件 + `os.Rename` 原子替换」（参考 `gameRepository.save` / `sessionRepository.save`）；持久化失败不应阻断核心流程（如自动存档失败只记日志）。
- 时间与 ID：统一用 `roleplay.NowISO()` 生成时间戳、`roleplay.NewID(prefix)` 生成 ID，不要自造格式。
- 日志：用 Wails runtime 日志（`s.logInfof/logErrorf/logWarningf`，内部封装 `wailsruntime.LogInfof` 等），不要用 `fmt.Println`。
- 测试：新增 repository/service 能力时补 `_test.go`，覆盖 save/load/list/delete 等往返；用 `t.TempDir()` 隔离磁盘。
- 数据结构跨语言：放在 `roleplay/types.go` 的 struct 是前后端共享契约，JSON tag 用 camelCase，新增可选字段加 `,omitempty` 并保证零值兼容。

## 前端约束

- 代码风格：2 空格缩进、单引号、**不写分号**（与现有文件保持一致）；函数组件 + Hooks。
- 模块导入：用 `@/` 别名（`@/components`、`@/lib`、`@/hooks`、`@/stores`）；UI 基础组件来自 `@/components/ui`（shadcn）。
- 后端调用：统一通过 `frontend/src/stores/game-store.ts`（zustand）封装的 action 调用，不要在页面里直接散落 import `wailsjs/go/main/App` 的方法（已存在的除外）。
- 状态管理：全局/跨页状态进 zustand store；页面局部状态用 `React.useState`。zustand 取值用细粒度 selector（`useGameStore((s) => s.xxx)`）。
- 样式：用 Tailwind + `cn()`（`@/lib/utils`）合并类名；遵循 Tailwind v4 写法。
- UI 文案：使用简体中文。
- 交互元素加可访问性属性：图标按钮要有 `aria-label`，悬浮说明用 `title`。

## 禁止事项

- 不要手改自动生成文件：`frontend/wailsjs/**`（绑定与 models）由 Wails 生成，改后端后用 `wails generate module` 重新生成。
- 不要提交：`frontend/dist`、`node_modules`、`go.sum`、`appconf.json`、`SaveData`、各类本地日志（见 `.gitignore`）。
- 不要把密钥/Token 写进代码或提交（AI token 存于本机 `appconf.json`）。
- 不要在持有锁时执行长耗时调用；不要在前端绕过 store 直接改后端状态文件。
- 不要新增重型依赖来替代既有纯文件持久化方案，除非另行讨论。

## 参考文档

- 需求与设计：`docs/prd.md`、`docs/backend.md`。
- 剧情包示例：`docs/example/`（`scene.md`/`rule.md`/`true.md`/`memory.md`/`endings.md` + `metadata.json`）。
