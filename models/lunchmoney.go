package models

type LunchMoneyTransaction struct {
	Date     string   `json:"date"`
	Amount   string   `json:"amount"`
	Payee    string   `json:"payee"`
	Currency string   `json:"currency"`
	AssetID  int64    `json:"asset_id,omitempty"`
	Notes    string   `json:"notes"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags"`
}

type LunchMoneyTag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Archived    bool   `json:"archived"`
}
