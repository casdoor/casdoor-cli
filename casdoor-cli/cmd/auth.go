package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/zalando/go-keyring"
)

type Account struct {
	AccessToken     string
	CasdoorEndpoint string
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
