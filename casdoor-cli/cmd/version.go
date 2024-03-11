/*
Copyright © 2024 Fabien CHEVALIER
*/
package cmd

import (
	"fmt"
	"github.com/kyokomi/emoji/v2"
	"github.com/spf13/cobra"
	"runtime"
)

var Version string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of Cobra CLI",
	Run: func(cmd *cobra.Command, args []string) {
		ShowOSAndAppVersion()
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ShowOSAndAppVersion() {
	emoji.Printf(":rocket: Casdoor CLI v%s\n", Version)
	fmt.Printf("on %s_%s", runtime.GOOS, runtime.GOARCH)
}
