package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
	"gitlab.com/sdv9972401/casdoor-cli/models"
)

const chunkSize int = 1024

func TokenDataToKeyring(tokenData *models.TokenData) error {
	var roleNames []string

	for _, role := range tokenData.IDTokenClaims.Groups {
		roleNames = append(roleNames, role)
	}

	if len(roleNames) == 0 {
		roleNames = append(roleNames, "not settled")
	}

	keyringData := map[string]string{
		"access_token":  tokenData.OAuth2Token.AccessToken,
		"refresh_token": tokenData.OAuth2Token.RefreshToken,
		"token_type":    tokenData.OAuth2Token.TokenType,
		"expiry":        tokenData.OAuth2Token.Expiry.String(),
		"owner":         tokenData.IDTokenClaims.Owner,
		"name":          tokenData.IDTokenClaims.Name,
		"id":            tokenData.IDTokenClaims.Sub,
		"jti":           strings.TrimPrefix(tokenData.IDTokenClaims.Jti, "admin/"),
		"is_admin":      strconv.FormatBool(tokenData.IDTokenClaims.IsAdmin),
		"groups":        strings.TrimPrefix(strings.Join(tokenData.IDTokenClaims.Groups, ", "), "casdoor-cli/"),
	}
	for key, value := range keyringData {
		err := saveChunkedData("casdoor-cli", key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func KeyringToTokenData() (*models.TokenData, error) {
	keyringData := []string{
		"access_token",
		"refresh_token",
		"token_type",
		"expiry",
		"owner",
		"name",
		"id",
		"jti",
		"groups",
		"is_admin",
	}

	var tokenData = new(models.TokenData)
	nonEmptyValues := 0

	for _, key := range keyringData {
		numParts := 0
		for {
			chunkKey := fmt.Sprintf("%s_chunk_%d", key, numParts)
			_, err := keyring.Get("casdoor-cli", chunkKey)
			if err != nil {
				if err.Error() != "secret not found in keyring" {
					return nil, err
				}
				break
			}
			numParts++
		}

		if numParts > 0 {
			nonEmptyValues++
			value, err := loadChunkedData("casdoor-cli", key, numParts)
			if err != nil {
				return nil, err
			}

			// Proceed with assigning the value to the appropriate field in tokenData
			switch key {
			case "access_token":
				tokenData.OAuth2Token.AccessToken = value
			case "refresh_token":
				tokenData.OAuth2Token.RefreshToken = value
			case "token_type":
				tokenData.OAuth2Token.TokenType = value
			case "expiry":
				expiryTime, _ := time.Parse(time.RFC3339, value)
				tokenData.OAuth2Token.Expiry = expiryTime
			case "owner":
				tokenData.IDTokenClaims.Owner = value
			case "name":
				tokenData.IDTokenClaims.Name = value
			case "id":
				tokenData.IDTokenClaims.Sub = value
			case "jti":
				tokenData.IDTokenClaims.Jti = value
			case "is_admin":
				isAdmin, _ := strconv.ParseBool(value)
				tokenData.IDTokenClaims.IsAdmin = isAdmin
			case "groups":
				roleNames := strings.Split(value, ", ")
				tokenData.IDTokenClaims.Groups = roleNames
			}
		}
	}
	if nonEmptyValues == 0 {
		return nil, fmt.Errorf("no token data found in the keychain")
	}
	return tokenData, nil
}

func ClearSavedToken() error {
	keyringData := []string{
		"access_token",
		"refresh_token",
		"token_type",
		"expiry",
		"owner",
		"name",
		"id",
		"jti",
		"groups",
		"is_admin",
	}

	for _, key := range keyringData {
		count := 0
		for {
			chunkKey := fmt.Sprintf("%s_chunk_%d", key, count)
			err := keyring.Delete("casdoor-cli", chunkKey)
			if err != nil {
				if err.Error() != "secret not found in keyring" {
					return err
				}
				break
			}
			count++
		}
	}
	return nil
}

// Helper functions
// Token is often too big to be stored in the keyring. Functions below split it in chunks of 1024 characters.
// Joining them is also handled.
func splitInChunks(token string, chunkSize int) []string {
	parts := make([]string, (len(token)+chunkSize-1)/chunkSize)
	for i := 0; i < len(parts); i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(token) {
			end = len(token)
		}
		parts[i] = token[start:end]
	}
	return parts
}

func joinChunks(parts []string) string {
	return strings.Join(parts, "")
}

func saveChunkedData(serviceName, key, value string) error {
	parts := splitInChunks(value, chunkSize)
	for i, part := range parts {
		chunkKey := fmt.Sprintf("%s_chunk_%d", key, i)
		err := keyring.Set(serviceName, chunkKey, part)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadChunkedData(serviceName, key string, numParts int) (string, error) {
	parts := make([]string, numParts)
	for i := 0; i < numParts; i++ {
		chunkKey := fmt.Sprintf("%s_chunk_%d", key, i)
		part, err := keyring.Get(serviceName, chunkKey)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve chunk %d for key %s: %w", i, key, err)
		}
		parts[i] = part
	}
	log.Debugf("Loaded %d chunks for key %s\n", numParts, key)
	return joinChunks(parts), nil
}
