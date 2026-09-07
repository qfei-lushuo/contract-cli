# Changelog

## Unreleased

- 临时恢复 dev 联调：固定 `contract-dev` profile、显式选择环境、独立授权与 dev-only 网络校验；上线前移除，不改变默认 prod 行为
- 为每个 OpenPlatform 逻辑请求生成 W3C `traceparent`，同时发送同值 `X-Log-Id`；重试复用 `trace_id`、每次 HTTP attempt 使用新 `span_id`，失败信息包含可检索的 `trace_id`
- `contract-cli update` 对齐 lark-cli：固定检查 npm `latest`，支持 `--check`、`--force`、`--json`，并按 npm/pnpm 安装来源执行精确版本自更新与安装后校验
- 自动更新提示改为先读 24 小时本地缓存、过期后后台刷新，不再阻塞业务命令；Windows 增加 `.old` 备份、失败恢复和 npm wrapper 启动自愈
- 新增逐请求调用环境识别：实际业务 HTTP 请求发送前回溯父进程；macOS 校验 Bundle ID + Team ID，Windows 校验 Package Family Name 或 Authenticode 证书指纹，并透传归一化来源 Header
- 新增 `contract-cli environment inspect` 本地诊断命令；识别结果不写入 profile 或 OAuth Token
- 新增 `internal/build`，支持 `contract-cli version` 与 `--version`
- 新增 `build.sh`、`Makefile`、`.goreleaser.yml`
- 新增 npm/npx 薄包装：`package.json`、`scripts/install.js`、`scripts/run.js`
- 新增 `tests/cli_e2e/smoke.sh` 作为发布前冒烟脚本

## 0.1.0

- 初始化 `contract-cli` 构建、发布与分发脚手架
