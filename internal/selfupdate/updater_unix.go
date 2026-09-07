//go:build !windows

package selfupdate

func (u *Updater) PrepareSelfReplace() (func(), error) {
	return func() {}, nil
}

func (u *Updater) CleanupStaleFiles() {}

func (u *Updater) CanRestorePreviousVersion() bool { return false }
