package lunchmoney

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	bearerToken string
}

func NewClient(bearerToken string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 60 * time.Second},
		baseURL:     "https://dev.lunchmoney.app/v1",
		bearerToken: bearerToken,
	}
}

func (c *Client) newRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	url := c.baseURL + endpoint

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) VerifyAssetExist(assetID int64) (bool, error) {
	if assetID <= 0 {
		return false, fmt.Errorf("invalid asset ID: %d", assetID)
	}

	req, err := c.newRequest("GET", "/assets", nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Assets []struct {
			ID int64 `json:"id"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, asset := range response.Assets {
		if asset.ID == assetID {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) AddTransaction() error {
	return nil
}

func (c *Client) GetTransactions() error {
	return nil
}

func (c *Client) GetTransactionByID() error {
	return nil
}

func (c *Client) UpdateTransaction() error {
	return nil
}

func (c *Client) DeleteTransaction() error {
	return nil
}
