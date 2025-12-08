package model

type LunchMoneyTransaction struct {
    Date        string   `json:"date"`
    Amount      string   `json:"amount"`
    Payee       string   `json:"payee"`
    Currency    string   `json:"currency"`
    AssetID     int64    `json:"asset_id,omitempty"`
    Notes       string   `json:"notes"`
    Status      string   `json:"status"`
    ExternalID  string   `json:"external_id"`
    Tags        []string `json:"tags"`
}