package cli_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIGapSkillsCoverNewCommands(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	testCases := []struct {
		path      string
		fragments []string
	}{
		{
			path: filepath.Join(root, "skills", "contract-cli-contract", "references", "openapi-gap-commands.md"),
			fragments: []string{
				"contract-cli contract search-v2",
				"contract-cli contract field update",
				"contract-cli contract cooperation file download",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-mdm-vendor", "SKILL.md"),
			fragments: []string{
				"contract-cli mdm vendor create",
				"contract-cli mdm vendor update <vendor-id>",
				"mdm vendor query-by-cert",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-mdm-legal", "SKILL.md"),
			fragments: []string{
				"contract-cli mdm legal create",
				"contract-cli mdm legal update <legal-entity-id>",
				"contract-cli mdm legal get --code",
				"mdm fields list --biz-line legal_entity",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-mdm-exchange", "SKILL.md"),
			fragments: []string{
				"contract-cli mdm fixed-exchange-rate get",
				"contract-cli mdm fixed-exchange-rate update",
				"/open-apis/mdm/v1/fixed_exchange_rate",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-mdm-file", "SKILL.md"),
			fragments: []string{
				"contract-cli mdm file download",
				"/open-apis/mdm/v1/file/download/{file_id}",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-event", "SKILL.md"),
			fragments: []string{
				"contract-cli event outbound-ip list",
				"/open-apis/event/v1/outbound_ip",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-rule", "SKILL.md"),
			fragments: []string{
				"contract-cli rule table list",
				"contract-cli rule table row create",
				"/open-apis/rule_engine/v1",
			},
		},
		{
			path: filepath.Join(root, "skills", "contract-cli-shared", "SKILL.md"),
			fragments: []string{
				"contract-cli-mdm-exchange",
				"contract-cli-event",
				"contract-cli-rule",
				"mdm legal list/get/create/update",
				"rule table *",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			content := readTextFile(t, tc.path)
			for _, fragment := range tc.fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", tc.path, fragment)
				}
			}
		})
	}
}

func TestMDMSkillsNoLongerMarkCreateUpdateAsMissing(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	for _, path := range []string{
		filepath.Join(root, "skills", "contract-cli-mdm-vendor", "SKILL.md"),
		filepath.Join(root, "skills", "contract-cli-mdm-legal", "SKILL.md"),
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			content := readTextFile(t, path)
			if strings.Contains(content, "create/update` 当成已有结构化命令") ||
				strings.Contains(content, "结构化命令未实现") {
				t.Fatalf("%s still says create/update is missing", path)
			}
		})
	}
}
