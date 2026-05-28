# 统一将 bot 身份改名为 app

## 变更摘要

- 将推荐身份值从 `bot` 改为 `app`，新文档和 help 统一展示 `--as app`。
- 为避免影响老用户，CLI 继续兼容旧参数 `--as bot`，运行时按 `app` 身份处理。
- 配置文件写入新字段 `app_token_endpoint`、`identities.app`、`default_identity=app`。
- 兼容读取旧配置里的 `bot_token_endpoint`、`identities.bot`、`default_identity=bot`，保存后归一为 app 字段。
- 新增 `CONTRACT_CLI_APP_ID` / `CONTRACT_CLI_APP_SECRET`，旧 `CONTRACT_CLI_BOT_APP_ID` / `CONTRACT_CLI_BOT_APP_SECRET` 继续作为 fallback。

## 关键逻辑

- `ParseIdentityKind("bot")` 作为旧别名返回 `IdentityApp`，但新文档不再主推旧称。
- app 登录仍使用 `app_id/app_secret` 换取 `tenant_access_token/internal`，开放平台接口路径不变。
- app-only 命令沿用原有身份限制语义，只把错误提示更新为 `only supports --as app`。
- 旧 `profile.bot.app_secret` 仍可通过旧配置中的 `secret_ref` 读取；下一次 `auth login --as app` 会写入新的 `profile.app.app_secret`。

## 验证

- 新增和更新测试覆盖身份解析、`--as bot` 旧别名、旧配置迁移、app 登录、旧环境变量 fallback、app-only 拦截和文档口径。
- 运行 `go test ./...` 验证整体回归。
