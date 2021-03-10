package commands

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use: "jam",
	}
)

func Execute() error {
	return rootCmd.Execute()
}
