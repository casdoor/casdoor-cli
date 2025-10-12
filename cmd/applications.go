package cmd

import (
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/sdv9972401/casdoor-cli/helpers"
	"gitlab.com/sdv9972401/casdoor-cli/models"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
)

var (
	appNameFlag        string
	clientIdFlag       string
	clientSecretFlag   string
	redirectUrisFlag   []string
	organizationFlag   string
	descriptionFlag    string
)

var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Manage Casdoor applications",
	Long:  "Manage Casdoor applications for OIDC integration",
}

var applicationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "list Casdoor applications",
	Long:  "list Casdoor applications",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}

		appManager := helpers.NewApplicationManager(config)
		applications, err := appManager.GetApplications()
		if err != nil {
			log.Error("Failed to get applications: ", err)
			return
		}
		utils.PrintApplicationsTable(applications)
	},
}

var applicationsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add Casdoor application",
	Long:  "add Casdoor application for OIDC integration",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}

		if appNameFlag == "" || clientIdFlag == "" || clientSecretFlag == "" {
			color.Red("Error: name, client-id, and client-secret flags are required")
			return
		}

		appManager := helpers.NewApplicationManager(config)
		
		application := &models.Application{
			Owner:        "admin",
			Name:         appNameFlag,
			DisplayName:  appNameFlag,
			Description:  descriptionFlag,
			Organization: organizationFlag,
			ClientId:     clientIdFlag,
			ClientSecret: clientSecretFlag,
			RedirectUris: redirectUrisFlag,
			TokenFormat:  "JWT",
			ExpireInHours: 168,
			RefreshExpireInHours: 168,
			EnablePassword: true,
			EnableSignUp: false,
			EnableSigninSession: false,
			EnableCodeSignin: false,
			EnableSamlCompress: false,
			EnableAutoSignin: false,
			Cert: "cert-built-in",
			GrantTypes: []string{"authorization_code", "implicit", "refresh_token"},
			ResponseTypes: []string{"code", "token", "id_token"},
			Scopes: []string{"openid", "profile", "email"},
		}

		err = appManager.AddApplication(application)
		if err != nil {
			color.Red("Failed to add application: %v", err)
			return
		}

		color.Green("Application '%s' added successfully", appNameFlag)
	},
}

var applicationsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete Casdoor application",
	Long:  "delete Casdoor application",
	Run: func(cmd *cobra.Command, args []string) {
		targetRoles := []string{
			"administrator",
		}
		config, err := checkLoggedInAndGetConfig(targetRoles)
		if err != nil {
			return
		}

		if appNameFlag == "" {
			color.Red("Error: name flag is required")
			return
		}

		appManager := helpers.NewApplicationManager(config)
		
		application := &models.Application{
			Owner: "admin",
			Name:  appNameFlag,
		}

		err = appManager.DeleteApplication(application)
		if err != nil {
			color.Red("Failed to delete application: %v", err)
			return
		}

		color.Green("Application '%s' deleted successfully", appNameFlag)
	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
	applicationsCmd.AddCommand(applicationsListCmd)
	applicationsCmd.AddCommand(applicationsAddCmd)
	applicationsCmd.AddCommand(applicationsDeleteCmd)

	// Add flags for application creation
	applicationsAddCmd.Flags().StringVarP(&appNameFlag, "name", "n", "", "Application name (required)")
	applicationsAddCmd.Flags().StringVarP(&clientIdFlag, "client-id", "c", "", "OAuth2 client ID (required)")
	applicationsAddCmd.Flags().StringVarP(&clientSecretFlag, "client-secret", "s", "", "OAuth2 client secret (required)")
	applicationsAddCmd.Flags().StringSliceVarP(&redirectUrisFlag, "redirect-uris", "r", []string{}, "OAuth2 redirect URIs (comma-separated)")
	applicationsAddCmd.Flags().StringVarP(&organizationFlag, "organization", "o", "built-in", "Organization name")
	applicationsAddCmd.Flags().StringVarP(&descriptionFlag, "description", "D", "", "Application description")

	// Add flags for application deletion
	applicationsDeleteCmd.Flags().StringVarP(&appNameFlag, "name", "n", "", "Application name (required)")
}