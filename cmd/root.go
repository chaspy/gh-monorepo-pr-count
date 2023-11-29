/*
Copyright © 2023 Takeshi Kondo
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-pr-count",
	Short: "gh extension to count the number of PRs with the same label as the directory name",
	Long: `
	gh extension to count the number of PRs with the same label as the directory name.

	For example:

	gh pr-count 2023-10-01 2023-10-31 # Count the number of PRs in October 2023
	gh pr-count 2023-11-01            # Count the number of PRs since November 1st, 2023 until now
	`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gh-pr-count.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
