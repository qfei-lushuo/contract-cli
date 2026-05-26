package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContractSkillFieldReferencesCoverDocumentedCommands(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	skillPath := filepath.Join(root, "skills", "contract-cli-contract", "SKILL.md")
	skillContent := readTextFile(t, skillPath)

	referenceFragments := map[string][]string{
		"search-contract-fields.md": {
			"contract-cli contract search",
			"combine_condition",
			"logic_search",
			"contract_status_in",
		},
		"contract-response-fields.md": {
			"contract-cli contract get",
			"contract_id",
			"counter_party_list",
			"payment_plan_list",
		},
		"patch-contract-fields.md": {
			"contract-cli contract patch",
			"ocr_file_id",
			"archive_attachment_map",
			"archive_attachment_file_ids",
		},
		"template-fields.md": {
			"contract-cli contract template list",
			"contract-cli contract template get",
			"template_fields",
			"value_scopes",
		},
		"template-instance-fields.md": {
			"contract-cli contract template instantiate",
			"template_field_list",
			"field_value",
			"create_employee_code",
		},
		"print-file-fields.md": {
			"contract-cli contract print-file",
			"operate_type",
			"file_id",
		},
		"category-fields.md": {
			"contract-cli contract category list",
			"category_resources",
			"abbreviation",
		},
		"share-cooperation-fields.md": {
			"contract-cli contract share get",
			"contract-cli contract cooperation record get",
			"cooperation_record_infos",
		},
		"contract-actions-fields.md": {
			"contract-cli contract upload-file",
			"contract-cli contract submit",
			"contract-cli contract download-file",
			"process_instance_id",
		},
	}

	for name, fragments := range referenceFragments {
		name := name
		fragments := fragments
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(skillContent, "references/"+name) {
				t.Fatalf("contract skill must link references/%s", name)
			}
			content := readTextFile(t, filepath.Join(root, "skills", "contract-cli-contract", "references", name))
			for _, fragment := range fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", name, fragment)
				}
			}
		})
	}
}

func TestContractSkillCommandsDoNotSuggestPatchTitleShortcut(t *testing.T) {
	t.Parallel()

	content := readTextFile(t, filepath.Join("..", "..", "skills", "contract-cli-contract", "references", "commands.md"))
	if strings.Contains(content, `contract patch 7023646046559404327 --profile contract --as bot --data '{"title":"demo"}'`) {
		t.Fatalf("contract patch commands should not suggest title-only patch payload")
	}
}

func readTextFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}
