//go:build windows

package selfupdate

import (
	"errors"
	"fmt"
	"os"
)

// PrepareSelfReplace renames the running executable so npm can replace it on
// Windows, where an executing file cannot be overwritten in place.
func (u *Updater) PrepareSelfReplace() (func(), error) {
	noop := func() {}
	executable, err := u.resolveExecutable()
	if err != nil {
		return noop, nil
	}
	backup := executable + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(executable, backup); err != nil {
		return noop, fmt.Errorf("cannot rename binary for update: %w", err)
	}
	u.backupCreated = true
	return func() {
		if _, err := os.Stat(backup); err != nil {
			u.backupCreated = false
			return
		}
		_ = os.Remove(executable)
		if err := os.Rename(backup, executable); err != nil {
			u.backupCreated = false
		}
	}, nil
}

func (u *Updater) CleanupStaleFiles() {
	executable, err := u.resolveExecutable()
	if err != nil {
		return
	}
	backup := executable + ".old"
	if _, err := os.Stat(backup); err != nil {
		return
	}
	if _, err := os.Stat(executable); errors.Is(err, os.ErrNotExist) {
		_ = os.Rename(backup, executable)
		return
	}
	_ = os.Remove(backup)
}

func (u *Updater) CanRestorePreviousVersion() bool { return u.backupCreated }
