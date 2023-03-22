package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cov/internal/coverage"
	"github.com/PaloAltoNetworks/cov/internal/git"
	"github.com/PaloAltoNetworks/cov/internal/statuscheck"
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
			thresholdExitCode := viper.GetInt("threshold-exit-code")
			reportPath := viper.GetString("report-path")
			writeReport := viper.GetString("write-report")
			sendRepo := viper.GetString("send-repo")
			sendToken := viper.GetString("send-token")

			if viper.GetBool("version") {
				fmt.Printf("cov %s (%s)\n", version, commit)
				return nil
			}

			if sendRepo != "" {
				if err := statuscheck.Send(reportPath, sendRepo, sendToken); err != nil {
					return fmt.Errorf("unable to send report as status check: %w", err)
				}
				return nil
			}

			if len(ignored) == 0 {
				data, err := os.ReadFile(".covignore")
				if err == nil {
					for _, line := range strings.Split(string(data), "\n") {
						if !strings.HasPrefix(line, "#") {
							ignored = append(ignored, line)
						}
					}
				}
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
				if len(files) == 0 {
					fmt.Println("no change in go files")
					if writeReport != "" {
						if err := statuscheck.WriteNoop(reportPath); err != nil {
							return fmt.Errorf("unable to write noop status: %w", err)
						}
					}
					return nil
				}
			}

			tree := coverage.NewTree(profiles, files)
			if !quiet {
				tree.Fprint(os.Stderr, true, "", float64(threshold))
			}

			coverage := tree.GetCoverage()
			isSuccess := threshold > 0 && threshold <= int(coverage)

			if isSuccess {
				fmt.Printf("up to standard. %.0f%% / %d%%\n", coverage, threshold)
			} else if threshold > 0 {
				fmt.Printf("not up to standard. %.0f%% / %d%%\n", coverage, threshold)
			} else {
				fmt.Printf("%.0f%% / %d%%\n", coverage, threshold)
			}

			if writeReport != "" {
				if err := statuscheck.Write(reportPath, int(coverage), threshold); err != nil {
					return fmt.Errorf("unable to write status check: %w", err)
				}
			}

			if !isSuccess {
				os.Exit(thresholdExitCode)
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
	rootCmd.Flags().IntP("threshold-exit-code", "e", 1, "Set the exit code on coverage threshold miss")

	rootCmd.Flags().String("report-path", "cov.report", "Defines the path for the status report.")
	rootCmd.Flags().Bool("write-report", false, "If set, write a status check report into --report-path")
	rootCmd.Flags().String("send-repo", "", "If set, set the status report from --report-path as status check. format: [repo]/[owner]@[sha]")
	rootCmd.Flags().String("send-token", "", "If set, use this token to send the status. If empty, $GITHUB_TOKEN will be used")

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
