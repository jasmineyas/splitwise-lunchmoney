package splitwise

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserInfo(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedUserID int64
		expectedFirst  string
		expectedLast   string
		expectError    bool
	}{
		{
			name:       "success with full name",
			statusCode: 200,
			responseBody: `{
				"user": {
					"id": 9792490,
					"first_name": "Jasmine",
					"last_name": "Zou"
				}
			}`,
			expectedUserID: 9792490,
			expectedFirst:  "Jasmine",
			expectedLast:   "Zou",
			expectError:    false,
		},
		{
			name:       "success with null last name",
			statusCode: 200,
			responseBody: `{
				"user": {
					"id": 9792490,
					"first_name": "Jasmine",
					"last_name": null
				}
			}`,
			expectedUserID: 9792490,
			expectedFirst:  "Jasmine",
			expectedLast:   "",
			expectError:    false,
		},
		{
			name:         "unauthorized",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
		{
			name:         "server error",
			statusCode:   500,
			responseBody: `{"error": "Internal server error"}`,
			expectError:  true,
		},
		{
			name:         "invalid json",
			statusCode:   200,
			responseBody: `{invalid json}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server for this test case
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path
				if r.URL.Path != "/get_current_user" {
					t.Errorf("Expected path /get_current_user, got %s", r.URL.Path)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewClient("test-token")
			client.baseURL = server.URL

			// Call the function
			user, err := client.GetUserInfo()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify results if no error expected
			if !tt.expectError {
				if user.ID != tt.expectedUserID {
					t.Errorf("Expected ID %d, got %d", tt.expectedUserID, user.ID)
				}
				if user.FirstName != tt.expectedFirst {
					t.Errorf("Expected FirstName '%s', got '%s'", tt.expectedFirst, user.FirstName)
				}
				if user.LastName != tt.expectedLast {
					t.Errorf("Expected LastName '%s', got '%s'", tt.expectedLast, user.LastName)
				}
			}
		})
	}
}

func TestGetAllExpenses(t *testing.T) {
	tests := []struct {
		name            string
		friendID        int64
		datedAfter      string
		expectedPath    string
		statusCode      int
		responseBody    string
		expectedCount   int
		expectedFirstID int64
		expectError     bool
	}{
		{
			name:         "success with expenses - no filters",
			friendID:     0,
			datedAfter:   "",
			expectedPath: "/get_expenses",
			statusCode:   200,
			responseBody: `{
				"expenses": [
					{
						"id": 4198563142,
						"group_id": null,
						"description": "Socks and underwear",
						"cost": "310.6",
						"currency_code": "CNY",
						"date": "2025-12-11T00:38:14Z",
						"deleted_at": null,
						"deleted_by": null,
						"repayments": [
							{
								"from": 50086667,
								"to": 9792490,
								"amount": "310.6"
							}
						],
						"users": [
							{
								"user_id": 50086667,
								"paid_share": "0.0",
								"owed_share": "310.6",
								"net_balance": "-310.6"
							},
							{
								"user_id": 9792490,
								"paid_share": "310.6",
								"owed_share": "0.0",
								"net_balance": "310.6"
							}
						]
					},
					{
						"id": 4198563143,
						"description": "Dinner",
						"cost": "50.0",
						"currency_code": "USD",
						"date": "2025-12-12T00:00:00Z",
						"deleted_at": null,
						"deleted_by": null,
						"repayments": [],
						"users": []
					}
				]
			}`,
			expectedCount:   2,
			expectedFirstID: 4198563142,
			expectError:     false,
		},
		{
			name:         "success with friend filter",
			friendID:     50086667,
			datedAfter:   "",
			expectedPath: "/get_expenses?friend_id=50086667",
			statusCode:   200,
			responseBody: `{
				"expenses": [
					{
						"id": 4198563142,
						"description": "Socks and underwear",
						"cost": "310.6",
						"currency_code": "CNY",
						"date": "2025-12-11T00:38:14Z",
						"deleted_at": null,
						"deleted_by": null,
						"repayments": [],
						"users": []
					}
				]
			}`,
			expectedCount:   1,
			expectedFirstID: 4198563142,
			expectError:     false,
		},
		{
			name:          "success with empty expenses",
			friendID:      0,
			datedAfter:    "",
			expectedPath:  "/get_expenses",
			statusCode:    200,
			responseBody:  `{"expenses": []}`,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:         "success with dated_after filter",
			friendID:     0,
			datedAfter:   "2025-12-01T00:00:00Z",
			expectedPath: "/get_expenses?dated_after=2025-12-01T00%3A00%3A00Z",
			statusCode:   200,
			responseBody: `{
				"expenses": [
					{
						"id": 4198563142,
						"description": "Recent expense",
						"cost": "100.0",
						"currency_code": "USD",
						"date": "2025-12-15T00:00:00Z",
						"deleted_at": null,
						"deleted_by": null,
						"repayments": [],
						"users": []
					}
				]
			}`,
			expectedCount:   1,
			expectedFirstID: 4198563142,
			expectError:     false,
		},
		{
			name:         "success with both friend and dated_after filters",
			friendID:     50086667,
			datedAfter:   "2025-12-01T00:00:00Z",
			expectedPath: "/get_expenses?dated_after=2025-12-01T00%3A00%3A00Z&friend_id=50086667",
			statusCode:   200,
			responseBody: `{
				"expenses": [
					{
						"id": 4198563142,
						"description": "Friend recent expense",
						"cost": "50.0",
						"currency_code": "USD",
						"date": "2025-12-10T00:00:00Z",
						"deleted_at": null,
						"deleted_by": null,
						"repayments": [],
						"users": []
					}
				]
			}`,
			expectedCount:   1,
			expectedFirstID: 4198563142,
			expectError:     false,
		},
		{
			name:         "unauthorized",
			friendID:     0,
			datedAfter:   "",
			expectedPath: "/get_expenses",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
		{
			name:         "server error",
			friendID:     0,
			datedAfter:   "",
			expectedPath: "/get_expenses",
			statusCode:   500,
			responseBody: `{"error": "Internal server error"}`,
			expectError:  true,
		},
		{
			name:         "invalid json",
			friendID:     0,
			datedAfter:   "",
			expectedPath: "/get_expenses",
			statusCode:   200,
			responseBody: `{invalid json}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server for this test case
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path (including query parameters)
				fullPath := r.URL.Path
				if r.URL.RawQuery != "" {
					fullPath += "?" + r.URL.RawQuery
				}
				if fullPath != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, fullPath)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewClient("test-token")
			client.baseURL = server.URL

			// Call the function
			expenses, err := client.GetAllExpenses(tt.friendID, tt.datedAfter)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify results if no error expected
			if !tt.expectError {
				if len(expenses) != tt.expectedCount {
					t.Errorf("Expected %d expenses, got %d", tt.expectedCount, len(expenses))
				}
				if tt.expectedCount > 0 && expenses[0].ID != tt.expectedFirstID {
					t.Errorf("Expected first expense ID %d, got %d", tt.expectedFirstID, expenses[0].ID)
				}
			}
		})
	}
}

func TestGetExpense(t *testing.T) {
	tests := []struct {
		name              string
		expenseID         int64
		expectedPath      string
		statusCode        int
		responseBody      string
		expectedExpenseID int64
		expectedDesc      string
		expectError       bool
	}{
		{
			name:         "success",
			expenseID:    4198563142,
			expectedPath: "/get_expense?id=4198563142",
			statusCode:   200,
			responseBody: `{
				"expense": {
					"id": 4198563142,
					"description": "Socks and underwear",
					"cost": "310.6",
					"currency_code": "CNY",
					"date": "2025-12-11T00:38:14Z",
					"deleted_at": null,
					"deleted_by": null,
					"repayments": [
						{
							"from": 50086667,
							"to": 9792490,
							"amount": "310.6"
						}
					],
					"users": [
						{
							"user_id": 50086667,
							"paid_share": "0.0",
							"owed_share": "310.6",
							"net_balance": "-310.6"
						},
						{
							"user_id": 9792490,
							"paid_share": "310.6",
							"owed_share": "0.0",
							"net_balance": "310.6"
						}
					]
				}
			}`,
			expectedExpenseID: 4198563142,
			expectedDesc:      "Socks and underwear",
			expectError:       false,
		},
		{
			name:        "invalid expense ID - zero",
			expenseID:   0,
			expectError: true,
		},
		{
			name:        "invalid expense ID - negative",
			expenseID:   -1,
			expectError: true,
		},
		{
			name:         "not found",
			expenseID:    999999999,
			expectedPath: "/get_expense?id=999999999",
			statusCode:   404,
			responseBody: `{"error": "Expense not found"}`,
			expectError:  true,
		},
		{
			name:         "unauthorized",
			expenseID:    123456,
			expectedPath: "/get_expense?id=123456",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
		{
			name:         "invalid json",
			expenseID:    123456,
			expectedPath: "/get_expense?id=123456",
			statusCode:   200,
			responseBody: `{invalid json}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip server setup for invalid ID tests
			if tt.expenseID <= 0 {
				client := NewClient("test-token")
				_, err := client.GetExpenseByID(tt.expenseID)
				if err == nil {
					t.Error("Expected error for invalid expense ID but got none")
				}
				return
			}

			// Create mock server for valid ID tests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path (including query parameters)
				fullPath := r.URL.Path
				if r.URL.RawQuery != "" {
					fullPath += "?" + r.URL.RawQuery
				}
				if fullPath != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, fullPath)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewClient("test-token")
			client.baseURL = server.URL

			// Call the function
			expense, err := client.GetExpenseByID(tt.expenseID)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify results if no error expected
			if !tt.expectError {
				if expense.ID != tt.expectedExpenseID {
					t.Errorf("Expected expense ID %d, got %d", tt.expectedExpenseID, expense.ID)
				}
				if expense.Description != tt.expectedDesc {
					t.Errorf("Expected description '%s', got '%s'", tt.expectedDesc, expense.Description)
				}
			}
		})
	}
}

func TestGetExpenseComments(t *testing.T) {
	tests := []struct {
		name          string
		expenseID     int64
		expectedPath  string
		statusCode    int
		responseBody  string
		expectedCount int
		expectError   bool
	}{
		{
			name:         "success with comments",
			expenseID:    4198563142,
			expectedPath: "/get_comments?expense_id=4198563142",
			statusCode:   200,
			responseBody: `{
				"comments": [
					{
						"id": 123,
						"content": "synced-to-LM",
						"comment_type": "System",
						"relation_type": "Expense",
						"relation_id": 4198563142,
						"created_at": "2025-12-11T00:38:40Z",
						"deleted_at": null,
						"user": {
							"id": 9792490,
							"first_name": "Jasmine",
							"last_name": null
						}
					},
					{
						"id": 124,
						"content": "LM transaction ID: 12345",
						"comment_type": "System",
						"relation_type": "Expense",
						"relation_id": 4198563142,
						"created_at": "2025-12-11T00:38:45Z",
						"deleted_at": null,
						"user": {
							"id": 9792490,
							"first_name": "Jasmine",
							"last_name": null
						}
					}
				]
			}`,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "success with no comments",
			expenseID:     4198563142,
			expectedPath:  "/get_comments?expense_id=4198563142",
			statusCode:    200,
			responseBody:  `{"comments": []}`,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:        "invalid expense ID - zero",
			expenseID:   0,
			expectError: true,
		},
		{
			name:        "invalid expense ID - negative",
			expenseID:   -1,
			expectError: true,
		},
		{
			name:         "not found",
			expenseID:    999999999,
			expectedPath: "/get_comments?expense_id=999999999",
			statusCode:   404,
			responseBody: `{"error": "Expense not found"}`,
			expectError:  true,
		},
		{
			name:         "unauthorized",
			expenseID:    123456,
			expectedPath: "/get_comments?expense_id=123456",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip server setup for invalid ID tests
			if tt.expenseID <= 0 {
				client := NewClient("test-token")
				_, err := client.GetExpenseComments(tt.expenseID)
				if err == nil {
					t.Error("Expected error for invalid expense ID but got none")
				}
				return
			}

			// Create mock server for valid ID tests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path (including query parameters)
				fullPath := r.URL.Path
				if r.URL.RawQuery != "" {
					fullPath += "?" + r.URL.RawQuery
				}
				if fullPath != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, fullPath)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewClient("test-token")
			client.baseURL = server.URL

			// Call the function
			comments, err := client.GetExpenseComments(tt.expenseID)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify results if no error expected
			if !tt.expectError {
				if len(comments) != tt.expectedCount {
					t.Errorf("Expected %d comments, got %d", tt.expectedCount, len(comments))
				}
			}
		})
	}
}

func TestAddCommentToExpense(t *testing.T) {
	tests := []struct {
		name         string
		expenseID    int64
		comment      string
		expectedPath string
		statusCode   int
		responseBody string
		expectError  bool
	}{
		{
			name:         "success",
			expenseID:    4198563142,
			comment:      "synced-to-LM",
			expectedPath: "/create_comment?expense_id=4198563142&content=synced-to-LM",
			statusCode:   200,
			responseBody: `{"success": true}`,
			expectError:  false,
		},
		{
			name:         "success with special characters",
			expenseID:    4198563142,
			comment:      "LM ID: 12345",
			expectedPath: "/create_comment?expense_id=4198563142&content=LM%20ID:%2012345",
			statusCode:   200,
			responseBody: `{"success": true}`,
			expectError:  false,
		},
		{
			name:        "invalid expense ID - zero",
			expenseID:   0,
			comment:     "test",
			expectError: true,
		},
		{
			name:        "invalid expense ID - negative",
			expenseID:   -1,
			comment:     "test",
			expectError: true,
		},
		{
			name:        "empty comment",
			expenseID:   4198563142,
			comment:     "",
			expectError: true,
		},
		{
			name:         "expense not found",
			expenseID:    999999999,
			comment:      "test",
			expectedPath: "/create_comment?expense_id=999999999&content=test",
			statusCode:   404,
			responseBody: `{"error": "Expense not found"}`,
			expectError:  true,
		},
		{
			name:         "unauthorized",
			expenseID:    123456,
			comment:      "test",
			expectedPath: "/create_comment?expense_id=123456&content=test",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip server setup for validation error tests
			if tt.expenseID <= 0 || tt.comment == "" {
				client := NewClient("test-token")
				err := client.AddCommentToExpense(tt.expenseID, tt.comment)
				if err == nil {
					t.Error("Expected error for invalid input but got none")
				}
				return
			}

			// Create mock server for valid input tests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				// Verify request path (including query parameters)
				fullPath := r.URL.Path
				if r.URL.RawQuery != "" {
					fullPath += "?" + r.URL.RawQuery
				}
				if fullPath != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, fullPath)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewClient("test-token")
			client.baseURL = server.URL

			// Call the function
			err := client.AddCommentToExpense(tt.expenseID, tt.comment)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	tests := []struct {
		name         string
		commentID    int64
		expectedPath string
		statusCode   int
		responseBody string
		expectError  bool
	}{
		{
			name:         "success",
			commentID:    123456,
			expectedPath: "/delete_comment/123456",
			statusCode:   200,
			responseBody: `{"success": true}`,
			expectError:  false,
		},
		{
			name:        "invalid comment ID - zero",
			commentID:   0,
			expectError: true,
		},
		{
			name:        "invalid comment ID - negative",
			commentID:   -1,
			expectError: true,
		},
		{
			name:         "comment not found",
			commentID:    999999999,
			expectedPath: "/delete_comment/999999999",
			statusCode:   404,
			responseBody: `{"error": "Comment not found"}`,
			expectError:  true,
		},
		{
			name:         "unauthorized",
			commentID:    123456,
			expectedPath: "/delete_comment/123456",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
		{
			name:         "server error",
			commentID:    123456,
			expectedPath: "/delete_comment/123456",
			statusCode:   500,
			responseBody: `{"error": "Internal server error"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip server setup for validation error tests
			if tt.commentID <= 0 {
				client := NewClient("test-token")
				err := client.DeleteComment(tt.commentID)
				if err == nil {
					t.Error("Expected error for invalid comment ID but got none")
				}
				return
			}

			// Create mock server for valid input tests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				// Verify request path
				if r.URL.Path != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client pointing to mock server
			client := NewClient("test-token")
			client.baseURL = server.URL

			// Call the function
			err := client.DeleteComment(tt.commentID)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
