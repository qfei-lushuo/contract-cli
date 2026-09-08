package build

import "fmt"

const Name = "contract-cli"

var (
	// Environment is selected by the build, never by a runtime environment variable.
	Environment = "prod"
	Version     = "dev"
	Commit      = "unknown"
	Date        = "unknown"
)

type Info struct {
	Name        string
	Environment string
	Version     string
	Commit      string
	Date        string
}

func Current() Info {
	return Info{
		Name:        Name,
		Environment: Environment,
		Version:     defaultString(Version, "dev"),
		Commit:      defaultString(Commit, "unknown"),
		Date:        defaultString(Date, "unknown"),
	}
}

func (i Info) String() string {
	return fmt.Sprintf("%s version %s (commit %s, built %s, environment %s)", i.Name, defaultString(i.Version, "dev"), defaultString(i.Commit, "unknown"), defaultString(i.Date, "unknown"), defaultString(i.Environment, "prod"))
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
