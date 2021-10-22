package main

import (
	"os"

	"github.com/fatih/color"
	"go.aporeto.io/cov/internal/configuration"
	"go.aporeto.io/cov/internal/coverage"
	"go.aporeto.io/cov/internal/git"
	"go.uber.org/zap"
	"golang.org/x/tools/cover"
)

func main() {

	cfg := configuration.NewConfiguration()

	zap.L().Debug("Coverage analysis", zap.Reflect("configuration", cfg))

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

	if !cfg.Quiet {
		tree.Fprint(os.Stdout, true, "", float64(cfg.CoverageThreshold))
	}

	coverage := tree.GetCoverage()

	if coverage < float64(cfg.CoverageThreshold) {
		color.Red("\n%s is not up to requested coverage target.\n - current coverage: %.0f%%\n - requested: %d%%", cfg.Name, coverage, cfg.CoverageThreshold)
		os.Exit(1)
	}

	color.Green("\n%s is up to requested target.\n - current coverage: %.0f%%\n - requested: %d%%", cfg.Name, coverage, cfg.CoverageThreshold)
}
