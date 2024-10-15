package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/brightfame/metamorph/internal/config"
	"github.com/brightfame/metamorph/pkg/pipeline"
	"github.com/brightfame/metamorph/pkg/runner"
)

var (
	verbose bool
	rootCmd = &cobra.Command{
		Use:   "metamorph",
		Short: "MetaMorph is a dependency updater tool",
		Long:  `MetaMorph is a CLI app for keeping dependencies up to date, even with backwards incompatible changes.`,
	}

	applyCmd = &cobra.Command{
		Use:   "apply [patch.sh]",
		Short: "Apply a patch to the specified repository",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// initialize the config
			cfg, err := config.DefaultConfig()
			if err != nil {
				return err
			}

			// check for GitLab CI username from GITLAB_CI_USERNAME
			if username, ok := os.LookupEnv("GITLAB_CI_USERNAME"); ok {
				cfg.PlatformAuthConfig.Username = username
			}

			// check for GitLab CI token from GITLAB_CI_TOKEN
			if token, ok := os.LookupEnv("GITLAB_CI_TOKEN"); ok {
				cfg.PlatformAuthConfig.Password = token
			}

			// if no manifest file is provided, then abort
			manifestFile, err := cmd.Flags().GetString("manifest")
			if err != nil || manifestFile == "" {
				return fmt.Errorf("no manifest file provided")
			}

			// load the manifest file
			p, err := pipeline.LoadManifestFile(cfg, manifestFile)
			if err != nil {
				return err
			}

			// create a basic pipeline with a single step and execute it
			//
			// p := pipeline.New(cfg, "apply")
			// p.AddStep(pipeline.Step{
			// 	Name:     "apply",
			// 	Image:    "golangci/golangci-lint:latest",
			// 	Commands: args,
			// 	Env:      map[string]string{"GOOS": "linux"},
			// 	WorkDir:  "/work",
			// })

			// // validate the pipeline
			// if err = p.Validate(); err != nil {
			// 	return err
			// }

			// create a new runner instance and execute the pipeline
			runner := runner.New(cfg, p)
			results, err := runner.Run(cmd.Context())
			if err != nil {
				return err
			}

			// print the results
			for _, result := range results {
				fmt.Printf("Step: %s\n", result.StepName)
				fmt.Printf("Exit Code: %d\n", result.ExitCode)
				fmt.Printf("Output: %s\n", string(result.Output))
				fmt.Printf("Error: %v\n", result.Error)
				fmt.Printf("Duration: %s\n", result.Duration)
			}

			return nil
		},
	}

	checkCmd = &cobra.Command{
		Use:   "check [repo-path]",
		Short: "Check for available updates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := args[0]
			if verbose {
				fmt.Printf("Checking for updates in %s\n", repoPath)
			}
			// Add your check logic here
			return nil
		},
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("config-file", "c", "", "config file")

	// Add commands
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(checkCmd)

	// Apply-specific flags
	applyCmd.Flags().Bool("dry-run", false, "show what would be updated without making changes")
	applyCmd.Flags().String("manifest", "", "path to the manifest file")
	applyCmd.Flags().StringP("branch", "b", "", "branch to use for applying changes")
	applyCmd.Flags().StringP("commit-msg", "m", "", "commit message to use for the commit")

	// Check command flags
	checkCmd.Flags().String("format", "text", "output format (text, json)")
}

func main() {
	// metamorph apply --branch=node-22 -m 'upgrade to node 22' upgrade.sh
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
