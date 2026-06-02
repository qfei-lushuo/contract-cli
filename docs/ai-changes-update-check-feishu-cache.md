# AI 变更记录：飞书式更新提示缓存

## 变更摘要

- 将普通业务命令的自动升级提示改为飞书式模型：同步读取本地 `update-check.json` 决定是否注入 `_notice.update`，后台异步刷新远端版本缓存。
- 保留手动 `contract-cli update check` 的同步远端检查行为。
- 保留 24 小时 TTL；cache fresh 时后台不访问 registry，因此刚发布的新包不会在当前命令中立刻提示。

## 关键逻辑

- 新增 `CheckCached` 只读缓存、不发网络请求。
- 新增 `RefreshCache` 按 TTL 刷新缓存，失败不阻断业务命令、不写失败缓存。
- `maybePrepareUpdateNotice` 先从缓存生成 notice，再启动后台刷新；后台刷新不修改当前命令的 notice。

## 验证

- 覆盖缓存可升级时当前命令立即注入 `_notice.update`。
- 覆盖 stale 缓存后台刷新并写入新版本。
- 覆盖 fresh 无更新缓存不访问 registry、不立即提示新包。
- 覆盖后台刷新失败不阻断业务命令。
