package cmd

import (
	"fmt"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/fatih/color"
	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli/models"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
	"strings"
)

var (
	infoFlag bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your Casdoor account",
	Long: emoji.Sprint(`:key: Login utility. See login -h to see all available options.

This command will open your default browser and redirect you to the Casdoor login page. Once connected, 
you may start using Casdoor CLI.
`),
	Run: func(cmd *cobra.Command, args []string) {
		runLogin()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your Casdoor account",
	Long:  "Logout from your Casdoor account",
	Run: func(cmd *cobra.Command, args []string) {
		runLogout()
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolVarP(&infoFlag, "info", "i", false, "get logged in user information")
	RootCmd.AddCommand(logoutCmd)
}

// runLogin performs the login operation for the Casdoor account.
//
// This function initializes Casdoor configuration, reads the saved token data,
// and logs in if the token is expired or doesn't exist. If the infoFlag is set,
// it displays the logged-in user's information.
func runLogin() {
	config, err := initCasdoorConfig()
	if err != nil {
		log.Fatal(err)
	}

	existingTokenData, err := utils.KeyringToTokenData()
	if err != nil {
		utils.Colorize(color.YellowString, "[⚠] %s", err.Error())
	}

	if err == nil && existingTokenData != nil {
		if infoFlag {
			displayLoggedInUserInfo(existingTokenData)
		} else {
			utils.Colorize(color.GreenString, "[✔] you are already logged in as %s", existingTokenData.IDTokenClaims.Name)
		}
	} else {
		utils.Colorize(color.CyanString, "[ℹ] saved token may be expired or invalid")
		attemptLoginWithErrorHandler(config)
	}
}

// displayLoggedInUserInfo displays the logged-in user's information.
//
// tokenData: the token data related to the logged-in user.
// This function displays the logged-in user's username, ID, owner, and group.
func displayLoggedInUserInfo(tokenData *models.TokenData) {
	utils.Colorize(color.CyanString, "[ℹ] current logged in user: %s", tokenData.IDTokenClaims.Name)

	loggedInUserInfo := map[string]interface{}{
		"username": tokenData.IDTokenClaims.Name,
		"id":       tokenData.IDTokenClaims.Sub,
		"owner":    tokenData.IDTokenClaims.Owner,
		"group":    strings.TrimPrefix(strings.Join(tokenData.IDTokenClaims.Groups, ", "), "casdoor-cli/"),
	}
	utils.PrintTable(loggedInUserInfo)
}

// attemptLoginWithErrorHandler attempts Casdoor login with error handling.
//
// config: Casdoor configuration to be used during the login attempt.
// This function attempts to log in with OAuthHandler, parses the OAuth response,
// and, if successful, saves the token data to the keyring.
func attemptLoginWithErrorHandler(config *models.CasdoorConfig) {
	utils.Colorize(color.CyanString, "[ℹ] attempting to log you in...")
	data, err := OAuthHandler(config)
	if err != nil {
		log.Fatal(err)
	}

	tokenData, err := ParseOAuthResponse(data)
	if err != nil {
		log.Fatal(err)
	}

	err = utils.TokenDataToKeyring(tokenData)
	if err != nil {
		log.Error(err)
	} else {
		utils.Colorize(color.GreenString, "[✔] you are now logged in as %s. session credentials will expire in 1 hour", tokenData.IDTokenClaims.Name)
	}
}

// runLogout logs out the user from their Casdoor account.
//
// The function checks if the user really wants to logout by asking for confirmation.
// If the confirmation is given, it logs out the user and clears the saved token.
func runLogout() {
	config, err := initCasdoorConfig()
	if err != nil {
		log.Fatal(err)
	}
	existingTokenData, err := utils.KeyringToTokenData()
	client := casdoorsdk.NewClient(config.Endpoint,
		config.ClientID,
		config.ClientSecret,
		config.Certificate,
		config.OrganizationName,
		config.ApplicationName)

	token := casdoorsdk.Token{
		Name: strings.TrimPrefix(existingTokenData.IDTokenClaims.Jti, "admin/"),
	}

	if userWantsToLogout() {
		utils.Colorize(color.CyanString, "[ℹ] logging you out")
		_, err = client.DeleteToken(&token)
		if err != nil {
			return
		}
		err = utils.ClearSavedToken()
		if err != nil {
			log.Fatal(err)
		}
		utils.Colorize(color.GreenString, "[✔] logged out successfully")
	} else {
		utils.Colorize(color.RedString, "[x] token clearing operation cancelled")
	}
}

// userWantsToLogout checks if the user really wants to logout.
//
// bool: true if the user gives a confirmation in the form of "y" or "yes", false otherwise.
// This function prompts the user for confirmation to logout and returns the answer.
func userWantsToLogout() bool {
	fmt.Print(color.YellowString("[⚠] this will delete all saved token. You will need to log in again in order to use Casdoor CLI. Are you sure about that ? [y/N]: "))

	var userResponse string
	_, err := fmt.Scanln(&userResponse)
	if err != nil {
		return false
	}

	return strings.ToLower(userResponse) == "y" || strings.ToLower(userResponse) == "yes"
}
