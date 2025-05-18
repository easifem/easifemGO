/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package internal

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// setenvCmd represents the setenv command
var setenvCmd = &cobra.Command{
	Use:   "setenv",
	Short: "Set environment variables for running easifem on your system.",
	Long:  easifem_setenv_intro,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(easifem_banner)
		if err := cmd.Help(); err != nil {
			log.Fatalln("[err] :: setenv.go | cmd.Help() ➡️ ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(setenvCmd)
}
