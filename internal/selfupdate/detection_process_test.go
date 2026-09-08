package selfupdate

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectInstallMethodUsesManagerGlobalRoot(t *testing.T) {
	directory := t.TempDir()
	root := filepath.Join(directory, "global", "node_modules")
	toolsDir := filepath.Join(directory, "tools")
	if err := os.MkdirAll(toolsDir, 0700); err != nil {
		t.Fatal(err)
	}
	manager := filepath.Join(toolsDir, "npm")
	script := "#!/bin/sh\nif [ \"$1\" != root ] || [ \"$2\" != -g ]; then exit 9; fi\nprintf '%s\\n' \"$CONTRACT_TEST_GLOBAL_ROOT\"\n"
	binaryName := "contract-cli"
	if runtime.GOOS == "windows" {
		manager += ".cmd"
		binaryName += ".exe"
		script = "@echo off\r\nif not \"%~1\"==\"root\" exit /b 9\r\nif not \"%~2\"==\"-g\" exit /b 9\r\necho %CONTRACT_TEST_GLOBAL_ROOT%\r\n"
	}
	if err := os.WriteFile(manager, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"global", "project", "_npx/cache"} {
		t.Run(kind, func(t *testing.T) {
			binary := filepath.Join(directory, kind, "node_modules", "@qfeius", "contract-cli", "bin", binaryName)
			if err := os.MkdirAll(filepath.Dir(binary), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(binary, content, 0700); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary, "-test.run=^TestDetectInstallProcessHelper$")
			command.Env = append(os.Environ(), "PATH="+toolsDir+string(os.PathListSeparator)+os.Getenv("PATH"), "CONTRACT_TEST_GLOBAL_ROOT="+root, "CONTRACT_TEST_DETECT_HELPER=1")
			output, err := command.Output()
			if err != nil {
				t.Fatal(err)
			}
			var result DetectResult
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("%v: %s", err, output)
			}
			if result.CanAutoUpdate() != (kind == "global") {
				t.Fatalf("%s: %+v", kind, result)
			}
		})
	}
}

func TestDetectInstallProcessHelper(t *testing.T) {
	if os.Getenv("CONTRACT_TEST_DETECT_HELPER") != "1" {
		return
	}
	if err := json.NewEncoder(os.Stdout).Encode(detectInstallMethod()); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
