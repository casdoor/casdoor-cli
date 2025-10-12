package helpers

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"gitlab.com/sdv9972401/casdoor-cli/models"
)

type ApplicationManager struct {
	client *casdoorsdk.Client
}

func NewApplicationManager(config *models.CasdoorConfig) *ApplicationManager {
	client := casdoorsdk.NewClient(config.Endpoint,
		config.ClientID,
		config.ClientSecret,
		config.Certificate,
		config.OrganizationName,
		config.ApplicationName)

	return &ApplicationManager{client: client}
}

func (am *ApplicationManager) GetApplications() ([]map[string]interface{}, error) {
	applications, err := am.client.GetApplications()
	if err != nil {
		return nil, err
	}

	var applicationList []map[string]interface{}
	for _, app := range applications {
		applicationInfo := map[string]interface{}{
			"Name":          app.Name,
			"DisplayName":   app.DisplayName,
			"Organization":  app.Organization,
			"ClientId":      app.ClientId,
			"Description":   app.Description,
		}
		applicationList = append(applicationList, applicationInfo)
	}

	return applicationList, nil
}

func (am *ApplicationManager) AddApplication(app *models.Application) error {
	casdoorApp := &casdoorsdk.Application{
		Owner:                app.Owner,
		Name:                 app.Name,
		DisplayName:          app.DisplayName,
		Description:          app.Description,
		Organization:         app.Organization,
		ClientId:             app.ClientId,
		ClientSecret:         app.ClientSecret,
		RedirectUris:         app.RedirectUris,
		TokenFormat:          app.TokenFormat,
		ExpireInHours:        app.ExpireInHours,
		RefreshExpireInHours: app.RefreshExpireInHours,
		EnablePassword:       app.EnablePassword,
		EnableSignUp:         app.EnableSignUp,
		EnableSigninSession:  app.EnableSigninSession,
		EnableCodeSignin:     app.EnableCodeSignin,
		EnableSamlCompress:   app.EnableSamlCompress,
		EnableAutoSignin:     app.EnableAutoSignin,
		Cert:                 app.Cert,
		GrantTypes:           app.GrantTypes,
	}

	_, err := am.client.AddApplication(casdoorApp)
	return err
}

func (am *ApplicationManager) DeleteApplication(app *models.Application) error {
	casdoorApp := &casdoorsdk.Application{
		Owner: app.Owner,
		Name:  app.Name,
	}

	_, err := am.client.DeleteApplication(casdoorApp)
	return err
}