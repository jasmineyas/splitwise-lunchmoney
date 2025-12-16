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
		expectedPath    string
		statusCode      int
		responseBody    string
		expectedCount   int
		expectedFirstID int64
		expectError     bool
	}{
		{
			name:         "success with expenses - no friend filter",
			friendID:     0,
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
			expectedPath:  "/get_expenses",
			statusCode:    200,
			responseBody:  `{"expenses": []}`,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:         "unauthorized",
			friendID:     0,
			expectedPath: "/get_expenses",
			statusCode:   401,
			responseBody: `{"error": "Invalid token"}`,
			expectError:  true,
		},
		{
			name:         "server error",
			friendID:     0,
			expectedPath: "/get_expenses",
			statusCode:   500,
			responseBody: `{"error": "Internal server error"}`,
			expectError:  true,
		},
		{
			name:         "invalid json",
			friendID:     0,
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
			expenses, err := client.GetAllExpenses(tt.friendID)

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
