// HTTP Client / API Consumer

package splitwise

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var userResp getCurrentUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return &userResp.User, nil
}

func (c *Client) GetAllExpenses(friendID int64, datedAfter string) ([]models.SplitwiseExpense, error) {
	params := url.Values{}

	if friendID > 0 {
		params.Add("friend_id", fmt.Sprintf("%d", friendID))
	}

	if datedAfter != "" {
		params.Add("dated_after", datedAfter)
	}

	endpoint := "/get_expenses"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var expensesResp getExpensesResponse
	if err := json.NewDecoder(resp.Body).Decode(&expensesResp); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return expensesResp.Expenses, nil
}

func (c *Client) GetExpenseByID(expenseID int64) (models.SplitwiseExpense, error) {
	if expenseID <= 0 {
		return models.SplitwiseExpense{}, fmt.Errorf("invalid expense ID: %d", expenseID)
	}

	endpoint := fmt.Sprintf("/get_expense?id=%d", expenseID)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return models.SplitwiseExpense{}, fmt.Errorf("creating request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.SplitwiseExpense{}, fmt.Errorf("performing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.SplitwiseExpense{}, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var expenseResp struct {
		Expense models.SplitwiseExpense `json:"expense"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&expenseResp); err != nil {
		return models.SplitwiseExpense{}, fmt.Errorf("decoding response failed: %w", err)
	}
	return expenseResp.Expense, nil
}

func (c *Client) GetExpenseComments(expenseID int64) ([]models.SplitwiseComment, error) {
	if expenseID <= 0 {
		return nil, fmt.Errorf("invalid expense ID: %d", expenseID)
	}

	endpoint := fmt.Sprintf("/get_comments?expense_id=%d", expenseID)

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

	var commentsResp struct {
		Comments []models.SplitwiseComment `json:"comments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commentsResp); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return commentsResp.Comments, nil
}

func (c *Client) AddCommentToExpense(expenseID int64, comment string) error {
	if expenseID <= 0 {
		return fmt.Errorf("invalid expense ID: %d", expenseID)
	}

	if comment == "" {
		return fmt.Errorf("comment cannot be empty")
	}

	endpoint := fmt.Sprintf("/create_comment?expense_id=%d&content=%s", expenseID, url.PathEscape(comment))
	fmt.Println(endpoint)

	req, err := c.newRequest("POST", endpoint, nil)
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

	return nil
}

func (c *Client) DeleteComment(commentID int64) error {
	if commentID <= 0 {
		return fmt.Errorf("invalid comment ID: %d", commentID)
	}

	endpoint := fmt.Sprintf("/delete_comment/%d", commentID)

	req, err := c.newRequest("POST", endpoint, nil)
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

	return nil
}
