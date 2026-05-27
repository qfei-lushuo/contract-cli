# AI 变更记录：升级提示 JSON notice 与 24 小时缓存

## 2026-05-27 飞书式升级提示改造

变更摘要：将自动升级提示从交互终端 stderr 文本提示改为 JSON 输出中的 `_notice.update`，并将自动远端检查改为 24 小时缓存节流。

关键逻辑/决策：

- 普通命令不再输出旧的 `A new contract-cli version is available` stderr 文本。
- JSON object 输出会在发现新版本时注入 `_notice.update`，其中 `command` 继续使用 npm 安装命令。
- `update-check.json` 作为 24 小时缓存：fresh cache 不再请求 npm registry，但仍可用于注入 notice。
- `contract-cli update check` 调整为飞书式手动校验：默认输出文本，带 `--json` 时返回顶层 `ok/action/current_version/latest_version/message` 等结构；手动检查不注入 `_notice.update`。
- `--raw`、yaml、table、纯文本命令不注入 `_notice.update`；CI 环境跳过自动远端检查。
- `contract-cli-shared` skill 的触发描述补充 “看到 JSON 输出中的 `_notice` / `_notice.update`”，对齐飞书 `lark-shared` 的 agent 触发方式。
- `contract-cli-shared` skill 补充普通查询不要默认加 `--raw`；说明 `--raw` 会绕过 JSON renderer，不会注入 `_notice.update`。

测试覆盖：

- 新增 update cache freshness、cache notice 生成测试。
- 新增 renderer 注入、合并和跳过 notice 的测试。
- 更新 CLI 自动检查测试，覆盖 JSON notice、fresh cache、禁用环境变量、手动 `update check` 默认文本输出和 `--json` 飞书式结构输出。
