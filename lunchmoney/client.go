package lunchmoney

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

func (c *Client) GetTransactionByID(transactionID int64) (models.LunchMoneyTransaction, error) {
	if transactionID <= 0 {
		return models.LunchMoneyTransaction{}, fmt.Errorf("invalid transaction ID: %d", transactionID)
	}

	req, err := c.newRequest("GET", "/transaction/"+fmt.Sprint(transactionID), nil)
	if err != nil {
		return models.LunchMoneyTransaction{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.LunchMoneyTransaction{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.LunchMoneyTransaction{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var transaction models.LunchMoneyTransaction
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		return models.LunchMoneyTransaction{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return transaction, nil
}

func (c *Client) GetTransactions(startDate string, endDate string, assetID int64, tag string) ([]models.LunchMoneyTransaction, error) {
	if tag != "" {
		_, err := c.getTagID(tag)
		if err != nil {
			return nil, fmt.Errorf("failed to get tag ID: %w", err)
		}
	}

	if (startDate == "") != (endDate == "") {
		return nil, fmt.Errorf("start_date and end_date must both be provided or both be empty")
	}

	if startDate != "" {
		_, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format (must be YYYY-MM-DD): %w", err)
		}
		_, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format (must be YYYY-MM-DD): %w", err)
		}
	}

	if assetID <= 0 {
		return nil, fmt.Errorf("invalid asset ID: %d", assetID)
	}

	// Build query parameters
	params := url.Values{}
	if startDate != "" {
		params.Add("start_date", startDate)
	}
	if endDate != "" {
		params.Add("end_date", endDate)
	}
	if assetID > 0 {
		params.Add("asset_id", strconv.FormatInt(assetID, 10))
	}
	if tag != "" {
		params.Add("tag_id", tag)
	}
	params.Add("debit_as_negative", "true")

	var allTransactions []models.LunchMoneyTransaction
	offset := 0
	hasMore := true

	var pageResp struct {
		Transactions []models.LunchMoneyTransaction `json:"transactions"`
		HasMore      bool                           `json:"has_more"`
	}

	for hasMore {
		// Add offset to params
		params.Set("offset", strconv.Itoa(offset))

		endpoint := "/transactions?" + params.Encode()

		req, err := c.newRequest("GET", endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request failed: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("performing request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
			return nil, fmt.Errorf("decoding response failed: %w", err)
		}

		allTransactions = append(allTransactions, pageResp.Transactions...)
		hasMore = pageResp.HasMore
		offset += 1000 // Default limit is 1000
	}

	return allTransactions, nil
}

func (c *Client) getTagID(tag string) (tagID int64, err error) {
	if tag == "" {
		return 0, fmt.Errorf("tag cannot be empty")
	}

	req, err := c.newRequest("GET", "/tags", nil)
	if err != nil {
		return 0, fmt.Errorf("creating request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("performing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tagsResp []models.LunchMoneyTag
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return 0, fmt.Errorf("decoding response failed: %w", err)
	}

	for _, t := range tagsResp {
		if t.Name == tag {
			return t.ID, nil
		}
	}

	return 0, fmt.Errorf("tag not found: %s", tag)
}

func (c *Client) UpdateTransaction(transactionID int64, updatedTransaction models.LunchMoneyTransaction) error {

	if transactionID <= 0 {
		return fmt.Errorf("invalid transaction ID: %d", transactionID)
	}

	requestBody := struct {
		Transaction     models.LunchMoneyTransaction `json:"transaction"`
		DebitAsNegative bool                         `json:"debit_as_negative"`
	}{
		Transaction:     updatedTransaction,
		DebitAsNegative: true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := c.newRequest("PUT", "/transaction/"+fmt.Sprint(transactionID), bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var transactionResp struct {
		Updated bool     `json:"updated"`
		Error   []string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&transactionResp); err != nil {
		return fmt.Errorf("decoding response failed: %w", err)
	}

	if len(transactionResp.Error) > 0 {
		return fmt.Errorf("API error: %v", transactionResp.Error)
	}

	if !transactionResp.Updated {
		return fmt.Errorf("transaction not updated")
	}

	return nil
}

func (c *Client) DeleteTransaction() error {
	// will come back to this later
	return nil
}
