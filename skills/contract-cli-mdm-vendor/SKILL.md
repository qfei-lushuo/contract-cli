---
name: contract-cli-mdm-vendor
version: 1.0.1
description: "contract-cli 交易方主数据技能：列出交易方候选列表、按 ID 获取详情，或用 app 身份创建、更新、全量分页查询、按证件 ID 查询交易方。当用户要使用 `contract-cli mdm vendor ...` 操作合同域交易方数据时触发。"
---

# contract-cli MDM Vendor

CRITICAL — 开始前 MUST 先读取 [../contract-cli-shared/SKILL.md](../contract-cli-shared/SKILL.md)。

## 适用命令

- `contract-cli mdm vendor list`
- `contract-cli mdm vendor get <vendor-id>`
- `contract-cli mdm vendor create`
- `contract-cli mdm vendor update <vendor-id>`
- `contract-cli mdm vendor list-all`
- `contract-cli mdm vendor query-by-cert`

## 快速决策

- 已知交易方 ID：直接用 `mdm vendor get`
- 只知道名称或想拿候选列表：用 `mdm vendor list`
- 想创建或更新交易方：用 `mdm vendor create|update --as app --user-id <operator-user-id> --input-file ...`
- 想按证件号精确查：用 `mdm vendor query-by-cert --as app`
- 创建或更新前如果不确定字段：先切到 [../contract-cli-mdm-fields/SKILL.md](../contract-cli-mdm-fields/SKILL.md)

## 关键规则

- `mdm vendor list` 支持 `--name`、`--page-size`、`--page-token`
- `mdm vendor list --as user` 走 `/open-apis/contract/v1/mcp/vendors`
- `mdm vendor list --as app` 走 `/open-apis/mdm/v1/vendors`
- `mdm vendor get --as user` 走 `/open-apis/contract/v1/mcp/vendors/{vendor_id}`
- `mdm vendor get --as app` 走 `/open-apis/mdm/v1/vendors/{vendor_id}`
- `mdm vendor create/update/list-all/query-by-cert` 当前仅支持 `--as app`
- `mdm vendor create` 走 `POST /open-apis/mdm/v1/vendors`
- `mdm vendor update` 走 `PUT /open-apis/mdm/v1/vendors/{vendor_id}`
- `mdm vendor list-all` 走 `GET /open-apis/mdm/v1/vendors/list_all`
- `mdm vendor query-by-cert` 走 `GET /open-apis/mdm/v1/vendors/query_vendors`
- 不暴露 `--operator`
- `mdm vendor create/update` 必须传 `--user-id`；`--user-id-type` 不传时默认 `user_id`
- `mdm vendor create` 请求体不要包含后端生成的 `vendor` 编码
- `mdm vendor update` 请求体必须包含后端返回的 `id` 和 `vendor` 编码
- 创建/更新请求体除上述规则外仍直接透传 JSON，其他字段是否必填受后台动态字段配置影响
- 推荐阅读顺序是：
  - 先读 [references/vendor-query-guide.md](references/vendor-query-guide.md) 选查询场景
  - 再读 [references/vendor-query-parameters.md](references/vendor-query-parameters.md) 查请求参数映射
  - 最后读 [references/commands.md](references/commands.md) 抄命令示例
- P3 命令按需读取完整参数文档：
  - 创建：[references/vendor-create-parameters.md](references/vendor-create-parameters.md)
  - 更新：[references/vendor-update-parameters.md](references/vendor-update-parameters.md)
  - 全量分页：[references/vendor-list-all-parameters.md](references/vendor-list-all-parameters.md)
  - 按证件查询：[references/vendor-query-by-cert-parameters.md](references/vendor-query-by-cert-parameters.md)

## 实现来源

- [internal/cli/vendor_command.go](../../internal/cli/vendor_command.go)
- [internal/openplatform/mdmvendor/service.go](../../internal/openplatform/mdmvendor/service.go)
- [references/vendor-query-guide.md](references/vendor-query-guide.md)
- [references/vendor-query-parameters.md](references/vendor-query-parameters.md)
- [references/commands.md](references/commands.md)
- [references/vendor-create-parameters.md](references/vendor-create-parameters.md)
- [references/vendor-update-parameters.md](references/vendor-update-parameters.md)
- [references/vendor-list-all-parameters.md](references/vendor-list-all-parameters.md)
- [references/vendor-query-by-cert-parameters.md](references/vendor-query-by-cert-parameters.md)

## 操作建议

- 只知道名称时，先 `mdm vendor list` 拿候选，再 `mdm vendor get` 查详情
- 创建合同前如果只是要选对方主体，优先记住交易方 id
- 创建或更新交易方前，先用 `mdm fields list --biz-line vendor` 确认动态字段
- 想确认写接口字段定义时，切到 `mdm fields list`

## 不要这样做

- 不要把 `mdm vendor list/get` 当成字段配置查询
- 不要在 create 请求体里传后端生成的 `vendor` 编码
- 不要在 create/update 请求体里记录 token、密钥或个人敏感信息之外的无关内容
