package main

import (
	"os"

	"github.com/aporeto-inc/cov/internal/configuration"
	"github.com/aporeto-inc/cov/internal/coverage"
	"github.com/aporeto-inc/cov/internal/git"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"golang.org/x/tools/cover"
)

// Start starts the service
func main() {

	cfg := configuration.NewConfiguration()

	zap.L().Info("Coverage analysis", zap.Reflect("configuration", cfg))

	// parse profiles
	profiles := []*cover.Profile{}

	for _, coveragePath := range cfg.CoverageFilePaths {
		profs, err := cover.ParseProfiles(coveragePath)
		if err != nil {
			zap.L().Fatal("Unable to parse coverage profile", zap.Error(err))
		}
		profiles = append(profiles, profs...)
	}

	files := cfg.Filters

	if cfg.BaseBranch != "" {
		gitFiles, err := git.GetDiffFiles(cfg.BaseBranch)
		if err != nil {
			zap.L().Fatal("Unable to get change files", zap.String("branch", cfg.BaseBranch), zap.Error(err))
		}
		files = append(files, gitFiles...)
	}

	tree := coverage.NewTree(profiles, files)

	tree.Fprint(os.Stdout, true, "", float64(cfg.CoverageThreshold))

	if !tree.IsProperlyCovered(float64(cfg.CoverageThreshold)) {
		color.Red("\n%s is not up to requested target: %d%%", cfg.Name, cfg.CoverageThreshold)
		os.Exit(1)
	}
}
