//go:build windows

package selfupdate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsPackageManagerBatchEntry(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "包管理器 & tools")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{".cmd", ".bat"} {
		t.Run(extension, func(t *testing.T) {
			path := filepath.Join(directory, "npm"+extension)
			if err := os.WriteFile(path, []byte("@echo off\r\necho [%~1][%~2][%~3]\r\n"), 0600); err != nil {
				t.Fatal(err)
			}
			command, err := packageManagerCommand(context.Background(), path, []string{"install", "-g", "@qfeius/contract-cli@1.8.4"})
			if err != nil {
				t.Fatal(err)
			}
			output, err := command.CombinedOutput()
			if err != nil || strings.TrimSpace(string(output)) != "[install][-g][@qfeius/contract-cli@1.8.4]" {
				t.Fatalf("output=%q error=%v", output, err)
			}
		})
	}
}

func TestWindowsBatchRejectsExpansion(t *testing.T) {
	for _, value := range []string{"%PATH%", "a\"&echo injected", "a\nb"} {
		if _, err := packageManagerCommand(context.Background(), `C:\tools\npm.cmd`, []string{value}); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}
