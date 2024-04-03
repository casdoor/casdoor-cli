package helpers

import (
	"errors"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"gitlab.com/sdv9972401/casdoor-cli/models"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
	"strings"
	"time"
)

type UserManager struct {
	client *casdoorsdk.Client
}

func NewUserManager(config *models.CasdoorConfig) *UserManager {
	client := casdoorsdk.NewClient(config.Endpoint,
		config.ClientID,
		config.ClientSecret,
		config.Certificate,
		config.OrganizationName,
		config.ApplicationName)

	return &UserManager{client: client}
}

func (um *UserManager) GetUsers() ([]map[string]interface{}, error) {
	users, err := um.client.GetUsers()

	if err != nil {
		return nil, err
	}
	var userList []map[string]interface{}

	for _, user := range users {
		userInfo := map[string]interface{}{
			"Name":  user.Name,
			"Owner": user.Owner,
			"Email": user.Email,
			"Id":    user.Id,
			"Roles": strings.TrimPrefix(strings.TrimPrefix(user.Groups[0], "casdoor-cli/"), "casdoor-cli/"),
		}

		userList = append(userList, userInfo)
	}
	return userList, nil
}

func (um *UserManager) AddUser() error {
	name, email, password, err := um.promptUserInput()
	if err != nil {
		return err
	}

	rolesResult, err := um.promptUserRoles()
	if err != nil {
		return err
	}

	user := casdoorsdk.User{
		Name:              name,
		Owner:             "casdoor-cli",
		Email:             email,
		Password:          password,
		Groups:            []string{rolesResult},
		Type:              "normal-user",
		CreatedTime:       time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		SignupApplication: "casdoor-cli",
	}

	_, err = um.client.AddUser(&user)

	if err != nil {
		return err
	}
	utils.Colorize(color.GreenString, "[✔] %v has been added successfully", name)
	return nil
}

func (um *UserManager) DeleteUser(name string) error {
	user := casdoorsdk.User{
		Name:  name,
		Owner: "casdoor-cli",
	}

	checkUsers, _ := um.client.GetUsers()
	userFound := false

	for _, checkUser := range checkUsers {
		if checkUser.Name == name {
			_, err := um.client.DeleteUser(&user)
			if err != nil {
				return err
			}
			utils.Colorize(color.GreenString, "[✔] %v has been deleted successfully", name)
			userFound = true
			break
		}
	}
	if !userFound {
		utils.Colorize(color.RedString, "[x] user %v doesn't exist", name)
	}
	return nil
}

func (um *UserManager) UpdateUser(name string) error {
	checkUsers, _ := um.client.GetUsers()
	userFound := false

	for _, checkUser := range checkUsers {
		if checkUser.Name == name {
			userInfo, err := um.client.GetUser(name)
			if err != nil {
				return err
			}

			email, err := um.promptUserEmail(userInfo.Email)
			if err != nil {
				return err
			}

			password, err := um.promptUserPassword()
			if err != nil {
				return err
			}

			rolesResult, err := um.promptUserRoles()
			if err != nil {
				return err
			}

			user := casdoorsdk.User{
				Name:              name,
				Id:                userInfo.Id,
				Owner:             userInfo.Owner,
				Email:             email,
				Password:          password,
				Groups:            []string{rolesResult},
				Type:              "normal-user",
				CreatedTime:       time.Now().UTC().Format("2006-01-02T15:04:05Z"),
				SignupApplication: userInfo.SignupApplication,
				DisplayName:       name,
			}

			_, err = um.client.UpdateUser(&user)

			if err != nil {
				return err
			}
			utils.Colorize(color.GreenString, "[✔] %v has been updated successfully", name)
			if err != nil {
				return err
			}
			userFound = true
			break
		}
	}
	if !userFound {
		utils.Colorize(color.RedString, "[x] user %v doesn't exist", name)
	}
	return nil
}

func (um *UserManager) promptUserInput() (string, string, string, error) {
	namePrompt := promptui.Prompt{
		Label: "Name",
	}
	name, err := namePrompt.Run()
	if err != nil {
		return "", "", "", err
	}

	emailPrompt := promptui.Prompt{
		Label: "Email",
	}
	email, err := emailPrompt.Run()
	if err != nil {
		return "", "", "", err
	}

	password, err := um.promptUserPassword()
	if err != nil {
		return "", "", "", err
	}

	return name, email, password, nil
}

func (um *UserManager) promptUserRoles() (string, error) {
	var rolesName []string
	roles, _ := um.client.GetGroups()

	for _, role := range roles {
		rolesName = append(rolesName, role.Name)
	}

	rolesPrompt := promptui.Select{
		Label: "Roles",
		Items: rolesName,
	}
	_, rolesResult, err := rolesPrompt.Run()

	return rolesResult, err
}

func (um *UserManager) promptUserEmail(defaultEmail string) (string, error) {
	emailPrompt := promptui.Prompt{
		Label:   "Email",
		Default: defaultEmail,
	}
	email, err := emailPrompt.Run()

	return email, err
}

func (um *UserManager) promptUserPassword() (string, error) {
	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}
	passwordPrompt := promptui.Prompt{
		Label:    "Password",
		Mask:     '*',
		Validate: validate,
	}
	password, err := passwordPrompt.Run()

	return password, err
}
