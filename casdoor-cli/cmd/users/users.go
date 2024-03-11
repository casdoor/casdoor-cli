/*
Copyright © 2024 Fabien
*/
package users

import (
	"github.com/kyokomi/emoji/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"gitlab.com/sdv9972401/casdoor-cli-go/cmd"
)

// usersCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "users",
	Short: emoji.Sprintf(":bust_in_silhouette: Manage users of your application"),
	Long: emoji.Sprintf(`:bust_in_silhouette: Manage users of your application

Using this command, you can:

- 'list' all users of your Casdoor application
- 'get' a detailed information about a user
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

func getAccount() (*Account, error) {
	service := "casdoor-cli"
	accessToken, err := keyring.Get(service, "access_token")
	if err != nil {
		log.Error().Msgf("Failed to retrieve access token: %v", err)
		return nil, err
	}
	casdoorEndpoint, err := keyring.Get(service, "endpoint")
	if err != nil {
		log.Error().Msgf("Failed to retrieve casdoor endpoint: %v", err)
		return nil, err
	}

	accountConfig := &Account{
		AccessToken:     accessToken,
		CasdoorEndpoint: casdoorEndpoint,
	}

	return accountConfig, nil
}
