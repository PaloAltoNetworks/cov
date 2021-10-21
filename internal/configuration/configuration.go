// Package configuration is a small package for handling configuration
package configuration

import (
	"fmt"

	"go.aporeto.io/addedeffect/lombric"
	"go.aporeto.io/underwater/logutils"
	"go.uber.org/zap"
)

// LoggingConf is the configuration for log.
type LoggingConf struct {
	LogFormat string `mapstructure:"log-format"   desc:"Log format"   default:"console"`
	LogLevel  string `mapstructure:"log-level"    desc:"Log level"    default:"info"`
}

// Configuration hold the service configuration.
type Configuration struct {
	LoggingConf `mapstructure:",squash"`

	BaseBranch        string   `mapstructure:"patch" desc:"The branch to use to check the patch coverage against. Example: master"`
	CoverageFilePaths []string `mapstructure:"coverage" desc:"The coverage files to use." required:"true"`
	CoverageThreshold int      `mapstructure:"target" desc:"The target of coverage in percent that is requested"`
	Filters           []string `mapstructure:"filter" desc:"The filters to use for coverage lookup"`
	Name              string   `mapstructure:"name" desc:"Meaning full name to use for output" default:"Project"`
}

// Prefix returns the configuration prefix.
func (c *Configuration) Prefix() string { return "cov" }

// PrintVersion prints the current version.
func (c *Configuration) PrintVersion() {
	fmt.Printf("cov - %s (%s)\n", ProjectVersion, ProjectSha)
}

// NewConfiguration returns a new configuration.
func NewConfiguration() *Configuration {

	c := &Configuration{}
	lombric.Initialize(c)
	logutils.Configure(c.LogLevel, c.LogFormat)

	if len(c.CoverageFilePaths) == 0 {
		zap.L().Fatal("No coverage files provided")

	}

	return c
}
