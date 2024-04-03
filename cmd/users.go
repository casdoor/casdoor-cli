package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli/helpers"
	"gitlab.com/sdv9972401/casdoor-cli/models"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
	"strings"
)

var (
	nameFlag string
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage Casdoor users",
	Long:  "Manage Casdoor users",
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "list Casdoor users",
	Long:  "list Casdoor users",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
			"editor",
			"lector",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}

		userManager := helpers.NewUserManager(config)
		users, err := userManager.GetUsers()
		utils.PrintTables(users)
	},
}

var usersAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add Casdoor user",
	Long:  "add Casdoor user",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
			"editor",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}
		utils.Colorize(color.CyanString, "[ℹ] follow the prompts in order to create a new user")
		userManager := helpers.NewUserManager(config)
		err = userManager.AddUser()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var usersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete Casdoor user",
	Long:  "delete Casdoor user",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}

		fmt.Print(color.YellowString("[⚠] This will delete the user %v. There is no undo. Are you sure about that ? [y/N]: ", nameFlag))

		var userResponse string
		_, err = fmt.Scanln(&userResponse)
		if err != nil {
			return
		}
		if strings.ToLower(userResponse) == "y" || strings.ToLower(userResponse) == "yes" {
			utils.Colorize(color.CyanString, "[ℹ] attempting to delete user %v", nameFlag)
			userManager := helpers.NewUserManager(config)
			err := userManager.DeleteUser(nameFlag)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			utils.Colorize(color.RedString, "[x] operation canceled")
		}
		return
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update Casdoor user",
	Long:  "update Casdoor user",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}
		userManager := helpers.NewUserManager(config)
		err = userManager.UpdateUser(nameFlag)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func checkLoggedInAndGetConfig(requiredRoles []string) (*models.CasdoorConfig, error) {
	config, err := initCasdoorConfig()
	if err != nil {
		log.Fatal(err)
	}
	existingTokenData, err := utils.KeyringToTokenData()
	if err != nil {
		utils.Colorize(color.RedString, "[x] you are not logged in. You can log in using casdoor login")
		return nil, err
	}

	if !helpers.HasRequiredRole(existingTokenData.IDTokenClaims.Groups, requiredRoles) {
		utils.Colorize(color.RedString, "[x] you don't have enough permissions to perform this action (required roles: %v)", requiredRoles)
		return nil, errors.New("insufficient permissions")
	}

	return config, nil
}

func init() {
	RootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersAddCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	usersCmd.AddCommand(userUpdateCmd)
	usersDeleteCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "name of the user")
	usersDeleteCmd.MarkFlagRequired("name")
	userUpdateCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "name of the user")
	userUpdateCmd.MarkFlagRequired("name")
}
