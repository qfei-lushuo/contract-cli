# 环境包构建与测试

每个发布包固定一个环境，沿用原有 profile、配置文件和授权流程。`BUILD_ENVIRONMENT` 只由构建脚本读取，通过 Go ldflags 注入；运行时设置同名环境变量不会切换环境。省略构建参数时仍为 prod。

```sh
BUILD_ENVIRONMENT=test bash scripts/build-release-assets.sh
```

支持构建 prod/dev/test/blue；使用对应环境原有公开地址。config add 的 --env 如传入，必须与包一致；旧 profile 或跨环境 URL 不可用于授权和业务请求。原有版本检查和更新流程保持不变；测试时应选择同环境的新包。

Contract test 默认 profile 为 contract-test，执行 `contract-cli config add` 即可初始化。配置仍保存在原有目录，其他 profile 内容不被迁移或改写。dev/test/blue 的 Device client 沿用 prod 的 `zscli_892efdadc11a3f53`；浏览器 OAuth 保留原有动态注册流程，与 Device client 分开保存。

验证命令：`contract-cli version` 显示 environment test；`contract-cli auth status` 应读取 contract-test；`contract-cli config add --env prod` 应拒绝。真实授权由使用者按现有 Skill 执行。

app 身份沿用原有 `CONTRACT_CLI_APP_ID` / `CONTRACT_CLI_APP_SECRET`、兼容别名和参数；请在测试进程中配置目标环境自己的凭据。版本检查保留 release 的 `contract-cli update check [--channel latest|beta]`，不由 CLI 自动安装新版本。
