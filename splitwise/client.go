// HTTP Client / API Consumer

package splitwise

import (
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

type getCurrentUserResponse struct {
	User models.User `json:"user"`
}

type getExpensesResponse struct {
	Expenses []models.SplitwiseExpense `json:"expenses"`
}

func NewClient(bearerToken string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 60 * time.Second},
		baseURL:     "https://secure.splitwise.com/api/v3.0",
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

func (c *Client) GetUserInfo() (*models.User, error) {
	req, err := c.newRequest("GET", "/get_current_user", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var userResp getCurrentUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return &userResp.User, nil
}

func (c *Client) GetAllExpenses(friendID int64) ([]models.SplitwiseExpense, error) {
	endpoint := "/get_expenses"
	if friendID > 0 {
		endpoint = fmt.Sprintf("/get_expenses?friend_id=%d", friendID)
	}

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
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var expensesResp getExpensesResponse
	if err := json.NewDecoder(resp.Body).Decode(&expensesResp); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return expensesResp.Expenses, nil
}

func (c *Client) GetExpense(expenseID int64) (models.SplitwiseExpense, error) {
	return models.SplitwiseExpense{}, nil
}

func (c *Client) AddCommentToExpense(expenseID int64, comment string) error {
	return nil
}

func (c *Client) DeleteCommentFromExpense(expenseID int64, commentID int64) error {
	return nil
}
