package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cov/internal/coverage"
	"github.com/PaloAltoNetworks/cov/internal/git"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "v0.0.0"
	commit  = "dev"
)

func main() {

	cobra.OnInitialize(initCobra)

	rootCmd := &cobra.Command{
		Use:           "cov cover.out...",
		Short:         "Analyzes coverage",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			branch := viper.GetString("branch")
			threshold := viper.GetInt("threshold")
			filters := viper.GetStringSlice("filters")
			ignored := viper.GetStringSlice("ignore")
			quiet := viper.GetBool("quiet")

			if viper.GetBool("version") {
				fmt.Printf("cov %s (%s)\n", version, commit)
				return nil
			}

			profiles, err := coverage.MergeProfiles(ignored, args)
			if err != nil {
				return fmt.Errorf("unable to read profiles: %s", err)
			}

			files := filters
			if branch != "" {
				gitFiles, err := git.GetDiffFiles(branch)
				if err != nil {
					return fmt.Errorf("unable to get change files for branch %s: %w", branch, err)
				}
				files = append(files, gitFiles...)
			}
			if len(files) == 0 {
				fmt.Println("no change in go files")
				return nil
			}

			tree := coverage.NewTree(profiles, files)

			if !quiet {
				tree.Fprint(os.Stdout, true, "", float64(threshold))
			}

			coverage := tree.GetCoverage()

			if threshold > 0 {
				if coverage < float64(threshold) {
					color.Red("Not up to requested coverage target. coverage: %.0f%% requested: %d%%\n", coverage, threshold)
					os.Exit(1)
				} else {
					color.Green("Up to requested target: coverage: %.0f%% requested: %d%%\n", coverage, threshold)
				}
			} else {
				fmt.Printf("Coverage: %.0f%%\n", coverage)
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().Bool("version", false, "show version")
	rootCmd.Flags().StringP("branch", "b", "", "The branch to use to check the patch coverage against. Example: master")
	rootCmd.Flags().IntP("threshold", "t", 0, "The target of coverage in percent that is requested")
	rootCmd.Flags().StringSliceP("filter", "f", nil, "The filters to use for coverage lookup")
	rootCmd.Flags().StringSliceP("ignore", "i", nil, "Define patterns to ignore matching files.")
	rootCmd.Flags().BoolP("quiet", "q", false, "Do not print details, just the verdict")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func initCobra() {
	viper.SetEnvPrefix("cov")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
