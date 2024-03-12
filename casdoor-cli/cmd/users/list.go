/*
Copyright © 2024 Fabien
*/
package users

import (
	"encoding/json"
	"fmt"
	"github.com/kyokomi/emoji/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli-go/cmd"
	"io"
	"net/http"
	"os"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: emoji.Sprintf(":bust_in_silhouette: List all users of your application"),
	Long:  `Output a list of users formatted in a table.`,
	Run: func(cmd *cobra.Command, args []string) {
		users, err := fetchUsers(accountConfig)
		if err != nil {
			log.Error().Msgf("Failed to fetch users: %v", err)
			return
		}
		printUsersInTable(users)
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func fetchUsers(account *cmd.Account) (GlobalUsersResponse, error) {
	url := fmt.Sprintf("%s/api/get-global-users", account.CasdoorEndpoint)
	log.Debug().Msgf("Fetching users using %s path", url)
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

func printUsersInTable(usersResponse GlobalUsersResponse) {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{
		"Name",
		"Organization",
		"ID",
		"Email",
		"Verified",
		"Admin",
	}

	table.SetHeader(header)

	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
	)

	table.SetAutoMergeCells(false)
	table.SetRowLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, user := range usersResponse.Data {
		emailVerified := emoji.Sprint(":cross_mark:")
		if user.EmailVerified {
			emailVerified = emoji.Sprint(":heavy_check_mark:")
		}

		isAdmin := emoji.Sprint(":cross_mark:")
		if user.IsAdmin {
			isAdmin = emoji.Sprint(":heavy_check_mark:")
		}

		data := []string{user.Name, user.Owner, user.ID, user.Email, emailVerified, isAdmin}
		table.Append(data)
	}

	table.Render()
	emoji.Printf("\n:cloud: Casdoor Enpoint : %s", accountConfig.CasdoorEndpoint)
}
