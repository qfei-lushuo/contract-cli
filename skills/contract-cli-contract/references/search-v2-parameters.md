# contract search-v2 Parameters

本页专用于 `contract-cli contract search-v2`。

- 接口：`POST /open-apis/contract/v1/contracts/searchV2`
- 身份：仅 `app`
- 请求体：JSON 必填，`--input-file` 与 `--data` 二选一
- 官方 OpenAPI：[搜索合同-V2](https://docs.qfei.cn/460498306e0.md)
- 校验基线：`contract-cli` 当前分支；CLM `master@334c18fc2e` 存在对应实现时，以 Controller、DTO 和业务校验补充官方规格。

## 目录

- [CLI 参数映射](#cli-参数映射)
- [请求体字段](#请求体字段)
- [枚举与约束](#枚举与约束)
- [示例](#示例)
- [来源差异说明](#来源差异说明)

## CLI 参数映射

| 参数 | 请求位置 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- | --- |
| --user-id-type | $query.user_id_type | string | 可选，默认 `user_id` | 用户 ID 类型 详情参考：用户身份体系<br>示例值："user_id"<br>可选值有：<br>- user_id：标识一个用户在某个租户内的身份。同一个用户在租户 A 和租户 B 内的 User ID 是不同的。在同一个租户内，一个用户的 User ID 在所有应用（包括商店应用）中都保持一致。User ID 主要用于在不同的应用间打通用户数据。<br>- union_id：标识一个用户在某个应用开发商下的身份。同一用户在同一开发商下的应用中的 Union ID 是相同的，在不同开发商下的应用中的 Union ID 是不同的。通过 Union ID，应用开发商可以把同个用户在多个应用中的身份关联起来。 |
| --input-file | $body | JSON file | 二选一必填 | 从文件读取 JSON；与 `--data` 互斥。 |
| --data | $body | JSON string | 二选一必填 | 内联 JSON；与 `--input-file` 互斥。 |
| --profile | 本地上下文 | string | 可选 | 不传时使用当前 profile。 |
| --as | 本地上下文 | enum | 可选 | 仅支持 `app`；不传时使用 profile 默认身份。 |
| --user-id | $query.user_id | string | 可选 | 传入时透传；MDM create/update 除外。 |
| --output | CLI 输出 | enum | 可选 | `json`、`yaml` 或 `table`；默认 `json`。 |
| --raw | CLI 输出 | boolean | 可选 | 原样输出服务端响应 body。 |

## 请求体字段

字段名、类型和服务端必填性来自官方 OpenAPI；“CLI 必填/禁止”是结构化命令的额外本地校验。父对象可选时，其内部必填字段标记为“父对象存在时必填”。

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| page_size | integer | 可选 | 分页大小，默认值：10。 |
| page_token | string | 可选 | 分页标记，第一次请求不填，表示从头开始遍历；分页查询结果还有更多项时会同时返回新的 page_token，下次遍历可采用该 page_token 获取查询结果<br>示例值："tblKz5D60T4JlfcT" |
| contract_number | string | 可选 | 合同编号。若传入该值，则忽略组合条件传参。开放平台普通输入支持模糊搜索，传入 ; 或 ； 分隔多个合同编号时按精确批量查询<br>示例值："CT20210708000030" 、"CT-001;CT-002；CT-003" |
| combine_condition | object | 可选 | 组合条件 |
| combine_condition.user_id | string | 可选 | 用户id<br>示例值："ou_8ebd4f35d7101ffdeb4771d7c8ec517e" |
| combine_condition.permission | integer | 可选 | 权限关系<br>示例值：0<br>可选值有：<br>0：全部合同<br>1：我的可见合同<br>2：申请的合同 |
| combine_condition.contract_status | integer | 可选 | 合同状态<br>示例值：3<br>可选值有：<br>0：editing<br>1：cancelled<br>2：revoked<br>3：process<br>4：rejected<br>5：approved<br>6：signing<br>7：signed<br>8：archiving<br>9：archived<br>10：changing<br>11：changed<br>12：我方已签约<br>13：对方已签约 |
| combine_condition.pay_type | integer | 可选 | 收支类型<br>示例值：0<br>可选值有：<br>0：未知<br>1：收入类<br>2：支出类<br>3：收入支出类<br>4：无金额 |
| combine_condition.contract_category_name | string | 可选 | 合同类型名称<br>示例值："技术咨询合同" |
| combine_condition.contract_name | string | 可选 | 合同名称<br>示例值："合同名称" |
| combine_condition.contract_number | string | 可选 | 合同编号<br>示例值："CT20210708000030" |
| combine_condition.archive_number | string | 可选 | 归档编号<br>示例值："AT20210708000030" |
| combine_condition.submited_time_start | string | 可选 | 查询提交时间范围值-起始值 |
| combine_condition.submited_time_end | string | 可选 | 查询提交时间范围值-结束值 |
| combine_condition.archived_time_start | string | 可选 | 查询归档时间范围值-起始值 |
| combine_condition.archived_time_end | string | 可选 | 查询归档时间范围值-结束值 |
| combine_condition.owner_user_id | string | 可选 | 合同归属人id<br>示例值："ou_b1b355b3099b3c1fd4c937e151c4fdd5" |
| combine_condition.form | array<object> | 可选 | 字段属性值匹配。仅支持关联前置单据搜索。 |
| combine_condition.form[].attribute_name | string | 父对象存在时必填 | — |
| combine_condition.form[].attribute_value | string | 父对象存在时必填 | — |
| combine_condition.form[].module_name | string | 父对象存在时必填 | — |
| combine_condition.contract_status_in | string | 可选 | 合同状态范围，数字含义参考contract_status，用逗号分割<br>示例值："1,2,3" |
| combine_condition.contract_category_abbreviation | string | 可选 | 合同类型缩写<br>示例值："CUBG" |
| combine_condition.submit_user_id | string | 可选 | — |
| combine_condition.create_time_start | string | 可选 | 查询创建时间范围值-起始值 |
| combine_condition.create_time_end | string | 可选 | 查询创建时间范围值-结束值 |
| combine_condition.update_time_start | string | 可选 | 查询更新时间范围值-起始值 |
| combine_condition.update_time_end | string | 可选 | 查询更新时间范围值-结束值 |
| logic_search | object | 可选 | 逻辑搜索，若传入combine_condition参数，忽略该字段<br>logic_search与combine_condition二选一 |
| logic_search.logic_type | integer | 父对象存在时必填 | 逻辑节点类型<br>- 0: and逻辑<br>- 1: or逻辑<br>- 2: 叶子节点 |
| logic_search.children | string | 可选 | 子条件 |
| logic_search.data | object | 可选 | 数据节点 |
| logic_search.data.field_name | string | 父对象存在时必填 | 字段名称，目前支持<br>- tradingPartyName: 交易方-对方名称<br>- contractNumber: 合同编号<br>- contractName: 合同名称<br>- categoryName: 合同类型名称<br>- contractStatus: 合同状态 |
| logic_search.data.operator | integer | 父对象存在时必填 | 操作符，目前支持<br>- 0: 等于<br>- 1: 不等于<br>- 6: 包含（模糊匹配）<br>- 7: 不包含 （模糊不匹配） |
| logic_search.data.field_value | object | 父对象存在时必填 | 字段值 |

## 枚举与约束

- CLM `master` 的 `ContractSearchDTO` 还包含 MCP 搜索字段；`searchV2` 会忽略 `condition_units` 和 `filter_units`。
- `page_size`（integer，可选）：分页大小，默认值：10。
- `combine_condition.permission`（integer，可选）：权限关系<br>示例值：0<br>可选值有：<br>0：全部合同<br>1：我的可见合同<br>2：申请的合同
- `combine_condition.contract_status`（integer，可选）：合同状态<br>示例值：3<br>可选值有：<br>0：editing<br>1：cancelled<br>2：revoked<br>3：process<br>4：rejected<br>5：approved<br>6：signing<br>7：signed<br>8：archiving<br>9：archived<br>10：changing<br>11：changed<br>12：我方已签约<br>13：对方已签约
- `combine_condition.pay_type`（integer，可选）：收支类型<br>示例值：0<br>可选值有：<br>0：未知<br>1：收入类<br>2：支出类<br>3：收入支出类<br>4：无金额
- `combine_condition.submited_time_start`（string，可选）：查询提交时间范围值-起始值
- `combine_condition.submited_time_end`（string，可选）：查询提交时间范围值-结束值
- `combine_condition.archived_time_start`（string，可选）：查询归档时间范围值-起始值
- `combine_condition.archived_time_end`（string，可选）：查询归档时间范围值-结束值
- `combine_condition.contract_status_in`（string，可选）：合同状态范围，数字含义参考contract_status，用逗号分割<br>示例值："1,2,3"
- `combine_condition.create_time_start`（string，可选）：查询创建时间范围值-起始值
- `combine_condition.create_time_end`（string，可选）：查询创建时间范围值-结束值
- `combine_condition.update_time_start`（string，可选）：查询更新时间范围值-起始值
- `combine_condition.update_time_end`（string，可选）：查询更新时间范围值-结束值

## 示例

```bash
contract-cli contract search-v2 --input-file request.json --profile contract --as app
```

官方请求体示例（动态字段接口仍须以当前租户配置为准）：

```json
{
  "page_size": 0,
  "page_token": "tblKz5D60T4JlfcT",
  "contract_number": "CT20210708000030",
  "combine_condition": {
    "user_id": "d84df55b",
    "permission": 0,
    "contract_status": 3,
    "pay_type": 0,
    "contract_category_name": "技术咨询合同",
    "contract_name": "合同名称",
    "contract_number": "CT20210708000030",
    "archive_number": "AT20210708000030",
    "create_time_start": "2021-10-01 11:11:11",
    "create_time_end": "2021-10-01 11:11:11",
    "update_time_start": "2021-10-01 11:11:11",
    "update_time_end": "2021-10-01 11:11:11",
    "submited_time_start": "2021-10-01 11:11:11",
    "submited_time_end": "2021-10-01 11:11:11",
    "archived_time_start": "2021-10-01 11:11:11",
    "archived_time_end": "2021-10-01 11:11:11",
    "owner_user_id": "d84df55b",
    "form": [
      {
        "attribute_name": "受理人",
        "attribute_value": "6982118143107809580",
        "module_name": "相关单据"
      },
      {
        "attribute_name": "受理人",
        "attribute_value": "6982118143107809580",
        "module_name": "相关单据"
      },
      {
        "attribute_name": "受理人",
        "attribute_value": "6982118143107809580",
        "module_name": "相关单据"
      }
    ],
    "contract_status_in": 123,
    "contract_category_abbreviation": "CUBG"
  }
}
```

## 来源差异说明

- 官方规格路径：`POST /open-apis/contract/v1/contracts/searchV2`
- 本页描述请求参数；响应 envelope 和输出格式遵循共享 Skill。
