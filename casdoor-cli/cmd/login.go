package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyokomi/emoji/v2"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	authCodeChan = make(chan string)
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your Casdoor account",
	Long: emoji.Sprint(`:key: Login to your Casdoor account.

This command will open your default browser and redirect you to the Casdoor login page.

:information_source: Casdoor CLI must be initialized before using this command (see 'casdoor init -h').
`),
	Run: func(cmd *cobra.Command, args []string) {
		emoji.Println(":key: Logging in... please wait for browser to open...")
		config, err := retrieveCredentials()
		if err != nil {
			log.Error().Msgf("Failed to retrieve credentials: %v", err)
			emoji.Println(":cross_mark: Failed to retrieve credentials. Have you run 'casdoor init' yet?")
			os.Exit(1)
			return
		}

		token, err := casdoorLogin(config)
		if err != nil {
			log.Error().Msgf("Login failed: %v", err)
			return
		}

		err = storeAccessToken(token)
		if err != nil {
			log.Error().Msgf("Failed to store access token: %v", err)
			return
		}

		emoji.Println(":white_check_mark: Login succeeded! You are all set :ok_hand: \n")
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func retrieveCredentials() (*CasdoorConfig, error) {
	service := "casdoor-cli"

	endpoint, err := keyring.Get(service, "endpoint")
	if err != nil {
		log.Error().Msgf("Failed to retrieve endpoint: %v\n", err)
		return nil, err
	}

	redirectURI, err := keyring.Get(service, "callback_url")
	if err != nil {
		log.Error().Msgf("Failed to retrieve redirect uri: %v\n", err)
		return nil, err
	}

	clientID, err := keyring.Get(service, "client_id")
	if err != nil {
		log.Error().Msgf("Failed to retrieve client id: %v\n", err)
		return nil, err
	}

	clientSecret, err := keyring.Get(service, "client_secret")
	if err != nil {
		log.Error().Msgf("Failed to retrieve client secret: %v\n", err)
		return nil, err
	}

	config := &CasdoorConfig{
		Endpoint:     endpoint,
		RedirectURI:  redirectURI,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	return config, nil
}

func casdoorLogin(config *CasdoorConfig) (string, error) {
	state, err := generateRandomState(32)
	if err != nil {
		log.Error().Msgf("Failed to generate random state: %v\n", err)
	}

	authURL := fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid&state=%s",
		config.Endpoint, config.ClientID, url.QueryEscape(config.RedirectURI), state)
	log.Debug().Msgf("Opening authorization URL: %s", authURL)
	err = browser.OpenURL(authURL)
	if err != nil {
		return "", fmt.Errorf("failed to open browser: %w", err)
	}

	log.Debug().Msg("Starting local server to receive authorization code")
	code, err := receiveAuthCode(config.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to receive authorization code: %w", err)
	}
	log.Debug().Msg("Received authorization code")

	log.Debug().Msg("Exchanging authorization code for access token")
	token, err := exchangeCodeForToken(config, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %w", err)
	}
	log.Debug().Msg("Access token granted")

	return token, nil
}

func receiveAuthCode(redirectURI string) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}
	port := u.Port()

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		authCodeChan <- authCode
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("✅ You have been successfully authenticated. You may now close this window."))
		if err != nil {
			log.Error().Msgf("Failed to write response: %v", err)
		}
	})

	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Error().Msgf("Failed to start local server: %v", err)
		}
	}()

	code := <-authCodeChan
	return code, nil
}

func exchangeCodeForToken(config *CasdoorConfig, code string) (string, error) {
	tokenURL := fmt.Sprintf("%s/api/login/oauth/access_token", config.Endpoint)
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"code":          {code},
	}

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to exchange code for token")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func generateRandomState(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func storeAccessToken(token string) error {
	log.Debug().Msg("Storing access token into keyring ...")
	service := "casdoor-cli"
	err := keyring.Set(service, "access_token", token)
	if err != nil {
		log.Error().Msgf("Failed to store access token: %v", err)
	}
	log.Debug().Msg("... done. All set.")
	return nil
}
