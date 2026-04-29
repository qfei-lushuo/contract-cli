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
contract-cli contract upload-file --profile contract --as bot --file ./合同正文.docx --file-type text
contract-cli contract upload-file --profile contract --as bot --file ./附件.pdf --file-type attachment --file-name 附件.pdf
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

- `contract upload-file` 当前仅支持 bot 身份，不支持 user/MCP 三段式上传
- `contract template fields` 尚未实现
- `contract create` 不自动帮你补模板信息；当前就是透传请求体
- `contract create --as bot` 时，`create_user_id` 需要你自己写进 JSON body
- `contract template list --as bot` 按生产文档通常需要 `category_number`、`user_id`、`user_id_type`，但 CLI 目前只负责透传，不做本地必填校验
- `contract template get --as bot` 按生产文档通常需要 `user_id`、`user_id_type`，但 CLI 目前只负责透传，不做本地必填校验
- `contract template instantiate --as bot` 按生产文档会用到 query `user_id_type` 和 body `create_user_id`，但 CLI 目前只负责透传，不做本地必填校验
