package lunchmoney

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jasmineyas/splitwise-lunchmoney/models"
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

func (c *Client) AddTransactions(transactions []models.LunchMoneyTransaction) (transactionIDs []string, err error) {
	if len(transactions) == 0 {
		return []string{}, fmt.Errorf("no transactions to add")
	}

	// Validate required fields for each transaction
	for i, tx := range transactions {
		if tx.Date == "" {
			return []string{}, fmt.Errorf("transaction[%d]: date is required", i)
		}
		if tx.Amount == "" {
			return []string{}, fmt.Errorf("transaction[%d]: amount is required", i)
		}
	}

	requestBody := struct {
		Transactions    []models.LunchMoneyTransaction `json:"transactions"`
		DebitAsNegative bool                           `json:"debit_as_negative"`
	}{
		Transactions:    transactions,
		DebitAsNegative: true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return []string{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := c.newRequest("POST", "/transactions", bytes.NewReader(jsonData))
	if err != nil {
		return []string{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()

	// seems like lunchmoney returns 200 even on errors right now, even though the doc says there would be 404 response
	// https://lunchmoney.dev/#insert-transactions

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return []string{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var responseBody struct {
		IDs   []string `json:"ids"`
		Error []string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return []string{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responseBody.Error) > 0 {
		return []string{}, fmt.Errorf("API error: %v", responseBody.Error)
	}

	return responseBody.IDs, nil
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
