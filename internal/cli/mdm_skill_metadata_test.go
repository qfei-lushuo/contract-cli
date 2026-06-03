package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMDMSkillAgentMetadataMatchesAppIdentityRoutes(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills")
	tests := []struct {
		name     string
		file     string
		required []string
	}{
		{
			name: "vendor",
			file: filepath.Join(root, "contract-cli-mdm-vendor", "agents", "openai.yaml"),
			required: []string{
				"user or app",
				"vendor candidates",
				"vendor details",
			},
		},
		{
			name: "legal",
			file: filepath.Join(root, "contract-cli-mdm-legal", "agents", "openai.yaml"),
			required: []string{
				"user or app",
				"legal entity candidates",
				"legal entity details",
			},
		},
		{
			name: "fields",
			file: filepath.Join(root, "contract-cli-mdm-fields", "agents", "openai.yaml"),
			required: []string{
				"user or app",
				"vendor",
				"legal_entity",
				"vendor_risk",
				"user/MCP",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content := readMDMSkillMetadataText(t, tt.file)
			for _, forbidden := range []string{"user-only", "user-authorized"} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("%s must not keep obsolete identity wording %q: %s", tt.file, forbidden, content)
				}
			}
			for _, required := range tt.required {
				if !strings.Contains(content, required) {
					t.Fatalf("%s missing identity wording %q: %s", tt.file, required, content)
				}
			}
		})
	}
}

func TestMDMSkillReferencesDoNotKeepObsoleteUserOnlyClaims(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("..", "..", "skills", "contract-cli-mdm-vendor", "references", "vendor-query-parameters.md"),
		filepath.Join("..", "..", "skills", "contract-cli-mdm-legal", "references", "entity-query-parameters.md"),
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()

			content := readMDMSkillMetadataText(t, file)
			for _, forbidden := range []string{
				"mdm legal` 和 `mdm fields` 仍保持 user-only",
				"mdm fields` 仍保持 user-only",
			} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("%s must not keep obsolete identity wording %q: %s", file, forbidden, content)
				}
			}
		})
	}
}

func readMDMSkillMetadataText(t *testing.T, file string) string {
	t.Helper()

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", file, err)
	}
	return string(content)
}
