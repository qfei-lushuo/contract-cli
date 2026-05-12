# Contract Commands Reference

## 查询与创建

```bash
contract-cli contract get 7023646046559404327 --profile contract --as user
contract-cli contract get 7023646046559404327 --profile contract --as bot
contract-cli contract get 7023646046559404327 --profile contract --as bot --user-id ou_xxx --user-id-type employee_id
contract-cli contract search --profile contract --as user --input-file contract-search.json
contract-cli contract search --profile contract --as user --data '{"contract_number":"CN-001"}'
contract-cli contract search --profile contract --as bot --input-file contract-search.json
contract-cli contract search --profile contract --as bot --input-file contract-search.json --user-id ou_xxx --user-id-type employee_id
contract-cli contract create --profile contract --input-file contract-create.json
contract-cli contract create --profile contract --data '{"title":"示例合同"}'
contract-cli contract create --profile contract --as bot --data '{"contract_name":"示例合同","create_user_id":"ou_xxx"}'
contract-cli contract upload-file --profile contract --as user --file ./合同正文.docx --file-type text
contract-cli contract upload-file --profile contract --as bot --file ./附件.pdf --file-type attachment --file-name 附件.pdf
contract-cli contract submit 7023646046559404327 --profile contract --as bot
contract-cli contract resubmit 7023646046559404327 --profile contract --as bot
contract-cli contract patch 7023646046559404327 --profile contract --as bot --input-file contract-patch.json
contract-cli contract download-file file_123 --profile contract --as bot --output-file ./contract.pdf
contract-cli contract delete 7023646046559404327 --profile contract --as bot
contract-cli contract print-file --profile contract --as bot --input-file print-file.json
contract-cli contract share get 7023646046559404327 --profile contract --as bot
contract-cli contract cooperation link get 7023646046559404327 --profile contract --as bot
contract-cli contract cooperation record get 7023646046559404327 --profile contract --as bot
contract-cli contract category list --profile contract --as bot --lang zh-CN
contract-cli contract template list --profile contract --as bot --category-number CAT-1 --page-size 20 --user-id ou_xxx --user-id-type employee_id
contract-cli contract template get tpl_123 --profile contract --as bot --user-id ou_xxx --user-id-type employee_id
contract-cli contract template instantiate --profile contract --as bot --data '{"template_number":"TMP001","create_user_id":"ou_xxx"}' --user-id-type employee_id
contract-cli contract sync-user-groups --profile contract --as user
contract-cli contract sync-user-groups --profile contract --as bot
contract-cli contract sync-user-groups --profile contract --as bot --user-id ou_xxx
contract-cli contract text 7023646046559404327 --profile contract --as user --full-text
contract-cli contract text 7023646046559404327 --profile contract --as bot --full-text
contract-cli contract text 7023646046559404327 --profile contract --as bot --full-text --user-id-type employee_id
```

## 分类、模板、枚举

```bash
contract-cli contract category list --profile contract --lang zh-CN
contract-cli contract template list --profile contract --category-number CAT-1 --page-size 20
contract-cli contract template get tpl_123 --profile contract
contract-cli contract template instantiate --profile contract --input-file template-instance.json
contract-cli contract enum list --profile contract --type contract_status
```

## 已知限制

- `contract upload-file` 当前同时支持 user/bot 身份，均走 `/open-apis/contract/v1/files/upload`
- `contract submit`、`contract resubmit`、`contract patch`、`contract download-file`、`contract delete`、`contract print-file`、`contract share get`、`contract cooperation link get`、`contract cooperation record get` 当前仅支持 bot 身份
- `contract template fields` 尚未实现
- `contract create` 不自动帮你补模板信息；当前就是透传请求体
- `contract create --as bot` 时，`create_user_id` 需要你自己写进 JSON body
- `contract template list --as bot` 按生产文档通常需要 `category_number`、`user_id`、`user_id_type`，但 CLI 目前只负责透传，不做本地必填校验
- `contract template get --as bot` 按生产文档通常需要 `user_id`、`user_id_type`，但 CLI 目前只负责透传，不做本地必填校验
- `contract template instantiate --as bot` 按生产文档会用到 query `user_id_type` 和 body `create_user_id`，但 CLI 目前只负责透传，不做本地必填校验

## bot-only 合同操作

```bash
# 提交/重新提交：请求体可选
contract-cli contract submit 7023646046559404327 --profile contract --as bot
contract-cli contract resubmit 7023646046559404327 --profile contract --as bot --data '{"comment":"修正后重新提交"}'

# 更新合同：请求体必填
contract-cli contract patch 7023646046559404327 --profile contract --as bot --input-file contract-patch.json

# 删除草稿合同：直接删除，不额外要求 --yes
contract-cli contract delete 7023646046559404327 --profile contract --as bot
```

接口路径：

- `submit`：`POST /open-apis/contract/v1/contracts/{contract_id}/submit`
- `resubmit`：`POST /open-apis/contract/v1/contracts/{contract_id}/resubmit`
- `patch`：`PATCH /open-apis/contract/v1/contracts/{contract_id}`
- `delete`：`DELETE /open-apis/contract/v1/contracts/{contract_id}`

## bot-only 文件命令

```bash
# 默认拉起保存弹窗；Agent/CI/远程环境建议显式传 --output-file
contract-cli contract download-file file_123 --profile contract --as bot --output-file ./contract.pdf

# 管道场景使用 --raw
contract-cli contract download-file file_123 --profile contract --as bot --raw > contract.pdf

# 生成合同打印文件，请求体必填
contract-cli contract print-file --profile contract --as bot --input-file print-file.json
```

接口路径：

- `download-file`：`GET /open-apis/contract/v1/files/{file_id}`
- `print-file`：`POST /open-apis/contract/v1/files`

注意：

- 正式命令是 `download-file`，不支持 `dowload-file` 拼写。
- `download-file --output-file` 遇到已存在文件会失败；需要覆盖时加 `--force`。

## bot-only 分享与协商查询

```bash
contract-cli contract share get 7023646046559404327 --profile contract --as bot
contract-cli contract cooperation link get 7023646046559404327 --profile contract --as bot
contract-cli contract cooperation record get 7023646046559404327 --profile contract --as bot
```

接口路径：

- `share get`：`GET /open-apis/contract/v1/contracts/{contract_id}/share_records`
- `cooperation link get`：`GET /open-apis/contract/v1/contracts/{contract_id}/cooperation_link`
- `cooperation record get`：`GET /open-apis/contract/v1/contracts/{contract_id}/cooperation_record_info`
