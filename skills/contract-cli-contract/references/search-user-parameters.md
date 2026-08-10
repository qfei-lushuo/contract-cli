# contract search User MCP Parameters

本页专用于 user 身份下的 MCP 合同搜索。

- 命令：`contract-cli contract search --profile contract --as user`
- 接口：`POST /open-apis/contract/v1/mcp/contracts/search`
- 身份：仅本页语义使用 `user` OAuth 身份
- 参数基线：MCP 工具 `search-contracts` 描述；CLM `MCPContractOpenPlatformController` 与 `McpContractSearchRequestDTO`
- 权限：服务端自动注入当前登录用户的可见范围，不需要在 body 中传 `user_id` 或 `permission`

## 目录

- [CLI 参数映射](#cli-参数映射)
- [请求体字段](#请求体字段)
- [组合条件](#组合条件)
- [关键词条件 condition_units](#关键词条件-condition_units)
- [结构化筛选 filter_units](#结构化筛选-filter_units)
- [排序与分页](#排序与分页)
- [枚举与约束](#枚举与约束)
- [动态字段发现](#动态字段发现)
- [示例](#示例)
- [响应摘要](#响应摘要)

## CLI 参数映射

| 参数 | 请求位置 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- | --- |
| `--as user` | 本地身份 | enum | 建议显式传 | 使用当前 OAuth 用户，并固定路由到 MCP 搜索。 |
| `--contract-number` | `$body.contract_number` | string | 可选 | 覆盖请求体同名字段。 |
| `--page-size` | `$body.page_size` | integer | 可选 | 覆盖请求体同名字段。 |
| `--page-token` | `$body.page_token` | string | 可选 | 覆盖请求体同名字段。 |
| `--input-file` | `$body` | JSON file | 与 `--data` 互斥 | 复杂条件推荐使用。 |
| `--data` | `$body` | JSON string | 与 `--input-file` 互斥 | 适合简单查询。 |
| `--user-id-type` | `$query.user_id_type` | string | 服务支持可选 | MCP 规格支持 `user_id` / `union_id`，默认 `user_id`；当前 CLI tool spec 固定 `user_id`，显式 flag 暂不能覆盖该默认值。 |
| `--user-id` | `$query.user_id` | string | 不需要 | MCP 根据 OAuth 登录态识别用户；不要用它代替 user 登录。 |
| `--profile` | 本地上下文 | string | 可选 | 不传时使用当前 profile。 |
| `--output` | CLI 输出 | enum | 可选 | `json`、`yaml`、`table`，默认 `json`。 |
| `--raw` | CLI 输出 | boolean | 可选 | 原样输出服务端响应。 |

除前三个 body flag 外，其余搜索字段都通过 `--input-file` 或 `--data` 放入 JSON body。

## 请求体字段

顶层参数都不是 schema 必填项；空对象表示在当前用户可见范围内按默认页签、排序和分页查询。

| JSON 路径 | 类型 | 必填性 | 默认值/约束 | 说明 |
| --- | --- | --- | --- | --- |
| `contract_number` | string | 可选 | 非空关键词 | 转成 `CONTRACT_NUMBER_FIELD` 的 `SHOULD` 条件。 |
| `combine_condition` | object | 可选 | 旧基础字段模型 | MCP/PC 兼容搜索优先使用 `condition_units` / `filter_units`。 |
| `condition_units` | array<object> | 可选 | 最多 50 项 | PC 兼容关键词条件。 |
| `filter_units` | array<object> | 可选 | 最多 50 项 | PC 兼容结构化筛选条件。 |
| `sort_type` | string | 可选 | 默认 `SUBMITTED_TIME` | MCP 排序字段。 |
| `order` | string | 可选 | 默认 `DESC` | `ASC` 或 `DESC`。 |
| `sort` | string | 可选 | 旧兼容字段 | 仅 `asc` 表示升序；否则按提交时间降序。优先使用 `sort_type/order`。 |
| `search_tab_code` | integer | 可选 | 默认 `0`；枚举 `0/1` | `0` 全部合同；`1` 我的合同；暂不支持草稿页签。 |
| `page_size` | integer | 可选 | 默认 `10`；建议不超过 `50` | 合同编号分号批量模式最大 `50`。 |
| `page_token` | string | 可选 | 首次不传 | 翻页时原样回传响应中的 token。 |
| `lang` | string | 可选 | 后端默认语言 | 例如 `zh-CN`、`en-US`。 |

`user_id_type` 位于 query，不属于 JSON body。

## 组合条件

`combine_condition` 用于旧开放接口基础字段。MCP 会自行填充权限，因此本模型不列出 `user_id` 和 `permission`。

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `combine_condition.archive_number` | string | 可选 | 归档编号，精确匹配。 |
| `combine_condition.contract_category_abbreviation` | string | 可选 | 合同分类缩写，精确匹配。 |
| `combine_condition.contract_category_name` | string | 可选 | 合同分类名称，精确匹配。 |
| `combine_condition.contract_name` | string | 可选 | 合同名称，支持模糊匹配。 |
| `combine_condition.contract_number` | string | 可选 | 旧组合条件编号语义；新搜索优先使用顶层编号或搜索单元。 |
| `combine_condition.contract_status_in` | string | 可选 | 状态 code，英文逗号分隔。 |
| `combine_condition.pay_type` | integer | 可选 | `1` 收入、`2` 支出、`3` 收入支出、`4` 无金额。 |
| `combine_condition.currency_in` | array<string> | 可选 | 币种英文码，例如 `["CNY","USD"]`。 |
| `combine_condition.demand_employee_ids` | array<string> | 可选 | 需求人 employeeId；当前后端只允许传 1 个。 |
| `combine_condition.saas_demand_department_id_in` | array<string> | 可选 | 需求人部门 ID 集合。 |
| `combine_condition.use_template` | integer | 可选 | `1` 使用模板；`0` 不使用模板。 |
| `combine_condition.create_time_start/end` | string | 可选 | 创建时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.update_time_start/end` | string | 可选 | 更新时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.submited_time_start/end` | string | 可选 | 提交时间范围；保留接口的 `submited` 拼写。 |
| `combine_condition.archived_time_start/end` | string | 可选 | 归档时间范围。 |
| `combine_condition.contract_sign_date_start/end` | string | 可选 | 签订时间 `signed_time` 范围。 |
| `combine_condition.contract_signed_date_start/end` | string | 可选 | 签约日期 `signed_date` 范围。 |

所有时间字符串使用 `YYYY-MM-DD HH:mm:ss`。

## 关键词条件 `condition_units`

单项结构：

| 字段 | 类型 | schema 必填性 | 执行约束 |
| --- | --- | --- | --- |
| `search_field` | string | 未标必填 | 有效条件必须提供；可传枚举名或 PC 字段名。 |
| `search_value` | string/number/boolean/array/object | 未标必填 | 有效条件必须提供；绝大多数关键词字段传 string。 |
| `union_type` | string | 可选 | `MUST`、`SHOULD`、`MUST_NOT`；关键词默认 `SHOULD`。 |
| `filter_unique_key` | string | 可选 | 通常不传；限定具体自定义字段时传 `attribute_key`。 |

可用关键词字段：

| 枚举名 | PC 字段名 | 搜索值与说明 |
| --- | --- | --- |
| `CONTRACT_NAME_FIELD` | `contractName` | 合同名称关键词 string。 |
| `CONTRACT_NUMBER_FIELD` | `contractNumber` | 合同编号 string；包含 `;` 或 `；` 时进入批量编号模式，`page_size<=50`。 |
| `CONTRACT_CREATOR_FIELD` | `contractCreator` | 创建人姓名关键词。 |
| `CONTRACT_OWNER_NAME` | `ownerEmployeeName` | 归属人姓名；后端先解析 employeeId。 |
| `CONTRACT_SUBMIT_NAME` | `submitterEmployeeName` | 申请人姓名；后端先解析 employeeId。 |
| `CONTRACT_DEPARTMENT_NAME` | `departmentName` | 部门名称；后端先解析部门 ID。 |
| `CONTRACT_TRADING_PARTY` | `contractTradingParty` | 交易方关键词。 |
| `CONTRACT_TRADING_PARTY_NAME_PRECISE` | `contractTradingPartyNamePrecise` | 交易方名称精确语义。 |
| `CONTRACT_LEGAL_ENTITY_NAME_PRECISE` | `contractLegalEntityNamePrecise` | 我方主体名称；后端先解析主体 ID。 |
| `CONTRACT_REMARK_FIELD` | `contractRemark` | 备注关键词。 |
| `CONTRACT_CAUSE_FIELD` | `contractCause` | 原因关键词。 |
| `CONTRACT_FORM_FIELD` | `dynamicForm` | 动态表单关键词。 |
| `CONTRACT_FORM_FIELDS_TEXT` | `contractFormFieldsText` | 自定义文本字段关键词。 |
| `ARCHIVE_NUMBER_FIELD` | `archiveNumber` | 归档编号关键词。 |
| `CONTRACT_TEXT_FIELD` | `contractText` | 合同正文文本。 |
| `CONTRACT_SCAN_TEXT` | `contractScan` | 扫描件文本。 |
| `CONTRACT_ARCHIVE_ATTACHMENT` | `contractArchiveAttachment` | 归档附件文本。 |
| `CONTRACT_ATTACHMENT_TEXT` | `contractAttachment` | 合同附件文本。 |

`CONTRACT_CURRENCY` 不在 `condition_units` 白名单，币种必须放在 `filter_units`。

顶层说明建议将稳定组合的合同编号放入 `filter_units.CONTRACT_NUMBER_FIELD + MUST`，但同一份工具描述的 filter 白名单和当前 CLM 策略都未列出该字段。确认服务端元数据前，使用 `condition_units.CONTRACT_NUMBER_FIELD + MUST`，不要依赖冲突写法。

## 结构化筛选 `filter_units`

单项结构与 `condition_units` 相同，但默认 `union_type=MUST`。自定义字段必须同时传 `filter_unique_key=attribute_key`。

固定字段：

| 枚举名 / PC 字段名 | `search_value` 类型与约束 |
| --- | --- |
| `CONTRACT_SUBMIT_ID` / `submitterEmployeeId` | 单个 employeeId。 |
| `DEPARTMENT_AUTHORITY` / `departmentAuthority` | 部门 OPENID 数组；不是部门名称搜索。 |
| `CONTRACT_DEMAND_PERSON` / `demandEmployeeId` | 单个 employeeId。 |
| `CONTRACT_DEMAND_PERSON_DEPARTMENT` / `demandDepartmentIds` | 需求部门 ID 数组。 |
| `CONTRACT_AMOUNT` / `contractAmount` | 二元数值范围 `[start,end]`，单边可用 `null`。 |
| `CONTRACT_CURRENCY` / `contractCurrency` | `"CNY"`、数字 code，或多值数组；无效币种返回空结果。 |
| `CONTRACT_LEGAL_ENTITY_NAME_PRECISE` / `contractLegalEntityNamePrecise` | 我方主体筛选值。 |
| `CONTRACT_TRADING_PARTY` / `contractTradingParty` | 交易方筛选值。 |
| `CONTRACT_ASSOCIATION_PARTY` / `contractAssociationParty` | 关联交易方筛选值。 |
| `CONTRACT_ANTI_DATED` / `antiDated` | 单值倒签标识。 |
| `CONTRACT_ANTI_DATED_MULTI` / `antiDatedMulti` | 倒签标识数组。 |
| `CONTRACT_SCAN_EXISTS` / `contractScanExists` | `0/1` 或 boolean。 |
| `CONTRACT_OWNER_EQUALS_SUBMITTER` / `ownerEqualsSubmitter` | `0/1` 或 boolean。 |
| `CONTRACT_SHARE_TO_ME` / `contractShareToMe` | 分享给我的筛选值。 |
| `CONTRACT_STANDARD_FRAMEWORK_AGREEMENT` / `contractFrameworkParentFlag` | `0/1` 或 boolean。 |
| `CONTRACT_SIGNED_DATE` / `signedDate` | 毫秒时间戳范围 `[start,end]`。 |
| `CONTRACT_FIRST_SIGNER` / `contractFirstSigner` | 首签方筛选值。 |
| `CONTRACT_SIGN_PLATFORM_TYPE` / `contractSignPlatformType` | 签署平台类型。 |
| `CONTRACT_SEAL_NUMBER` / `contractSealNumber` | 印章编号 string。 |
| `CONTRACT_TEMPLATE_MULTI_CONTRACT` / `contractTemplateMultiContract` | 模板多合同筛选值。 |
| `CONTRACT_EFFECTIVE_STATUS` / `contractEffectiveStatus` | 生效状态数组。 |
| `CONTRACT_NODE_OCR_USAGE` / `contractNodeOcrUsage` | OCR 使用状态数字枚举 code。 |
| `CONTRACT_OWNER_AUTHORIZED_EMPLOYEES` / `ownerAuthorizedEmployeeId` | employeeId 数组。 |
| `CONTRACT_PRE_AUTHORIZED_EMPLOYEES` / `preAuthorizedEmployeeId` | employeeId 数组。 |

自定义字段：

| 枚举名 / PC 字段名 | `search_value` | 额外约束 |
| --- | --- | --- |
| `CONTRACT_FORM_FIELDS_TEXT` / `contractFormFieldsText` | 文本值 | 必传 `filter_unique_key`。 |
| `CONTRACT_FORM_FIELDS_DOUBLE` / `contractFormFieldsDouble` | `[start,end]` | 单边可为 `null`；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_DATE` / `contractFormFieldsDate` | 毫秒时间戳范围 | 不传日期字符串；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_OPTION` / `contractFormFieldsOption` | 选项 label | 必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_OPTION_ID` / `contractFormFieldsOptionId` | 选项原始 value/id | 必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT` / `contractFormFieldsEmployeeDepartment` | 人员/部门显示名 | 必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT_ID` / `contractFormFieldsEmployeeDepartmentId` | 人员/部门 ID | 必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_CURRENCY` / `contractFormFieldsCurrency` | 币种值 | 必传唯一 key。 |
| `CONTRACT_HYPERLINK` / `contractHyperlink` | `{"url":"https://example.com","title":"标题"}` | 必传唯一 key。 |

## 排序与分页

`sort_type` 枚举：

| 值 | 含义 | 值 | 含义 |
| --- | --- | --- | --- |
| `SUBMITTED_TIME` | 申请时间 | `SIGNED_TIME` | 签订时间 |
| `CREATE_TIME` | 创建时间 | `ARCHIVED_TIME` | 归档时间 |
| `CONTRACT_NUMBER` | 合同编号 | `START_DATE` | 合同开始日期 |
| `CONTRACT_STATUS` | 合同状态 | `END_DATE` | 合同结束日期 |
| `AMOUNT` | 合同金额 | `ARCHIVE_NUMBER` | 归档编号 |

- `order` 只允许 `ASC`、`DESC`，默认 `DESC`。
- `sort` 是旧兼容字段；仅值 `asc` 表示升序，且排序字段固定为提交时间。
- 首次查询不传 `page_token`；后续直接使用响应 token，不自行计算。

## 枚举与约束

合同状态 code：

| code | 含义 | code | 含义 |
| --- | --- | --- | --- |
| `0` | 正编辑/草稿 | `9` | 已归档 |
| `1` | 已作废 | `10` | 变更中 |
| `2` | 已撤回 | `11` | 已变更 |
| `3` | 审批中 | `12` | 我方已签约，归入签订中 |
| `4` | 已拒绝 | `13` | 对方已签约，归入签订中 |
| `5` | 审批已通过 | `16` | 终止中 |
| `6` | 签订中 | `17` | 已终止 |
| `7` | 已签订 | `19` | 审批流程被干预中止，PC 按已作废展示 |
| `8` | 归档中 | `23` | 已完成 |

PC 聚合状态：已作废=`1,19`，签订中=`6,12,13`，其余使用对应单个 code。

其他硬约束：

- `condition_units` 和 `filter_units` 各最多 50 项。
- `demand_employee_ids` 当前只允许一个元素。
- 自定义 filter 字段必须传 `filter_unique_key`。
- `DEPARTMENT_AUTHORITY` 传部门 ID 数组；按部门名称搜索使用 `condition_units.CONTRACT_DEPARTMENT_NAME`。
- 日期型 filter 传毫秒时间戳范围，不传 `YYYY-MM-DD` 字符串。
- 不传 `search_tab_code` 时默认 `0`；非法页签、排序字段或排序方向返回参数错误。

## 动态字段发现

MCP 工具描述要求：按非固定字段、自定义字段、选项、人员或部门字段搜索前，优先按展示名调用 `list-contract-search-filter-fields`，再按返回的 `request_location`、`request_paths` 和 `filter_unique_key` 组织请求。

当前 `contract-cli` 尚未提供这个字段发现接口的结构化命令，且通用 `api call` 未开放。因此：

- 已知 `attribute_key` 时可按本页构造请求。
- 不知道字段技术名或唯一 key 时不要猜测，应先说明 CLI 能力缺口，等待补充字段发现命令或由用户提供元数据结果。

## 示例

按合同编号关键词搜索：

```bash
contract-cli contract search --profile contract --as user --data '{"contract_number":"CT2026","page_size":10}'
```

关键词与结构化条件组合，保存为 `search-user.json`：

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "sort_type": "SUBMITTED_TIME",
  "order": "DESC",
  "condition_units": [
    {
      "search_field": "CONTRACT_NAME_FIELD",
      "search_value": "采购",
      "union_type": "MUST"
    }
  ],
  "filter_units": [
    {
      "search_field": "CONTRACT_AMOUNT",
      "search_value": [1000, 5000],
      "union_type": "MUST"
    },
    {
      "search_field": "CONTRACT_CURRENCY",
      "search_value": ["CNY", "USD"],
      "union_type": "MUST"
    }
  ]
}
```

```bash
contract-cli contract search --profile contract --as user --input-file search-user.json
```

按自定义日期字段筛选：

```json
{
  "filter_units": [
    {
      "search_field": "CONTRACT_FORM_FIELDS_DATE",
      "filter_unique_key": "custom_sign_date",
      "search_value": [1704067200000, 1706659200000],
      "union_type": "MUST"
    }
  ]
}
```

## 响应摘要

- 常用路径：`data.items[]`、`data.has_more`、`data.page_token`、`data.pagination`。
- `pagination.unit` 固定为 `CONTRACT_GROUP`。
- `pagination` 包含 `page_index`、`requested_size`、`returned_units`、`returned_items`、`total_units`。
- MCP 列表项还会返回 `contract_status_display_name`、`contract_status_display_name_source`、`contract_status_display_policy_version`。
- 完整合同字段见 [contract-response-fields.md](contract-response-fields.md)。
