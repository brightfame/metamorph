package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	rootCmd = &cobra.Command{
		Use:   "metamorph",
		Short: "Batch changes with a sprinkle of AI ✨",
		Long:  `MetaMorph is a tool for making batch updates to multiple repositories.`,
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("config-file", "c", "", "config file")

	// Add commands
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
