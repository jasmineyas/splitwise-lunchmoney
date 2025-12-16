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
			userInfo, err := client.GetUserInfo()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify results if no error expected
			if !tt.expectError {
				if userInfo.UserID != tt.expectedUserID {
					t.Errorf("Expected UserID %d, got %d", tt.expectedUserID, userInfo.UserID)
				}
				if userInfo.FirstName != tt.expectedFirst {
					t.Errorf("Expected FirstName '%s', got '%s'", tt.expectedFirst, userInfo.FirstName)
				}
				if userInfo.LastName != tt.expectedLast {
					t.Errorf("Expected LastName '%s', got '%s'", tt.expectedLast, userInfo.LastName)
				}
			}
		})
	}
}
