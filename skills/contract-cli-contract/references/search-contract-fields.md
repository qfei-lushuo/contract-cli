# Contract Search Routing

`contract-cli` 的合同搜索按身份和搜索语义路由到三套接口。先选接口，再读取对应参数文档；不要把三套请求体字段混用。

## 接口选择

| 用户意图 | 命令 | 接口 | 参数文档 |
| --- | --- | --- | --- |
| 以当前登录用户的可见权限搜索，使用 PC 兼容关键词或结构化筛选 | `contract search --as user` | `POST /open-apis/contract/v1/mcp/contracts/search` | [search-user-parameters.md](search-user-parameters.md) |
| app 按合同编号精确查询，或使用旧组合条件、逻辑条件 | `contract search --as app` | `POST /open-apis/contract/v1/contracts/search` | [search-app-parameters.md](search-app-parameters.md) |
| app 按合同编号做 ES 模糊查询或批量查询，并分页返回多条结果 | `contract search-v2 --as app` | `POST /open-apis/contract/v1/contracts/searchV2` | [search-v2-parameters.md](search-v2-parameters.md) |

## 不可混用的关键语义

- `contract search --as user` 才支持 `condition_units`、`filter_units`、`search_tab_code` 和 MCP 排序字段；权限由当前 OAuth 用户决定。
- `contract search --as app` 的顶层 `contract_number` 是精确查询，最多返回一条；未命中返回业务错误 `110107`。
- `contract search-v2 --as app` 的顶层 `contract_number` 是 ES 模糊查询；传分号分隔的多个编号时进入精确批量查询。
- app V2 不是 user MCP 搜索的替代品：当前后端明确忽略 `condition_units` 和 `filter_units`。
- app V1 与 app V2 都保留 `combine_condition` / `logic_search`，但 app V2 只有在未传顶层 `contract_number` 时才委托旧搜索逻辑。

## CLI 公共行为

- `contract search` 支持 `--contract-number`、`--page-size`、`--page-token`；这些 flag 会覆盖并写入 JSON body 的同名字段。
- 复杂请求体使用 `--input-file`，简单请求可使用 `--data`；两者互斥。
- 示例必须显式携带 `--as user` 或 `--as app`，避免 profile 默认身份改变实际接口。
- 响应字段入口见 [contract-response-fields.md](contract-response-fields.md)。
