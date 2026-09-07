# CLI 临时 dev 联调

本分支暂时恢复 dev 支持，**上线前必须移除**。不是 beta 自动切换 dev，也不是维护第二套业务代码。

## 环境与隔离

| 项目 | prod（默认） | dev（显式选择） |
| --- | --- | --- |
| 建议 profile | contract | contract-dev（固定） |
| OpenAPI / resource | https://open.qfei.cn | https://dev-open.qtech.cn |
| OAuth 域名 | https://myaccount.qfei.cn | https://dev-myaccount.qtech.cn |

dev 使用相同的业务命令、每次请求前的环境识别 Hook、7 个来源 Header、`traceparent` 和 `X-Log-Id`。新增 dev profile 不切换默认 profile，也不修改生产授权。dev 的 OAuth、Token 刷新/撤销、业务请求及重定向只允许访问上述 dev 域名；不回退生产。

配置和授权按 profile 隔离。Device CredentialStore 本身以 profile 名构成存储键，dev 还要求保存的 Device profile 快照与 dev 环境匹配。不要把生产 Token 复制或改名为 dev。`CONTRACT_CLI_CONFIG_DIR` 可另选测试配置目录，但授权和查询必须始终使用同一目录及同一客户端会话。

## 安装和初始化

```bash
npm install -g /绝对路径/qfeius-contract-cli-测试版本.tgz
contract-cli --version
contract-cli skills install --force
contract-cli config add --env dev --name contract-dev
```

Skill 默认安装位置以命令输出为准。为其他客户端更新时，使用 `skills install --target <该客户端实际使用的技能目录> --force`，不要凭空猜目录。WorkBuddy 更新技能后退出并新建任务；其他客户端也应使用重新加载技能的新任务。

dev 初始化只读取 OAuth 元数据，不代表已授权。2026-09-04 已检查 dev OAuth 元数据可访问，授权、注册、Device、Token、撤销端点均为 dev。Device 公共 client ID 和 scope 来自部署仓库 `common-organization-v2/nacos/common-organization-v2-dev.yaml` 的 `oauth2.device-authorization.clients.contract`；服务端实际签发仍需真实授权验收。

## 授权与只读验证

在目标客户端新建任务，明确说明：

> 使用 contract-cli 在 dev 环境联调，只使用 contract-dev profile。请查询我有没有待处理的合同；未授权请先发起 dev 授权，不要访问生产环境。

支持 Device Grant 的客户端按现有流程执行：

```bash
contract-cli auth init --profile contract-dev --output json
# 展示链接后结束当前轮次；用户完成授权并回复后才执行一次：
contract-cli auth complete --profile contract-dev --output json
```

Device Grant 仍要求原有客户端会话运行环境，并未新增普通终端的 Device CredentialStore。普通本地终端/Codex 可沿用旧 Authorization Code + PKCE 流程，浏览器由用户完成 dev 登录：

```bash
contract-cli auth login --profile contract-dev --as user
```

如需 app 测试，在对话外安全配置 dev 凭据：仅读取 `CONTRACT_CLI_DEV_APP_ID` / `CONTRACT_CLI_DEV_APP_SECRET` 或该 profile 已存凭据，不继承生产 app 环境变量。

所有实际业务命令必须显式携带 `--profile contract-dev`。以客户端该次输出的 request trace 为线索，在 dev 服务端验证 Header 和日志；本地 mock 测试不等同于服务端透传成功。

## 上线前移除清单

- 删除 `resolveEnvironment` 的 dev 入口及 `dev_environment.go` 中临时预设、profile 和网络放行能力；同步恢复所有调用点的 prod-only 校验。
- 保留并运行生产 profile、pending Device 状态、Token 和网络层防混用回归测试；补回 `config add --env dev` 零网络拒绝测试。
- 恢复 Auth/Shared Skills、帮助和命令文档的 prod-only 边界，删除本临时说明入口。
- 从清理后的源码重新构建六个平台，不复用含 dev 能力的旧二进制；核验 tgz 的实际版本与命令行为。
- 不自动迁移、注销或删除用户的 dev 凭据；正式包应拒绝使用，用户需要时单独清理。
