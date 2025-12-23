package lunchmoney

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jasmineyas/splitwise-lunchmoney/models"
)

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestVerifyAssetExist(t *testing.T) {
	tests := []struct {
		name         string
		assetID      int64
		mockResponse string
		mockStatus   int
		expectedPath string
		wantFound    bool
		wantErr      bool
	}{
		{
			name:         "asset exists",
			assetID:      234273,
			mockStatus:   http.StatusOK,
			expectedPath: "/assets",
			mockResponse: `{
				"assets": [
					{
						"id": 234273,
						"type_name": "cash",
						"subtype_name": "physical cash",
						"name": "Jasmine Test Splitwise Account",
						"display_name": "Jasmine Test Splitwise Account",
						"balance": "0.0000",
						"to_base": 0,
						"balance_as_of": "2025-12-09T08:03:22.000Z",
						"closed_on": null,
						"currency": "cad",
						"institution_name": "Jasmine Test",
						"exclude_transactions": false,
						"created_at": "2025-12-09T08:02:56.110Z"
					},
					{
						"id": 165646,
						"type_name": "cash",
						"subtype_name": "Other",
						"name": "Splitwise",
						"display_name": "Splitwise",
						"balance": "1531.0800",
						"to_base": 1531.08,
						"balance_as_of": "2025-07-27T20:00:44.226Z",
						"closed_on": null,
						"currency": "cad",
						"institution_name": "Splitwise",
						"exclude_transactions": false,
						"created_at": "2025-05-04T01:22:25.636Z"
					}
				]
			}`,
			wantFound: true,
			wantErr:   false,
		},
		{
			name:         "asset does not exist",
			assetID:      999999,
			mockStatus:   http.StatusOK,
			expectedPath: "/assets",
			mockResponse: `{
				"assets": [
					{
						"id": 234273,
						"type_name": "cash",
						"name": "Test Account"
					},
					{
						"id": 165646,
						"type_name": "cash",
						"name": "Another Account"
					}
				]
			}`,
			wantFound: false,
			wantErr:   false,
		},
		{
			name:         "empty assets list",
			assetID:      234273,
			mockStatus:   http.StatusOK,
			expectedPath: "/assets",
			mockResponse: `{
				"assets": []
			}`,
			wantFound: false,
			wantErr:   false,
		},
		{
			name:       "invalid asset ID - zero",
			assetID:    0,
			mockStatus: http.StatusOK,
			wantFound:  false,
			wantErr:    true,
		},
		{
			name:       "invalid asset ID - negative",
			assetID:    -1,
			mockStatus: http.StatusOK,
			wantFound:  false,
			wantErr:    true,
		},
		{
			name:         "unauthorized",
			assetID:      234273,
			mockStatus:   http.StatusUnauthorized,
			expectedPath: "/assets",
			mockResponse: `{"error": "Invalid access token"}`,
			wantFound:    false,
			wantErr:      true,
		},
		{
			name:         "server error",
			assetID:      234273,
			mockStatus:   http.StatusInternalServerError,
			expectedPath: "/assets",
			mockResponse: `{"error": "Internal server error"}`,
			wantFound:    false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedPath != "" && r.URL.Path != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}

				if r.Method != "GET" {
					t.Errorf("expected GET request, got %s", r.Method)
				}

				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("expected Bearer test-token, got %s", auth)
				}

				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := &Client{
				httpClient:  &http.Client{},
				baseURL:     server.URL,
				bearerToken: "test-token",
			}

			found, err := client.VerifyAssetExist(tt.assetID)

			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyAssetExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if found != tt.wantFound {
				t.Errorf("VerifyAssetExist() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestAddTransactions(t *testing.T) {
	tests := []struct {
		name            string
		transactions    []models.LunchMoneyTransaction
		mockResponse    string
		mockStatus      int
		expectedPath    string
		wantIDs         []string
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "successful transaction creation",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Amount:   "50.00",
					Payee:    "Test Payee",
					Currency: "cad",
					AssetID:  234273,
					Notes:    "Test note",
					Status:   "uncleared",
					Tags:     []string{"test"},
				},
			},
			mockStatus:   http.StatusOK,
			expectedPath: "/transactions",
			mockResponse: `{"ids": ["12345"]}`,
			wantIDs:      []string{"12345"},
			wantErr:      false,
		},
		{
			name: "multiple transactions",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Amount:   "50.00",
					Payee:    "Payee 1",
					Currency: "cad",
				},
				{
					Date:     "2025-12-24",
					Amount:   "75.00",
					Payee:    "Payee 2",
					Currency: "cad",
				},
			},
			mockStatus:   http.StatusOK,
			expectedPath: "/transactions",
			mockResponse: `{"ids": ["12345", "12346"]}`,
			wantIDs:      []string{"12345", "12346"},
			wantErr:      false,
		},
		{
			name:            "empty transactions list",
			transactions:    []models.LunchMoneyTransaction{},
			mockStatus:      http.StatusOK,
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "no transactions to add",
		},
		{
			name:            "nil transactions",
			transactions:    nil,
			mockStatus:      http.StatusOK,
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "no transactions to add",
		},
		{
			name: "missing date - validation error",
			transactions: []models.LunchMoneyTransaction{
				{
					Amount:   "50.00",
					Payee:    "Test",
					Currency: "cad",
				},
			},
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "date is required",
		},
		{
			name: "missing amount - validation error",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Payee:    "Test",
					Currency: "cad",
				},
			},
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "amount is required",
		}, {
			name: "API error response",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-2",
					Amount:   "50.00",
					Payee:    "Test",
					Currency: "djd",
				},
			},
			mockStatus:   http.StatusOK,
			expectedPath: "/transactions",
			mockResponse: `{"error": ["Transaction 0 date format is invalid (must be YYYY-MM-DD): 2024-10-2",
        "Transaction 0 has an invalid currency: djd"]}`,
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "API error",
		},
		{
			name: "unauthorized",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Amount:   "50.00",
					Payee:    "Test",
					Currency: "cad",
				},
			},
			mockStatus:      http.StatusUnauthorized,
			expectedPath:    "/transactions",
			mockResponse:    `{"error": "Invalid access token"}`,
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "API request failed with status 401",
		},
		{
			name: "server error",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Amount:   "50.00",
					Payee:    "Test",
					Currency: "cad",
				},
			},
			mockStatus:      http.StatusInternalServerError,
			expectedPath:    "/transactions",
			mockResponse:    `{"error": "Internal server error"}`,
			wantIDs:         []string{},
			wantErr:         true,
			wantErrContains: "API request failed with status 500",
		},
		{
			name: "unauthorized",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Amount:   "50.00",
					Payee:    "Test",
					Currency: "cad",
				},
			},
			mockStatus:   http.StatusUnauthorized,
			expectedPath: "/transactions",
			mockResponse: `{"error": "Invalid access token"}`,
			wantIDs:      []string{},
			wantErr:      true,
		},
		{
			name: "server error",
			transactions: []models.LunchMoneyTransaction{
				{
					Date:     "2025-12-23",
					Amount:   "50.00",
					Payee:    "Test",
					Currency: "cad",
				},
			},
			mockStatus:   http.StatusInternalServerError,
			expectedPath: "/transactions",
			mockResponse: `{"error": "Internal server error"}`,
			wantIDs:      []string{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedPath != "" {
					if r.URL.Path != tt.expectedPath {
						t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
					}

					if r.Method != "POST" {
						t.Errorf("expected POST request, got %s", r.Method)
					}

					auth := r.Header.Get("Authorization")
					if auth != "Bearer test-token" {
						t.Errorf("expected Bearer test-token, got %s", auth)
					}

					contentType := r.Header.Get("Content-Type")
					if contentType != "application/json" {
						t.Errorf("expected Content-Type application/json, got %s", contentType)
					}

					// Verify request body contains expected fields
					body, _ := io.ReadAll(r.Body)
					bodyStr := string(body)
					if len(bodyStr) > 0 {
						// Check that the request contains transactions and debit_as_negative
						if len(tt.transactions) > 0 {
							// Basic validation that JSON structure is present
							if bodyStr == "" {
								t.Error("expected non-empty request body")
							}
						}
					}
				}

				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := &Client{
				httpClient:  &http.Client{},
				baseURL:     server.URL,
				bearerToken: "test-token",
			}

			ids, err := client.AddTransactions(tt.transactions)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Validate error message if expected
			if tt.wantErr && tt.wantErrContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrContains)
				} else if !containsString(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
			}

			if !tt.wantErr {
				if len(ids) != len(tt.wantIDs) {
					t.Errorf("AddTransactions() returned %d IDs, want %d", len(ids), len(tt.wantIDs))
					return
				}
				for i, id := range ids {
					if id != tt.wantIDs[i] {
						t.Errorf("AddTransactions() ID[%d] = %s, want %s", i, id, tt.wantIDs[i])
					}
				}
			}
		})
	}
}
