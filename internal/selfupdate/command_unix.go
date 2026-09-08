//go:build !windows

package selfupdate

import (
	"context"
	"os/exec"
)

func packageManagerCommand(ctx context.Context, path string, args []string) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, path, args...), nil
}
