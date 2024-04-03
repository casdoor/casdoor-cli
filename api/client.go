/*
 * Copyright (c) 2024 Fabien CHEVALIER
 * All rights reserved.
 */

package api

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"gitlab.com/sdv9972401/casdoor-cli/utils"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL     string
	Token       string
	ContentType string
}

// NewClient creates a new Client with the provided baseURL and token.
//
// Parameters:
// - baseURL: a string representing the base URL
// - token: a string representing the authentication token
// Returns a pointer to a Client
func NewClient(baseURL string, token string) *Client {
	return &Client{
		BaseURL:     baseURL,
		Token:       token,
		ContentType: "application/json",
	}
}

// sendRequest sends a request to the specified endpoint using the given HTTP method and data.
// It returns the response body and an error, if any.
func (c *Client) sendRequest(method, endpoint string, data []byte) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", c.ContentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(resp.Body)
	log.Debug(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("response status code is not successful: %d", resp.StatusCode)
	}
	log.Debug(string(body))
	if strings.Contains(string(body), "Access token has expired") {
		utils.Colorize(color.RedString, "[x] access token has expired. Please log in again.")
		return nil, fmt.Errorf("error : %s", string(body))
	} else if strings.Contains(string(body), "Unauthorized operation") {
		utils.Colorize(color.RedString, "[x] unauthorized operation. You may need administrator privileges")
		return nil, fmt.Errorf("error : %s", string(body))
	}
	return body, nil
}

// Get retrieves data from the specified endpoint.
//
// It takes a string endpoint as a parameter and returns a byte array and an error.
func (c *Client) Get(endpoint string) ([]byte, error) {
	return c.sendRequest(http.MethodGet, endpoint, nil)
}

// Post sends a POST request to the specified endpoint with the provided data.
//
// endpoint string, data []byte
// []byte, error
func (c *Client) Post(endpoint string, data []byte) ([]byte, error) {
	return c.sendRequest(http.MethodPost, endpoint, data)
}
