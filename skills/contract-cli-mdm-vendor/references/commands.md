# Vendor Commands Reference

```bash
contract-cli mdm vendor list --profile contract --name "供应商A"
contract-cli mdm vendor list --profile contract --name "供应商A" --page-size 20 --page-token next
contract-cli mdm vendor list --profile contract --as app --name "V00000001" --page-size 20 --user-id-type employee_id
contract-cli mdm vendor get 1063197165850985296 --profile contract
contract-cli mdm vendor get 7003410079584092448 --profile contract --as app --user-id-type employee_id
contract-cli mdm vendor create --profile contract --as app --user-id <operator-user-id> --input-file vendor-create.json
contract-cli mdm vendor update 7003410079584092448 --profile contract --as app --user-id <operator-user-id> --input-file vendor-update.json
contract-cli mdm vendor list-all --profile contract --as app --page-size 20
contract-cli mdm vendor query-by-cert --profile contract --as app --certification-id 91110105 --ad-country CN
```

说明：

- `mdm vendor list` 会按身份路由：
  - `user` -> `contract/v1/mcp/vendors`
  - `app` -> `mdm/v1/vendors`
- `mdm vendor get` 也会按身份路由：
  - `user` -> `contract/v1/mcp/vendors/{vendor_id}`
  - `app` -> `mdm/v1/vendors/{vendor_id}`
- 默认输出可直接用于查候选方或确认详情
- 需要查看原始响应时可加 `--raw`
- `create/update/list-all/query-by-cert` 当前仅支持 app 身份
- `create/update` 必须传 `--user-id`
- `create` 请求体不要带后端生成的 `vendor` 编码；`update` 请求体必须带后端返回的 `id` 和 `vendor` 编码
- 创建/更新前建议先查 `contract-cli mdm fields list --profile contract --as app --biz-line vendor`
