package cli

import (
	"cn.qfei/contract-cli/internal/config"
	"fmt"
)

func (a *App) environmentProfileName() string {
	if a.buildEnvironment == "" || a.buildEnvironment == "prod" {
		return defaultProfileName
	}
	return "contract-" + a.buildEnvironment
}

func (a *App) environmentMismatch() error {
	return fmt.Errorf("this package only supports %s; install the corresponding environment package instead of switching environments", a.buildEnvironment)
}

func (a *App) validateBuildProfile(profile config.Profile) error {
	if a.buildEnvironment != "" && profile.Environment != a.buildEnvironment {
		return a.environmentMismatch()
	}
	return nil
}
