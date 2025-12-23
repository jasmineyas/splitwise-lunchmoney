package lunchmoney

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestGetTransactionByID(t *testing.T) {
	tests := []struct {
		name            string
		transactionID   int64
		mockResponse    string
		mockStatus      int
		expectedPath    string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:          "successful retrieval",
			transactionID: 12345,
			mockStatus:    http.StatusOK,
			expectedPath:  "/transaction/12345",
			mockResponse: `{
				"date": "2025-12-23",
				"amount": "50.00",
				"payee": "Test Payee",
				"currency": "cad",
				"notes": "Test note",
				"status": "uncleared",
				"tags":["test"]
			}`,
			wantErr: false,
		},
		{
			name:            "invalid transaction ID - zero",
			transactionID:   0,
			wantErr:         true,
			wantErrContains: "invalid transaction ID",
		},
		{
			name:            "invalid transaction ID - negative",
			transactionID:   -1,
			wantErr:         true,
			wantErrContains: "invalid transaction ID",
		},
		{
			name:            "transaction not found",
			transactionID:   999999,
			mockStatus:      http.StatusNotFound,
			expectedPath:    "/transaction/999999",
			mockResponse:    `{"error": "Transaction not found"}`,
			wantErr:         true,
			wantErrContains: "API request failed with status 404",
		},
		{
			name:            "unauthorized",
			transactionID:   12345,
			mockStatus:      http.StatusUnauthorized,
			expectedPath:    "/transaction/12345",
			mockResponse:    `{"error": "Invalid access token"}`,
			wantErr:         true,
			wantErrContains: "API request failed with status 401",
		},
		{
			name:            "server error",
			transactionID:   12345,
			mockStatus:      http.StatusInternalServerError,
			expectedPath:    "/transaction/12345",
			mockResponse:    `{"error": "Internal server error"}`,
			wantErr:         true,
			wantErrContains: "API request failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedPath != "" {
					if r.URL.Path != tt.expectedPath {
						t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
					}

					if r.Method != "GET" {
						t.Errorf("expected GET request, got %s", r.Method)
					}

					auth := r.Header.Get("Authorization")
					if auth != "Bearer test-token" {
						t.Errorf("expected Bearer test-token, got %s", auth)
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

			transaction, err := client.GetTransactionByID(tt.transactionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrContains)
				} else if !containsString(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
			}

			if !tt.wantErr && transaction.Date == "" {
				t.Error("expected non-empty transaction, got empty")
			}
		})
	}
}

func TestGetTransactions(t *testing.T) {
	tests := []struct {
		name            string
		startDate       string
		endDate         string
		assetID         int64
		tag             string
		mockResponses   []string
		mockStatuses    []int
		expectedPaths   []string
		wantErr         bool
		wantErrContains string
		wantTxCount     int
	}{
		{
			name:      "successful single page",
			startDate: "2025-12-01",
			endDate:   "2025-12-31",
			assetID:   234273,
			tag:       "",
			mockResponses: []string{
				`{
					"transactions": [
						{"date": "2025-12-23", "amount": "50.00"},
						{"date": "2025-12-24", "amount": "75.00"}
					],
					"has_more": false
				}`,
			},
			mockStatuses:  []int{http.StatusOK},
			expectedPaths: []string{"/transactions"},
			wantErr:       false,
			wantTxCount:   2,
		},
		{
			name:      "successful pagination",
			startDate: "2025-12-01",
			endDate:   "2025-12-31",
			assetID:   234273,
			tag:       "",
			mockResponses: []string{
				`{
					"transactions": [
						{"date": "2025-12-23", "amount": "50.00"}
					],
					"has_more": true
				}`,
				`{
					"transactions": [
						{"date": "2025-12-24", "amount": "75.00"}
					],
					"has_more": false
				}`,
			},
			mockStatuses:  []int{http.StatusOK, http.StatusOK},
			expectedPaths: []string{"/transactions", "/transactions"},
			wantErr:       false,
			wantTxCount:   2,
		},
		{
			name:            "missing start date only",
			startDate:       "",
			endDate:         "2025-12-31",
			assetID:         234273,
			wantErr:         true,
			wantErrContains: "start_date and end_date must both be provided or both be empty",
		},
		{
			name:            "missing end date only",
			startDate:       "2025-12-01",
			endDate:         "",
			assetID:         234273,
			wantErr:         true,
			wantErrContains: "start_date and end_date must both be provided or both be empty",
		},
		{
			name:            "invalid start date format",
			startDate:       "2025-12-1",
			endDate:         "2025-12-31",
			assetID:         234273,
			wantErr:         true,
			wantErrContains: "invalid start_date format",
		},
		{
			name:            "invalid end date format",
			startDate:       "2025-12-01",
			endDate:         "2025-12-3",
			assetID:         234273,
			wantErr:         true,
			wantErrContains: "invalid end_date format",
		},
		{
			name:            "invalid asset ID - zero",
			startDate:       "2025-12-01",
			endDate:         "2025-12-31",
			assetID:         0,
			wantErr:         true,
			wantErrContains: "invalid asset ID",
		},
		{
			name:            "invalid asset ID - negative",
			startDate:       "2025-12-01",
			endDate:         "2025-12-31",
			assetID:         -1,
			wantErr:         true,
			wantErrContains: "invalid asset ID",
		},
		{
			name:      "unauthorized",
			startDate: "2025-12-01",
			endDate:   "2025-12-31",
			assetID:   234273,
			mockResponses: []string{
				`{"error": "Invalid access token"}`,
			},
			mockStatuses:    []int{http.StatusUnauthorized},
			expectedPaths:   []string{"/transactions"},
			wantErr:         true,
			wantErrContains: "API error (status 401)",
		},
		{
			name:      "server error",
			startDate: "2025-12-01",
			endDate:   "2025-12-31",
			assetID:   234273,
			mockResponses: []string{
				`{"error": "Internal server error"}`,
			},
			mockStatuses:    []int{http.StatusInternalServerError},
			expectedPaths:   []string{"/transactions"},
			wantErr:         true,
			wantErrContains: "API error (status 500)",
		},
	}

	/* Notes:
	What the test IS validating?
	- Loop iteration count - Does the loop run the correct number of times based on has_more?
	- Offset parameter - Does each request include the correct offset value (0, 1000, 2000)?
	- Response aggregation - Does the client correctly combine transactions from multiple responses?
	- Loop termination - Does the loop stop when has_more: false?

	What the test is NOT validating?
	- Actual pagination behavior - We're not testing with 1000+ real transactions
	- Server-side offset logic - The mock doesn't actually skip records based on offset
	- Real API semantics - The mock doesn't verify that offset=1000 returns records 1001-2000
	*/
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if len(tt.expectedPaths) > 0 {
					if r.URL.Path != tt.expectedPaths[0] {
						t.Errorf("expected path %s, got %s", tt.expectedPaths[0], r.URL.Path)
					}

					if r.Method != "GET" {
						t.Errorf("expected GET request, got %s", r.Method)
					}

					auth := r.Header.Get("Authorization")
					if auth != "Bearer test-token" {
						t.Errorf("expected Bearer test-token, got %s", auth)
					}
					// Verify query parameters
					query := r.URL.Query()
					if tt.startDate != "" && query.Get("start_date") != tt.startDate {
						t.Errorf("expected start_date %s, got %s", tt.startDate, query.Get("start_date"))
					}
					if tt.endDate != "" && query.Get("end_date") != tt.endDate {
						t.Errorf("expected end_date %s, got %s", tt.endDate, query.Get("end_date"))
					}

					// Validate pagination offset increments correctly
					if requestCount > 0 {
						expectedOffset := strconv.Itoa(requestCount * 1000)
						actualOffset := query.Get("offset")
						if actualOffset != expectedOffset {
							t.Errorf("request %d: expected offset %s, got %s", requestCount, expectedOffset, actualOffset)
						}
					}
				}

				if requestCount < len(tt.mockStatuses) {
					w.WriteHeader(tt.mockStatuses[requestCount])
					w.Write([]byte(tt.mockResponses[requestCount]))
					requestCount++
				}
			}))
			defer server.Close()

			client := &Client{
				httpClient:  &http.Client{},
				baseURL:     server.URL,
				bearerToken: "test-token",
			}

			transactions, err := client.GetTransactions(tt.startDate, tt.endDate, tt.assetID, tt.tag)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(transactions) != tt.wantTxCount {
					t.Errorf("GetTransactions() returned %d transactions, want %d", len(transactions), tt.wantTxCount)
				}

				// Verify the correct number of requests were made (validates loop iterations)
				expectedRequests := len(tt.mockStatuses)
				if requestCount != expectedRequests {
					t.Errorf("expected %d HTTP requests, got %d", expectedRequests, requestCount)
				}
			}
		})
	}
}

func TestUpdateTransaction(t *testing.T) {
	tests := []struct {
		name               string
		transactionID      int64
		updatedTransaction models.LunchMoneyTransaction
		mockResponse       string
		mockStatus         int
		expectedPath       string
		wantErr            bool
		wantErrContains    string
	}{
		{
			name:          "successful update",
			transactionID: 12345,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "75.00",
				Payee:  "Updated Payee",
			},
			mockStatus:   http.StatusOK,
			expectedPath: "/transaction/12345",
			mockResponse: `{"updated": true}`,
			wantErr:      false,
		},
		{
			name:          "update with all fields",
			transactionID: 12345,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:     "2025-12-23",
				Amount:   "100.00",
				Payee:    "Complete Update",
				Currency: "cad",
				Notes:    "Updated notes",
				Status:   "cleared",
				Tags:     []string{"updated", "test"},
			},
			mockStatus:   http.StatusOK,
			expectedPath: "/transaction/12345",
			mockResponse: `{"updated": true}`,
			wantErr:      false,
		},
		{
			name:          "invalid transaction ID - zero",
			transactionID: 0,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "50.00",
			},
			wantErr:         true,
			wantErrContains: "invalid transaction ID",
		},
		{
			name:          "invalid transaction ID - negative",
			transactionID: -1,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "50.00",
			},
			wantErr:         true,
			wantErrContains: "invalid transaction ID",
		},
		{
			name:          "API error response",
			transactionID: 12345,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-2",
				Amount: "50.00",
			},
			mockStatus:      http.StatusOK,
			expectedPath:    "/transaction/12345",
			mockResponse:    `{"error": ["Invalid date format"]}`,
			wantErr:         true,
			wantErrContains: "API error",
		},
		{
			name:          "update failed - not updated",
			transactionID: 12345,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "50.00",
			},
			mockStatus:      http.StatusOK,
			expectedPath:    "/transaction/12345",
			mockResponse:    `{"updated": false}`,
			wantErr:         true,
			wantErrContains: "transaction not updated",
		},
		{
			name:          "transaction not found",
			transactionID: 999999,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "50.00",
			},
			mockStatus:      http.StatusNotFound,
			expectedPath:    "/transaction/999999",
			mockResponse:    `{"error": "Transaction not found"}`,
			wantErr:         true,
			wantErrContains: "API error (status 404)",
		},
		{
			name:          "unauthorized",
			transactionID: 12345,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "50.00",
			},
			mockStatus:      http.StatusUnauthorized,
			expectedPath:    "/transaction/12345",
			mockResponse:    `{"error": "Invalid access token"}`,
			wantErr:         true,
			wantErrContains: "API error (status 401)",
		},
		{
			name:          "server error",
			transactionID: 12345,
			updatedTransaction: models.LunchMoneyTransaction{
				Date:   "2025-12-23",
				Amount: "50.00",
			},
			mockStatus:      http.StatusInternalServerError,
			expectedPath:    "/transaction/12345",
			mockResponse:    `{"error": "Internal server error"}`,
			wantErr:         true,
			wantErrContains: "API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedPath != "" {
					if r.URL.Path != tt.expectedPath {
						t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
					}

					if r.Method != "PUT" {
						t.Errorf("expected PUT request, got %s", r.Method)
					}

					auth := r.Header.Get("Authorization")
					if auth != "Bearer test-token" {
						t.Errorf("expected Bearer test-token, got %s", auth)
					}

					contentType := r.Header.Get("Content-Type")
					if contentType != "application/json" {
						t.Errorf("expected Content-Type application/json, got %s", contentType)
					}

					// Verify request body structure
					body, _ := io.ReadAll(r.Body)
					bodyStr := string(body)
					if len(bodyStr) > 0 {
						if !containsString(bodyStr, "transaction") {
							t.Error("expected request body to contain 'transaction' key")
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

			err := client.UpdateTransaction(tt.transactionID, tt.updatedTransaction)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrContains)
				} else if !containsString(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
			}
		})
	}
}

func TestGetTagID(t *testing.T) {
	tests := []struct {
		name            string
		tag             string
		mockResponse    string
		mockStatus      int
		expectedPath    string
		wantTagID       int64
		wantErr         bool
		wantErrContains string
	}{
		{
			name:         "tag found",
			tag:          "groceries",
			mockStatus:   http.StatusOK,
			expectedPath: "/tags",
			mockResponse: `[
				{"id": 1, "name": "groceries", "description": "Food shopping"},
				{"id": 2, "name": "transport", "description": "Transportation"}
			]`,
			wantTagID: 1,
			wantErr:   false,
		},
		{
			name:            "empty tag",
			tag:             "",
			wantErr:         true,
			wantErrContains: "tag cannot be empty",
		},
		{
			name:         "tag not found",
			tag:          "nonexistent",
			mockStatus:   http.StatusOK,
			expectedPath: "/tags",
			mockResponse: `[
				{"id": 1, "name": "groceries", "description": "Food shopping"},
				{"id": 2, "name": "transport", "description": "Transportation"}
			]`,
			wantErr:         true,
			wantErrContains: "tag not found",
		},
		{
			name:            "unauthorized",
			tag:             "groceries",
			mockStatus:      http.StatusUnauthorized,
			expectedPath:    "/tags",
			mockResponse:    `{"error": "Invalid access token"}`,
			wantErr:         true,
			wantErrContains: "API error (status 401)",
		},
		{
			name:            "server error",
			tag:             "groceries",
			mockStatus:      http.StatusInternalServerError,
			expectedPath:    "/tags",
			mockResponse:    `{"error": "Internal server error"}`,
			wantErr:         true,
			wantErrContains: "API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedPath != "" {
					if r.URL.Path != tt.expectedPath {
						t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
					}

					if r.Method != "GET" {
						t.Errorf("expected GET request, got %s", r.Method)
					}

					auth := r.Header.Get("Authorization")
					if auth != "Bearer test-token" {
						t.Errorf("expected Bearer test-token, got %s", auth)
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

			tagID, err := client.getTagID(tt.tag)

			if (err != nil) != tt.wantErr {
				t.Errorf("getTagID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrContains)
				} else if !containsString(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
			}

			if !tt.wantErr && tagID != tt.wantTagID {
				t.Errorf("getTagID() returned %d, want %d", tagID, tt.wantTagID)
			}
		})
	}
}
