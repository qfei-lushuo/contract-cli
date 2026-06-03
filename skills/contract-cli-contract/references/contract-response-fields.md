# Contract Response Fields

这份文档用于读取 `contract-cli contract get` 和 `contract-cli contract search` 的合同响应字段。

官方来源：

- [查看合同详情](https://docs.qfei.cn/366437223e0)
- [搜索合同](https://docs.qfei.cn/366415247e0)

## 1. CLI 命令面

```bash
contract-cli contract get <contract-id> --profile contract --as app --output json
contract-cli contract search --profile contract --as app --input-file contract-search.json --output json
```

读取路径：

- `contract get`：重点看 `data.contract`
- `contract search`：重点看 `data.items[]`

## 2. 合同基础字段

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `contract_id` | `string|number` | 合同 id。用于 `get/submit/patch/delete/print-file/share/cooperation` 等命令。 |
| `contract_number` | `string` | 合同编号。 |
| `contract_name` | `string` | 合同名称。 |
| `contract_version` | `integer` | 合同版本。 |
| `contract_status_code` | `integer` | 合同状态编码。 |
| `contract_status_name` | `string` | 合同状态名称。 |
| `business_type_code` | `integer` | 合同业务类型编码，常见 `0` 合同申请、`2` 合同变更、`3` 合同终止。 |
| `business_type_name` | `string` | 合同业务类型名称。 |
| `contract_category_number` | `string` | 合同分类编码。 |
| `contract_category_name` | `string` | 合同分类名称。 |
| `contract_category_abbreviation` | `string` | 合同分类缩写。 |
| `parent_contract_category_number` | `string` | 父级分类编码。 |
| `parent_contract_category_name` | `string` | 父级分类名称。 |
| `amount` | `number|string` | 合同金额。无金额合同时可能不返回。 |
| `estimated_amount` | `number` | 合同预估金额。 |
| `currency_code` | `string` | 币种国际编码，如 `CNY`、`USD`。 |
| `currency_name` | `string` | 币种名称。 |
| `pay_type_code` | `integer` | 收支类型编码。 |
| `pay_type_name` | `string` | 收支类型名称。 |
| `property_type_code` | `integer` | 计价方式编码。 |
| `property_type_name` | `string` | 计价方式名称。 |
| `source_id` | `string` | 外部来源 id。 |
| `source_type_code` | `integer` | 来源类型编码。 |
| `remark` | `string` | 备注。 |

## 3. 人员与时间字段

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `create_user_id` | `string` | 创建人 id。 |
| `create_user_name` | `string` | 创建人名称。 |
| `create_employee_id` | `number` | 创建员工 id。 |
| `create_employee_name` | `string|null` | 创建员工名称。 |
| `owner_employee_id` | `number` | 归属员工 id。 |
| `department_id` | `string|null` | 合同创建部门 id。 |
| `department_name` | `string|null` | 合同创建部门名称。 |
| `create_time` | `string` | 创建时间。 |
| `update_time` | `string` | 更新时间。 |
| `submitted_time` | `string` | 提交时间。 |
| `approved_time` | `string` | 审批完成时间。 |
| `signed_time` | `string` | 签订完成时间。 |
| `archived_time` | `string` | 归档完成时间。 |
| `end_date` | `string` | 合同终止日期。 |

## 4. 主体字段

`counter_party_list` 表示对方主体：

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `counter_party_id` | `string` | 对方主体 id。 |
| `counter_party_code` | `string` | 对方主体 code。 |
| `counter_party_name` | `string` | 对方主体名称。 |
| `counter_party_sign_info_resource` | `object` | 对方签约信息，详情命令中可能返回。 |
| `bank_accounts` | `array<object>` | 银行账户信息，详情命令中可能返回。 |

`our_party_list` 表示我方主体：

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `our_party_id` | `string` | 我方主体 id。 |
| `our_party_code` | `string` | 我方主体 code。 |
| `our_party_name` | `string` | 我方主体名称。 |

`legal_entity_list` 表示法人实体：

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `legal_entity_id` | `string` | 法人实体 id。 |
| `legal_entity_code` | `string` | 法人实体 code。 |
| `legal_entity_name` | `string` | 法人实体名称。 |

## 5. 表单、模板与关联字段

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `form_id` | `number` | 合同表单 id。 |
| `form` | `string` | 合同表单 JSON 字符串，字段控件内容。 |
| `template_id` | `number|string` | 合同模板 id。 |
| `group_id` | `number` | 合同分组 id。 |
| `relation.relation_contracts[]` | `array<object>` | 关联的合同。 |
| `relation.related_contract_ids[]` | `array<string>` | 被关联的合同 id 列表。 |

## 6. 付款与预算字段

`payment_plan_list[]` 常用字段：

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `payment_plan_id` | `string` | 付款计划 id。 |
| `source_id` | `string` | 外部导入的付款计划 id。 |
| `payment_amount` | `number|null` | 付款金额。 |
| `currency_code` | `string` | 付款币种。 |
| `payment_date` | `string|null` | 付款日期。 |
| `payment_desc` | `string` | 付款说明。 |
| `payment_condition` | `string` | 付款条件。 |
| `payment_rate` | `integer|null` | 付款比例。 |
| `need_check` | `boolean` | 是否需要验收。 |
| `prepaid` | `boolean` | 是否预付。 |

`contract_budget_list[]` 常用字段：

| 字段 | 类型 | 业务含义 |
| --- | --- | --- |
| `contract_budget_id` | `string` | 预算 id。 |
| `budget_year` | `string` | 预算年度。 |
| `budget_department_id` | `string` | 预算部门 id。 |
| `budget_subject_code` | `string` | 预算科目编号。 |
| `cost_center_code` | `string` | 成本中心编码。 |
| `budget_code` | `string` | 预算编码。 |
| `budget_occupied_amount` | `number` | 预算占用金额，不含税。 |
| `tax_rate` | `number` | 税率。 |
| `tax_amount` | `number` | 税额。 |

## 7. 使用建议

- 后续要提交、删除、打印、分享或查协商信息时，优先保存 `contract_id`
- 后续要按模板创建合同，优先保存 `template_id` / `template_number`
- 后续要选择主体，优先保存 `our_party_id`、`our_party_code`、`counter_party_id`、`counter_party_code`
- `form` 是 JSON 字符串，不是已经展开的对象；需要解析后再读取控件内容
