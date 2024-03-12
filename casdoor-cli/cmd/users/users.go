/*
Copyright © 2024 Fabien
*/
package users

import (
	"encoding/json"
	"fmt"
	"github.com/kyokomi/emoji/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli-go/cmd"
	"io"
	"net/http"
)

var accountConfig *cmd.Account

// usersCmd represents the user command
var usersCmd = &cobra.Command{
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
	var err error
	usersCmd.AddCommand(listCmd)
	accountConfig, err = cmd.SetupAccount()
	if err != nil {
		log.Error().Msgf("Failed to setup account: %v", err)
	}

	cobra.OnInitialize(cmd.InitLogger)
	cmd.RootCmd.AddCommand(usersCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// usersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	usersCmd.Flags().BoolP("info", "i", false, "show logged in user information")
}

func getUserInfo(account *cmd.Account) (GlobalUsersResponse, error) {
	url := fmt.Sprintf("%s/api/get-account", account.CasdoorEndpoint)
	log.Debug().Msgf("Fetching logged in username using %s path", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error().Msgf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", account.AccessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Msgf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf("Failed to read response body: %v", err)
	}

	log.Debug().Msgf("Response status code: %d", resp.StatusCode)
	log.Debug().Msgf("Response headers: %v", resp.Header)

	usersResponse := GlobalUsersResponse{}
	err = json.Unmarshal(body, &usersResponse)
	if err != nil {
		log.Error().Msgf("Failed to unmarshal response body: %v", err)
	}

	return usersResponse, nil
}
