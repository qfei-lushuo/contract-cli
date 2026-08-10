package cli_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestContractSearchSkillSeparatesUserAndAppContracts(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract")
	skillContent := readTextFile(t, filepath.Join(root, "SKILL.md"))

	references := []struct {
		file      string
		fragments []string
	}{
		{
			file: "search-contract-fields.md",
			fragments: []string{
				"contract search --as user",
				"contract search --as app",
				"contract search-v2 --as app",
				"search-user-parameters.md",
				"search-app-parameters.md",
				"search-v2-parameters.md",
			},
		},
		{
			file: "search-user-parameters.md",
			fragments: []string{
				"contract-cli contract search --profile contract --as user",
				"POST /open-apis/contract/v1/mcp/contracts/search",
				"## CLI 参数映射",
				"## 请求体字段",
				"condition_units",
				"filter_units",
				"search_tab_code",
				"sort_type",
				"user_id_type",
				"filter_unique_key",
				"CONTRACT_GROUP",
				"## 枚举与约束",
				"## 示例",
			},
		},
		{
			file: "search-app-parameters.md",
			fragments: []string{
				"contract-cli contract search --profile contract --as app",
				"POST /open-apis/contract/v1/contracts/search",
				"## CLI 参数映射",
				"## 请求体字段",
				"精确查询",
				"110107",
				"combine_condition",
				"logic_search",
				"## 枚举与约束",
				"## 示例",
			},
		},
		{
			file: "search-v2-parameters.md",
			fragments: []string{
				"contract-cli contract search-v2 --profile contract --as app",
				"POST /open-apis/contract/v1/contracts/searchV2",
				"ES 模糊查询",
				"page_size",
				"10000",
				"condition_units",
				"filter_units",
				"忽略",
			},
		},
	}

	for _, reference := range references {
		reference := reference
		t.Run(reference.file, func(t *testing.T) {
			t.Parallel()

			link := "references/" + reference.file
			if !strings.Contains(skillContent, link) {
				t.Fatalf("contract skill must link %s", link)
			}
			content := readTextFile(t, filepath.Join(root, "references", reference.file))
			for _, fragment := range reference.fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", reference.file, fragment)
				}
			}
		})
	}

	if strings.Contains(skillContent, "想用新版搜索条件：用 `contract search-v2 --as app`") {
		t.Fatalf("contract skill must not describe search-v2 as the generic newer search contract")
	}
}

func TestContractSearchReferencesUseRunnableIdentitySpecificExamples(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract", "references")
	checks := []struct {
		file      string
		required  []string
		forbidden []string
	}{
		{
			file: "search-user-parameters.md",
			required: []string{
				"--as user --input-file",
				"--as user --data",
			},
		},
		{
			file: "search-app-parameters.md",
			required: []string{
				"--as app --input-file",
				"--as app --data",
			},
		},
		{
			file: "search-v2-parameters.md",
			required: []string{
				"--as app --input-file",
			},
			forbidden: []string{
				`"page_size": 0`,
			},
		},
	}

	for _, check := range checks {
		content := readTextFile(t, filepath.Join(root, check.file))
		for _, fragment := range check.required {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing runnable example fragment %q", check.file, fragment)
			}
		}
		for _, fragment := range check.forbidden {
			if strings.Contains(content, fragment) {
				t.Errorf("%s contains invalid example fragment %q", check.file, fragment)
			}
		}
	}
}

func TestContractSearchUserFilterValueContractsMatchCurrentCLIProfile(t *testing.T) {
	t.Parallel()

	content := readTextFile(t, filepath.Join(
		"..", "..", "skills", "contract-cli-contract", "references", "search-user-parameters.md",
	))

	required := []string{
		"当前 `contract-cli` 不发送 `X-MCP-Response-Profile`",
		"`CONTRACT_SUBMIT_ID` / `submitterEmployeeId` | `string` / `array<string>` | 飞书 `user_id`",
		"`CONTRACT_AMOUNT` / `contractAmount` | `array` | 恰好两个元素 `[start,end]`",
		"`CONTRACT_CURRENCY` / `contractCurrency` | `string` / `integer` / `array`",
		"`CONTRACT_SEAL_NUMBER` / `contractSealNumber` | `integer` / `array<integer>`",
		"`CONTRACT_FORM_FIELDS_OPTION` / `contractFormFieldsOption` | `string` / `array<string>`",
		"`CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT_ID` / `contractFormFieldsEmployeeDepartmentId` | `string` / `array<string>`",
		"JSON integer `0` 或 `1`",
		"不传 JSON boolean",
		"盖章份数",
	}
	for _, fragment := range required {
		if !strings.Contains(content, fragment) {
			t.Errorf("search-user-parameters.md missing filter value contract %q", fragment)
		}
	}

	forbidden := []string{
		"`0/1` 或 boolean",
		"印章编号 string",
		"单个 employeeId",
	}
	for _, fragment := range forbidden {
		if strings.Contains(content, fragment) {
			t.Errorf("search-user-parameters.md contains obsolete filter value contract %q", fragment)
		}
	}
}
