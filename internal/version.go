package internal

import "github.com/spf13/cobra"

// Version is the release version of the easifem CLI.
const Version = "2026.10.31"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the easifem CLI version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := cmd.OutOrStdout().Write([]byte("easifem version " + Version + "\n"))
		return err
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
