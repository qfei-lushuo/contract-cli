---
name: contract-cli-rule
version: 1.0.0
description: "contract-cli 审批矩阵规则表技能：用 app 身份查询规则表、查询列头、创建/查询/搜索/更新/删除规则表行，以及预发布或发布规则表。当用户要使用 `contract-cli rule table ...` 操作 `/open-apis/rule_engine/v1` 审批矩阵接口时触发。"
---

# contract-cli Rule

CRITICAL — 开始前 MUST 先读取 [../contract-cli-shared/SKILL.md](../contract-cli-shared/SKILL.md)。

## 适用命令

- `contract-cli rule table list`
- `contract-cli rule table pre-release`
- `contract-cli rule table release`
- `contract-cli rule table column-headers list`
- `contract-cli rule table row create`
- `contract-cli rule table row get <row-id>`
- `contract-cli rule table row list`
- `contract-cli rule table row search`
- `contract-cli rule table row update <row-id>`
- `contract-cli rule table row delete <row-id>`

## 快速决策

- 先确认规则表：`rule table list --product-id --group-id`
- 查列头：`rule table column-headers list --product-id --group-id --table-id`
- 新增或修改规则行：`row create|update --input-file row.json`
- 按条件查规则行：`row search --input-file row-search.json`
- 发布前先预发布：`pre-release`，确认后再 `release`

## 关键规则

- 全部命令当前仅支持 `--as app`
- 公共定位参数：`--product-id`、`--group-id`
- 涉及单表时还需要 `--table-id`
- `row create`、`row search`、`row update` 必须传 JSON 请求体
- `pre-release`、`release` 支持可选 JSON body；当前自动化样例不发送请求体
- GET/DELETE 类命令不接受 `--input-file` / `--data`

## 示例

```bash
contract-cli rule table list --profile contract --as app --product-id <product-id> --group-id <group-id> --page-size 10
contract-cli rule table column-headers list --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id>
contract-cli rule table row create --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id> --input-file row.json
contract-cli rule table row search --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id> --input-file row-search.json
contract-cli rule table row update <row-id> --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id> --input-file row.json
contract-cli rule table row delete <row-id> --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id>
```

`row.json` 最小形态：

```json
{
  "table_cells": [
    {
      "table_column_id": "column-1",
      "table_cell_content_type": "STRING",
      "table_cell_content": {
        "string": "示例值"
      }
    }
  ]
}
```
