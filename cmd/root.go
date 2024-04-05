package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kyokomi/emoji/v2"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/sdv9972401/casdoor-cli/logger"
	"gitlab.com/sdv9972401/casdoor-cli/models"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

var RootCmd = &cobra.Command{
	Use:   "casdoor",
	Short: "A clean and straightforward command line interface for Casdoor.",
	Long: emoji.Sprintf(`:rocket: Casdoor CLI 

Casdoor CLI allows you to perform various operations using Casdoor's Public 
API such as:

- User management
- Permissions management
`),
}

var debug bool

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentPreRun = rootPreRun
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "verbose logging")

}

// rootPreRun is a pre-run function for the CLI app. It checks if a configuration file exists
// and tries to load it. If a configuration file is not found, it prompts the user to provide
// the path to their config.yaml file and creates a new configuration.
func rootPreRun(*cobra.Command, []string) {
	logger.ToggleDebug(debug)
	folderExist, fileExists := checkCasdoorConfig()

	if folderExist || fileExists {
		log.Debug("a config file has been found. Will now attempt to load it")
		config, err := initCasdoorConfig()
		if err != nil {
			log.Fatal(err)
		}
		log.Debug("config file loaded")
		// check if endpoint is avaliable
		resp, err := http.Get(config.Endpoint)
		if err != nil {
			utils.Colorize(color.RedString, "[x] endpoint %v is unavailable.", config.Endpoint)
			log.Fatal(err)
		} else {
			log.Debugf("connection to endpoint successful : %v", resp.StatusCode)
		}
	} else {
		log.Debug("no config file has been found. Will now initialize")
		utils.Colorize(color.YellowString, "[⚠] no config file detected. Please provide the path to your config.yaml file below :")

		prompt := promptui.Prompt{
			Label:   "config.yaml path",
			Default: "./config.yaml",
		}
		configPath, err := prompt.Run()
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		if err != nil {
			log.Errorf("error reading config file: %s. Please make sure config.yaml exists.", err)
		}
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		configKeys := []string{
			"casdoor_endpoint",
			"client_id",
			"client_secret",
			"certificate",
			"organization_name",
			"application_name",
			"redirect_uri",
		}

		for _, key := range configKeys {
			viper.GetString(key)
		}

		casdoorFolder := filepath.Join(usr.HomeDir, ".casdoor-cli")
		err = os.MkdirAll(casdoorFolder, 0755)
		if err != nil {
			log.Fatal(err)
		}

		err = createConfigFile(casdoorFolder)
		if err != nil {
			log.Fatal(err)
		}

		log.Debug("config file initialized")
	}
}

// initCasdoorConfig loads the Casdoor configuration from the config.yaml file.
// It decodes the base64-encoded values and returns a CasdoorConfig struct.
func initCasdoorConfig() (*models.CasdoorConfig, error) {
	_, configFile, err := getCasdoorFolderAndConfig()
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigFile(configFile)

	err = viper.ReadInConfig()
	if err != nil {
		log.Errorf("error reading config file: %s. Please make sure config.yaml exists.", err)
	}

	decodedConfig := make(map[string]string)
	for key, encodedValue := range viper.AllSettings() {
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedValue.(string))
		if err != nil {
			return nil, fmt.Errorf("error decoding base64 for %s: %v", key, err)
		}
		decodedConfig[key] = string(decodedBytes)
	}

	casdoorConfig := &models.CasdoorConfig{
		Endpoint:         decodedConfig["casdoor_endpoint"],
		ClientID:         decodedConfig["client_id"],
		ClientSecret:     decodedConfig["client_secret"],
		Certificate:      decodedConfig["certificate"],
		OrganizationName: decodedConfig["organization_name"],
		ApplicationName:  decodedConfig["application_name"],
		RedirectURI:      decodedConfig["redirect_uri"],
	}

	for key, value := range map[string]string{
		"casdoor_endpoint":  casdoorConfig.Endpoint,
		"client_id":         casdoorConfig.ClientID,
		"client_secret":     casdoorConfig.ClientSecret,
		"certificate":       casdoorConfig.Certificate,
		"organization_name": casdoorConfig.OrganizationName,
		"application_name":  casdoorConfig.ApplicationName,
		"redirect_uri":      casdoorConfig.RedirectURI,
	} {
		if value == "" {
			errorMsg := fmt.Sprintf("error getting config: %s", key)
			log.Error(errorMsg)
			return nil, errors.New(errorMsg)
		}
	}

	return casdoorConfig, err
}

// checkCasdoorConfig checks if the Casdoor configuration file and folder exist.
// It returns two boolean values representing the existence of the folder and the file.
func checkCasdoorConfig() (bool, bool) {
	casdoorFolder, configFile, err := getCasdoorFolderAndConfig()
	if err != nil {
		return false, false
	}

	folderExists := false
	fileExists := false
	if _, err = os.Stat(casdoorFolder); !os.IsNotExist(err) {
		folderExists = true

		if _, err := os.Stat(configFile); !os.IsNotExist(err) {
			fileExists = true
		}
	}

	return folderExists, fileExists
}

// getCasdoorFolderAndConfig returns the paths of the Casdoor configuration folder
// and the config.yaml file. If an error occurs while getting the user's home
// directory, the error is returned.
func getCasdoorFolderAndConfig() (string, string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", "", err
	}

	casdoorFolder := filepath.Join(usr.HomeDir, ".casdoor-cli")
	configFile := filepath.Join(casdoorFolder, "config.yaml")

	return casdoorFolder, configFile, nil
}

// createConfigFile creates a new Casdoor configuration file in the specified
// folder using the values from the user-provided config.yaml file.
// It base64-encodes the configuration values and writes them to the new file.
func createConfigFile(casdoorFolder string) error {
	configFile := filepath.Join(casdoorFolder, "config.yaml")

	encodedConfig := make(map[string]string)
	for key, value := range viper.AllSettings() {
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value.(string)))
		encodedConfig[key] = encodedValue
	}

	for key, value := range encodedConfig {
		viper.Set(key, value)
	}
	viper.SetConfigFile(configFile)

	err := viper.WriteConfigAs(configFile)
	if err != nil {
		return err
	}
	return nil
}
