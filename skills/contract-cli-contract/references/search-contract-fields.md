# Contract Search Fields

这份文档是 `contract-cli contract search` 的字段主入口，用来把“我要按什么条件找合同”翻译成可执行 JSON。

官方来源：[搜索合同](https://docs.qfei.cn/366415247e0)

## 1. CLI 命令面与硬约束

```bash
contract-cli contract search --profile contract --as app --input-file contract-search.json
```

也可以直接传 JSON 字符串：

```bash
contract-cli contract search --profile contract --as user --data '{"contract_number":"CT20210708000030"}'
```

硬约束：

- `--as user` 走 `/open-apis/contract/v1/mcp/contracts/search`
- `--as app` 走 `POST /open-apis/contract/v1/contracts/search`
- `--input-file` 与 `--data` 互斥
- `--contract-number`、`--page-size`、`--page-token` 会合并进 JSON body
- 顶层 `contract_number` 一旦传入，后端会忽略组合条件
- `combine_condition` 与 `logic_search` 二选一；传入 `combine_condition` 时忽略 `logic_search`

## 2. 最小请求体

按合同编号精确搜索：

```json
{
  "contract_number": "CT20210708000030"
}
```

按合同名称和状态组合搜索：

```json
{
  "page_size": 10,
  "combine_condition": {
    "permission": 0,
    "contract_name": "采购合同",
    "contract_status": 3
  }
}
```

分页：

```json
{
  "page_size": 20,
  "page_token": "tblKz5D60T4JlfcT",
  "combine_condition": {
    "permission": 0
  }
}
```

## 3. 顶层字段

| 字段 | 类型 | 必填性 | 业务含义 | 注意 |
| --- | --- | --- | --- | --- |
| `page_size` | `integer` | 可选 | 分页大小，默认值通常为 `10`。 | CLI flag `--page-size` 会写入该字段。 |
| `page_token` | `string` | 可选 | 分页标记。 | 首次请求不传；翻页时使用响应 `data.page_token`。 |
| `contract_number` | `string` | 可选 | 合同编号。 | 顶层传入时忽略 `combine_condition` 和 `logic_search`。 |
| `combine_condition` | `object` | 可选 | 常规组合条件。 | 与 `logic_search` 二选一。 |
| `logic_search` | `object` | 可选 | 逻辑条件树。 | 适合 AND/OR/叶子节点组合查询。 |

## 4. `combine_condition`

| 字段 | 类型 | 必填性 | 业务含义 | 示例 |
| --- | --- | --- | --- | --- |
| `user_id` | `string` | 可选 | 用户 id。 | `ou_xxx` |
| `permission` | `integer` | 可选 | 权限关系。 | `0` 全部合同；`1` 我的可见合同；`2` 申请的合同 |
| `contract_status` | `integer` | 可选 | 单个合同状态。 | `3` |
| `contract_status_in` | `string` | 可选 | 多个合同状态。 | `"1,2,3"` |
| `pay_type` | `integer` | 可选 | 收支类型。 | `1` 收入类；`2` 支出类 |
| `contract_category_name` | `string` | 可选 | 合同类型名称。 | `"技术咨询合同"` |
| `contract_category_abbreviation` | `string` | 可选 | 合同类型缩写。 | `"CUBG"` |
| `contract_name` | `string` | 可选 | 合同名称。 | `"合同名称"` |
| `contract_number` | `string` | 可选 | 合同编号。 | `"CT20210708000030"` |
| `archive_number` | `string` | 可选 | 归档编号。 | `"AT20210708000030"` |
| `owner_user_id` | `string` | 可选 | 合同归属人 id。 | `ou_xxx` |
| `submit_user_id` | `string` | 可选 | 提交人 id。 | `ou_xxx` |
| `create_time_start` / `create_time_end` | `string` | 可选 | 创建时间范围。 | `"2021-10-01 11:11:11"` |
| `update_time_start` / `update_time_end` | `string` | 可选 | 更新时间范围。 | `"2021-10-01 11:11:11"` |
| `submited_time_start` / `submited_time_end` | `string` | 可选 | 提交时间范围。 | 官方字段拼写是 `submited`。 |
| `archived_time_start` / `archived_time_end` | `string` | 可选 | 归档时间范围。 | `"2021-10-01 11:11:11"` |
| `form` | `array<object>` | 可选 | 字段属性值匹配。 | 仅支持关联前置单据搜索。 |

状态枚举：

| 值 | 含义 |
| --- | --- |
| `0` | editing |
| `1` | cancelled |
| `2` | revoked |
| `3` | process |
| `4` | rejected |
| `5` | approved |
| `6` | signing |
| `7` | signed |
| `8` | archiving |
| `9` | archived |
| `10` | changing |
| `11` | changed |
| `12` | 我方已签约 |
| `13` | 对方已签约 |

收支类型：

| 值 | 含义 |
| --- | --- |
| `0` | 未知 |
| `1` | 收入类 |
| `2` | 支出类 |
| `3` | 收入支出类 |
| `4` | 无金额 |

`form` 子项：

| 字段 | 类型 | 必填性 | 业务含义 |
| --- | --- | --- | --- |
| `attribute_name` | `string` | 必填 | 匹配字段名。 |
| `attribute_value` | `string` | 必填 | 字段值。 |
| `module_name` | `string` | 必填 | 模块名。 |

## 5. `logic_search`

`logic_search` 用于表达逻辑条件树：

```json
{
  "logic_search": {
    "logic_type": 0,
    "children": "[]"
  }
}
```

字段说明：

| 字段 | 类型 | 必填性 | 业务含义 |
| --- | --- | --- | --- |
| `logic_type` | `integer` | 必填 | `0` AND；`1` OR；`2` 叶子节点。 |
| `children` | `string` | 可选 | 子条件。官方文档定义为字符串。 |
| `data.field_name` | `string` | 叶子节点必填 | 字段名。 |
| `data.operator` | `integer` | 叶子节点必填 | 操作符。 |
| `data.field_value` | `object` | 叶子节点必填 | 字段值。 |

当前支持的 `field_name`：

- `tradingPartyName`：交易方-对方名称
- `contractNumber`：合同编号
- `contractName`：合同名称
- `categoryName`：合同类型名称
- `contractStatus`：合同状态

当前支持的 `operator`：

- `0`：等于
- `1`：不等于
- `6`：包含，模糊匹配
- `7`：不包含，模糊不匹配

## 6. 响应读取

常用路径：

- `data.items[]`：合同列表
- `data.has_more`：是否还有下一页
- `data.page_token`：下一页令牌
- `data.items[].contract_id`：合同 id
- `data.items[].contract_number`：合同编号
- `data.items[].contract_name`：合同名称
- `data.items[].contract_status_code` / `contract_status_name`：状态

更完整的合同字段看 [contract-response-fields.md](contract-response-fields.md)。

