//go:build windows

package selfupdate

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func packageManagerCommand(ctx context.Context, path string, args []string) (*exec.Cmd, error) {
	extension := strings.ToLower(filepath.Ext(path))
	if extension != ".cmd" && extension != ".bat" {
		return exec.CommandContext(ctx, path, args...), nil
	}
	// cmd does not use Go's normal Windows argument escaping. Quote every word,
	// disable AutoRun/delayed expansion, and reject percent expansion and quotes.
	// Fail safely for those unusual paths instead of interpreting them as code.
	words := append([]string{path}, args...)
	for index, word := range words {
		if strings.ContainsAny(word, "\"%\r\n\x00") {
			return nil, fmt.Errorf("cannot safely invoke Windows package manager argument %q", word)
		}
		words[index] = "\"" + word + "\""
	}
	interpreter, err := exec.LookPath("cmd.exe")
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, interpreter)
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: syscall.EscapeArg(interpreter) + " /d /v:off /s /c \"" + strings.Join(words, " ") + "\"",
	}
	return command, nil
}
