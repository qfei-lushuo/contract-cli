# AI 变更记录：skills list 展示 CLI 版本

## 2026-05-26 skills list 版本口径统一

变更摘要：`contract-cli skills list` 中每个内置 skill 展示的版本改为当前 CLI 运行时版本，与 `contract-cli --version` 保持一致。

涉及文件/模块：

- `internal/cli/skills_command.go`
- `internal/cli/app_test.go`
- `docs/bot-command-development-guide.md`

关键逻辑/决策：

- `skills list` 不再使用 `SKILL.md` front matter 里的 `version` 作为展示版本。
- 展示版本统一从 `internal/build.Current().Version` 获取，和 `contract-cli --version` 同源。
- `SKILL.md` 的 `version` 字段保留为 skill 文档内部元数据，`skills install` 仍原样复制文件，不动态改写。
- 不新增 `cli_version` 或 `skill_version` 字段，避免引入第二套对外版本口径。

验证策略：

- 更新 `skills list` 测试，临时设置 CLI 版本为 `1.2.3`，确认所有 skill 行都展示 `1.2.3`。
- 同一测试确认 fixture 中的 `SKILL.md version` 不再出现在列表版本列。
- 保留隐藏禁用 skill 的断言，避免 `contract-cli-api-call` 被展示。
- 执行 `go test ./...` 验证整体回归。
