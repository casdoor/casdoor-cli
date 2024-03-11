/*
Copyright © 2024 Fabien CHEVALIER
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/kyokomi/emoji/v2"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"net"
	"net/url"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: emoji.Sprintf(":wrench: Initialize Casdoor CLI configuration"),
	Long: emoji.Sprintf(`:wrench: Casdoor CLI initialization

You have to run this command at least once in order to set up your configuration.
You then must provide the following information:

:cloud: Casdoor Endpoint
:key: Casdoor Application Client ID
:lock: Casdoor Application Client Secret

Casdoor CLI will provide you with a callback URL that need to be set up on your Casdoor Application's settings page.

Required steps are fully explained on Casdoor CLI's documentation, but you can find more information
on the official Casdoor website : https://casdoor.org/docs/
`),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Retrieving credentials from user's input...")
		credentials, err := getCredentials()
		if err != nil {
			log.Fatal().Msg(err.Error())
			return
		}
		log.Debug().Msg("... done")

		log.Debug().Msg("Storing credentials...")
		err = storeCredentials(credentials)
		if err != nil {
			log.Fatal().Msg(err.Error())
			return
		}
		log.Debug().Msg("... done")

		emoji.Println(":white_check_mark: Credentials stored successfully! You can now run `casdoor login` to authenticate.")
		fmt.Println("Press enter to continue...")
		_, _ = fmt.Scanln()
	},
}

type CasdoorConfig struct {
	Endpoint     string
	RedirectURI  string
	ClientID     string
	ClientSecret string
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func getAvailablePort() (string, error) {
	log.Debug().Msg("Retrieving available port...")
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Error().Msgf("Failed to retrieve available port: %v\n", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Error().Msgf("Failed to close listener: %v\n", err)
		}
	}(listener)
	log.Debug().Msgf("... done. Using port : %d", listener.Addr().(*net.TCPAddr).Port)

	return fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port), nil
}

func getCredentials() (*CasdoorConfig, error) {
	emoji.Println(`In order to authenticate the Casdoor CLI, you must provide the following information:

:cloud: Casdoor Endpoint
:key: Casdoor Application Client ID
:lock: Casdoor Application Client Secret

Then, Casdoor CLI will provide you with a callback URL that need to be set up on your Casdoor Application's settings page.

:information_source: Application Credentials are securely stored within your system keyring backend API.`)

	fmt.Println("\nPress enter to continue...")
	_, _ = fmt.Scanln()

	// Get casdoorEndpoint
	validate := func(input string) error {
		_, err := url.ParseRequestURI(input)
		if err != nil {
			return errors.New("invalid URL")
		}
		return nil
	}

	endpointPrompt := promptui.Prompt{
		Label:    "Casdoor Endpoint",
		Validate: validate,
		Default:  "http://localhost:8000",
	}

	casdoorEndpoint, err := endpointPrompt.Run()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	// Get casdoorClientId
	clientIdPrompt := promptui.Prompt{
		Label: "Casdoor Client ID",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("input cannot be empty")
			}
			return nil
		},
		Default: "",
	}

	casdoorClientId, err := clientIdPrompt.Run()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	// Get casdoorClientSecret
	clientSecretPrompt := promptui.Prompt{
		Label: "Casdoor Client Secret",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("input cannot be empty")
			}
			return nil
		},
		Default: "",
		Mask:    '*',
	}

	casdoorClientSecret, err := clientSecretPrompt.Run()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	// Get redirectURI
	callbackPort, err := getAvailablePort()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	callbackURL := fmt.Sprintf("http://localhost:%s/callback", callbackPort)

	config := &CasdoorConfig{
		Endpoint:     casdoorEndpoint,
		RedirectURI:  callbackURL,
		ClientID:     casdoorClientId,
		ClientSecret: casdoorClientSecret,
	}

	return config, nil
}

func storeCredentials(config *CasdoorConfig) error {
	service := "casdoor-cli"

	log.Debug().Msgf("storing endpoint: %v\n", config.Endpoint)
	err := keyring.Set(service, "endpoint", config.Endpoint)
	if err != nil {
		log.Error().Msgf("Failed to store endpoint: %v\n", err)
		return err
	}

	log.Debug().Msgf("storing client id: %v\n", config.ClientID)
	err = keyring.Set(service, "client_id", config.ClientID)
	if err != nil {
		log.Error().Msgf("Failed to store client id: %v\n", err)
		return err
	}

	log.Debug().Msg("storing client secret: masked")
	err = keyring.Set(service, "client_secret", config.ClientSecret)
	if err != nil {
		log.Error().Msgf("Failed to store client secret: %v\n", err)
		return err
	}

	log.Debug().Msgf("storing callback url: %v\n", config.RedirectURI)
	err = keyring.Set("casdoor-cli", "callback_url", config.RedirectURI)
	if err != nil {
		log.Error().Msgf("Failed to store callback url: %v\n", err)
		return err
	}

	emoji.Println(":warning: Please set the following callback URL on your Casdoor Application settings: %s\n",
		config.RedirectURI)

	return nil
}
