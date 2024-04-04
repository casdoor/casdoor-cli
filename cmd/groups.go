package cmd

import (
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli/helpers"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
	"strings"
)

var groupNameFlag string

var permissionsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage Casdoor permissions",
	Long:  "Manage Casdoor permissions",
}

var permissionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "list Casdoor permissions",
	Long:  "list Casdoor permissions",
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
		groups, err := userManager.GetGroups()
		utils.PrintTables(groups)
	},
}

var permissionsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add Casdoor permission",
	Long:  "add Casdoor permission",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
			"editor",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}
		utils.Colorize(color.CyanString, "[ℹ] follow the prompts in order to create a new group")
		userManager := helpers.NewUserManager(config)
		err = userManager.AddGroup()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var permissionsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete Casdoor permission",
	Long:  "delete Casdoor permission",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}

		fmt.Print(color.YellowString("[⚠] This will delete the group %v. There is no undo. Are you sure about that ? [y/N]: ", groupNameFlag))

		var userResponse string
		_, err = fmt.Scanln(&userResponse)
		if err != nil {
			return
		}
		if strings.ToLower(userResponse) == "y" || strings.ToLower(userResponse) == "yes" {
			utils.Colorize(color.CyanString, "[ℹ] attempting to delete group %v", groupNameFlag)
			userManager := helpers.NewUserManager(config)
			err := userManager.DeleteGroup(groupNameFlag)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			utils.Colorize(color.RedString, "[x] operation canceled")
		}
		return
	},
}

var permissionUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update Casdoor permission",
	Long:  "update Casdoor permission",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}
		userManager := helpers.NewUserManager(config)
		err = userManager.UpdateGroup(groupNameFlag)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(permissionsCmd)
	permissionsCmd.AddCommand(permissionsListCmd)
	permissionsCmd.AddCommand(permissionsAddCmd)
	permissionsCmd.AddCommand(permissionsDeleteCmd)
	permissionsCmd.AddCommand(permissionUpdateCmd)
	permissionsDeleteCmd.Flags().StringVarP(&groupNameFlag, "name", "n", "", "name of the group")
	permissionsDeleteCmd.MarkFlagRequired("name")
	permissionUpdateCmd.Flags().StringVarP(&groupNameFlag, "name", "n", "", "name of the group")
	permissionUpdateCmd.MarkFlagRequired("name")
}
