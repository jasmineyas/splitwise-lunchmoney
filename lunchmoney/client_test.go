package lunchmoney

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
