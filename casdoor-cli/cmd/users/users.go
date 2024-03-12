/*
Copyright © 2024 Fabien
*/
package users

import (
	"github.com/kyokomi/emoji/v2"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli-go/cmd"
)

// usersCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users of your application",
	Long: emoji.Sprintf(`:bust_in_silhouette: Manage users of your application

Using this command, you can:

- 'list' all users of your Casdoor application
- 'get' detailed information about a user
- 'add' a new user
- 'update' an existing user
- 'delete' an existing user

:information_source: You must have initialized Casdoor CLI application and have logged in before using this command.`),
}

func init() {
	cobra.OnInitialize(cmd.InitLogger)
	cmd.RootCmd.AddCommand(userCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// usersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// usersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
