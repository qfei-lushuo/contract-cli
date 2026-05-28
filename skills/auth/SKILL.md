---
name: auth
version: 1.1.1
description: "contract-cli 登录与身份切换技能：初始化 prod profile、执行 user OAuth 登录、录入 app 的 app_id/app_secret 并立即兑换 tenant_access_token、查看状态、切换默认身份、排查本地 config/secrets 持久化问题。当用户需要 `contract-cli config add`、`contract-cli auth login --as user|app`、`contract-cli auth status/logout/use` 或排查登录异常时触发。"
---

# contract-cli Auth

本技能指导你如何在本仓库中使用 `contract-cli` 的登录与身份切换能力，并保持和当前实现一致。

## 适用范围

- 首次初始化本地 `profile`
- 以 `user` 身份走 OAuth 授权码 + PKCE 登录
- 以 `app` 身份录入 `app_id/app_secret`
- 查看或清理本地身份状态
- 切换默认业务身份
- 排查 `config.json` 和 `secrets.json` 的本地持久化问题

## 实现来源

- [internal/cli/app.go](../../internal/cli/app.go)
- [internal/cli/auth_provider.go](../../internal/cli/auth_provider.go)
- [internal/config/store.go](../../internal/config/store.go)
- [internal/config/secrets.go](../../internal/config/secrets.go)

## 配置初始化

首次使用前，必须先执行：

```bash
contract-cli config add --env prod --name contract
```

当前实现仅内置 `prod` 环境，默认环境为 `prod`，默认 profile 名为 `contract`。该命令会：

- 发现 well-known 元数据
- 保存 MCP server / resource / OAuth server 配置
- 将 `default_identity` 初始化为 `user`

## 身份模型

同一个 profile 下维护两种身份：

| 身份 | 命令 | 本地存储 | 当前实现 |
|------|------|----------|----------|
| `user` | `contract-cli auth login --as user` | `profiles.<name>.identities.user.token` | 已实现 OAuth 登录 |
| `app` | `contract-cli auth login --as app` | `profiles.<name>.identities.app` + `secrets.json` | 已实现凭据录入和 `tenant_access_token` 兑换 |

额外还有一个默认身份指针：

- `profiles.<name>.default_identity`
- 由 `contract-cli auth use --as user|app` 修改
- `auth login --as ...` 成功后也会自动切换到对应身份

## 快速流程

### `user` 登录

```bash
contract-cli config add --env prod --name contract
contract-cli auth login --as user
contract-cli auth status --as user
```

行为约束：

- `user` 登录会自动注册 `client_id`
- 使用授权码模式 + PKCE
- 会启动本地回调服务，回调地址来自 profile 中的 `redirect_url`
- 默认自动打开浏览器；如需仅打印链接，使用 `--no-open-browser`
- 登录成功后写入 `identities.user.token`

### `app` 登录

```bash
contract-cli auth login --as app --app-id "<app_id>" --app-secret "<app_secret>"
contract-cli auth status --as app
```

也可以通过环境变量提供凭据：

```bash
export CONTRACT_CLI_APP_ID="<app_id>"
export CONTRACT_CLI_APP_SECRET="<app_secret>"
contract-cli auth login --as app
```

运行时也兼容旧变量 `CONTRACT_CLI_BOT_APP_ID` / `CONTRACT_CLI_BOT_APP_SECRET` 和 `DEMOCLI_BOT_APP_ID` / `DEMOCLI_BOT_APP_SECRET`，但后续新增配置统一使用 `CONTRACT_CLI_APP_ID` / `CONTRACT_CLI_APP_SECRET`。

为兼容老用户脚本，旧命令 `--as bot` 仍可执行，运行时等价于 `--as app`；新命令示例和解释统一使用 `app`。

凭据优先级固定为：

- 命令行参数
- 环境变量
- 本地已保存凭据

行为约束：

- `app` 登录会先保存 `app_id/app_secret`，再调用 `tenant_access_token/internal` 兑换 token
- `app_secret` 不写入 `config.json`
- token 成功后写入 `identities.app.token`
- token 兑换失败时，会保留新凭据，但不会切换默认身份到 `app`
- 登录成功后会切换 `default_identity=app`

## 状态、退出与切换

### 查看状态

```bash
contract-cli auth status --as user
contract-cli auth status --as app
```

规则：

- 不传 `--as` 时，`auth status` 默认查看 `user`
- `user` 显示 `authorized` 或 `unauthorized`
- `app` 显示 `authorized`、`expired`、`configured` 或 `unconfigured`
- `app` 状态会显示 `Token Protocol: tenant_access_token/internal` 和过期时间（若有），不展示 endpoint 地址

### 退出登录

```bash
contract-cli auth logout --as user
contract-cli auth logout --as app
```

规则：

- `logout --as user` 只清理 `user.token`
- `logout --as app` 只清理 `app.token`，保留 `app_id/app_secret` 和对应 secret
- 不传 `--as` 时，`auth logout` 默认处理 `user`

### 切换默认身份

```bash
contract-cli auth use --as user
contract-cli auth use --as app
```

规则：

- 该命令只修改 `default_identity`
- 不会重新登录
- 不会校验目标身份一定已拿到 token

## 本地文件

默认路径如下，若设置了 `CONTRACT_CLI_CONFIG_DIR`，则改为该目录：

- `~/.contract-cli/config.json`
- `~/.contract-cli/secrets.json`

运行时也兼容旧的 `DEMOCLI_CONFIG_DIR` 以及历史默认目录 `~/.democli`，用于平滑读取已有本地登录态。

存储约束：

- `config.json` 保存 profile、identity 元数据和 token
- `secrets.json` 只保存 app 的 `app_secret`
- `user.token` 与 `app.token` 分离存储，不共享
- 旧版平铺 OAuth 字段会自动迁移到 `identities.user`

## 安全规则

- 禁止在终端或文档中明文输出 `app_secret`、`access_token`、`refresh_token`
- 不要把 `app` 登出描述成“删除凭据”，当前实现只清 token、不删 `app_id/app_secret`
- 不要让用户误以为 `default_identity` 会影响 `auth status` 或 `auth logout` 的默认目标，这两个命令未传 `--as` 时仍按 `user`
- 涉及写入、清理本地凭据时，先确认是在当前 profile 上操作

## 故障排查

- `user identity is not configured`：先执行 `contract-cli config add --env prod --name contract`
- 浏览器未自动打开：改用 `--no-open-browser`，手动访问输出的授权链接
- 回调超时：检查 `redirect_url` 对应端口是否可监听，必要时调大 `--timeout`
- app 凭据不完整：补齐 `--app-id/--app-secret` 或设置 `CONTRACT_CLI_APP_ID/CONTRACT_CLI_APP_SECRET`
- app 登录提示缺少 `app_token_endpoint`：说明 profile 过旧，重跑 `contract-cli config add --env prod --name <profile>`
- app 状态显示 `expired`：重新执行 `contract-cli auth login --as app`
- 旧脚本仍传 `--as bot`：可以继续执行；后续新脚本请改写为 `--as app`
