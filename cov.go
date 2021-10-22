package main

import (
	"os"

	"github.com/fatih/color"
	"go.aporeto.io/cov/internal/configuration"
	"go.aporeto.io/cov/internal/coverage"
	"go.aporeto.io/cov/internal/git"
	"go.uber.org/zap"
)

func main() {

	cfg := configuration.NewConfiguration()

	zap.L().Debug("Coverage analysis", zap.Reflect("configuration", cfg))

	profiles, err := coverage.MergeProfiles(cfg.CoverageFilePaths...)
	if err != nil {
		zap.L().Fatal("Unable to read profiles", zap.Error(err))
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

	if cfg.CoverageThreshold > 0 {
		if coverage < float64(cfg.CoverageThreshold) {
			color.Red("\n%s is not up to requested coverage target.\n - current coverage: %.0f%%\n - requested: %d%%", cfg.Name, coverage, cfg.CoverageThreshold)
			os.Exit(1)
		} else {

			color.Green("\n%s is up to requested target.\n - current coverage: %.0f%%\n - requested: %d%%", cfg.Name, coverage, cfg.CoverageThreshold)
		}
	}

	color.Green("\n%s coverage: %.0f%%\n", cfg.Name, coverage)

}
