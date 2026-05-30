# README

## About

This is the official Wails React-TS template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## 剧情包 CLI

从仓库根目录运行：

- 初始化剧情包：`go run ./cli init --title "我的规则怪谈" ./packs/my-story`
- 校验剧情包：`go run ./cli validate ./packs/my-story`

`init` 会生成 `metadata.json`、`briefing.json`、`scene.md`、`rule.md`、`true.md`、`memory.md`、`endings.md`，以及 `photo/`、`bgm/` 下的素材元数据文件。已有脚手架文件默认不会被覆盖；需要覆盖时加 `--force`。

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.
