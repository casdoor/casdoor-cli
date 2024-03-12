/*
Copyright © 2024 Fabien CHEVALIER
*/
package cmd

import (
	"github.com/zalando/go-keyring"
	"os"

	"github.com/kyokomi/emoji/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Account struct {
	AccessToken     string
	CasdoorEndpoint string
}

var (
	flagDebug bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "casdoor",
	Short: emoji.Sprintf(":key: A clean and straightforward command line interface for Casdoor."),
	Long: emoji.Sprintf(`:rocket: Casdoor CLI 

:key: A clean and straightforward command line interface for Casdoor.

Casdoor CLI allows you to perform various operations using Casdoor's Public 
API such as:

- User management
- Permissions management

:information_source: Begin with using casdoor init in order to set up your configuration.`),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Init functions
	cobra.OnInitialize(InitLogger)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "enable debug output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// Subcommand handling

}

func InitLogger() {
	if flagDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "", // default time format
			NoColor:    false,
		}

		log.Logger = log.Output(consoleWriter)

	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
}

// SetupAccount factorize token and endpoint retrieval
func SetupAccount() (*Account, error) {
	service := "casdoor-cli"
	accessToken, err := keyring.Get(service, "access_token")
	if err != nil {
		log.Error().Msgf("Failed to retrieve access token: %v", err)
		return nil, err
	}
	casdoorEndpoint, err := keyring.Get(service, "endpoint")
	if err != nil {
		log.Error().Msgf("Failed to retrieve casdoor endpoint: %v", err)
		return nil, err
	}

	accountConfig := &Account{
		AccessToken:     accessToken,
		CasdoorEndpoint: casdoorEndpoint,
	}
	return accountConfig, nil
}
