package helpers

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
)

func (um *UserManager) GetGroups() ([]map[string]interface{}, error) {
	groups, err := um.client.GetGroups()

	if err != nil {
		return nil, err
	}
	var groupList []map[string]interface{}

	for _, group := range groups {
		groupInfo := map[string]interface{}{
			"Name":  group.Name,
			"Owner": group.Owner,
		}

		groupList = append(groupList, groupInfo)
	}
	return groupList, nil
}

func (um *UserManager) AddGroup() error {
	name, err := um.promptGroupName()
	if err != nil {
		return err
	}

	group := casdoorsdk.Group{
		Name:  name,
		Owner: "casdoor-cli",
	}

	_, err = um.client.AddGroup(&group)

	if err != nil {
		return err
	}
	utils.Colorize(color.GreenString, "[✔] %v group has been added successfully", name)
	return nil
}

func (um *UserManager) DeleteGroup(name string) error {
	group := casdoorsdk.Group{
		Name:  name,
		Owner: "casdoor-cli",
	}

	checkGroups, _ := um.client.GetGroups()
	groupFound := false

	for _, checkGroup := range checkGroups {
		if checkGroup.Name == name {
			_, err := um.client.DeleteGroup(&group)
			if err != nil {
				return err
			}
			utils.Colorize(color.GreenString, "[✔] %v group has been deleted successfully", name)
			groupFound = true
			break
		}
	}
	if !groupFound {
		utils.Colorize(color.RedString, "[x] group %v doesn't exist", name)
	}
	return nil
}

func (um *UserManager) UpdateGroup(name string) error {
	checkGroups, _ := um.client.GetGroups()
	groupFound := false

	for _, checkGroup := range checkGroups {
		if checkGroup.Name == name {
			newName, err := um.promptGroupName()
			if err != nil {
				return err
			}

			group := casdoorsdk.Group{
				Name:  newName,
				Owner: checkGroup.Owner,
			}

			_, err = um.client.UpdateGroup(&group)

			if err != nil {
				return err
			}
			utils.Colorize(color.GreenString, "[✔] %v group has been updated successfully", newName)
			if err != nil {
				return err
			}
			groupFound = true
			break
		}
	}
	if !groupFound {
		utils.Colorize(color.RedString, "[x] group %v doesn't exist", name)
	}
	return nil
}

func (um *UserManager) promptGroupName() (string, error) {
	namePrompt := promptui.Prompt{
		Label: "Group Name",
	}
	return namePrompt.Run()
}
