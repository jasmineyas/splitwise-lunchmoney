package models

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

type Asset struct {
    ID                  int64  `json:"id"`
    TypeName            string `json:"type_name"`
    SubtypeName         string `json:"subtype_name"`
    Name                string `json:"name"`
    DisplayName         string `json:"display_name"`
    Balance             string `json:"balance"`
    ToBase              float64 `json:"to_base"`
    BalanceAsOf         string `json:"balance_as_of"`
    ClosedOn            *string `json:"closed_on"`
    Currency            string `json:"currency"`
    InstitutionName     string `json:"institution_name"`
    ExcludeTransactions bool   `json:"exclude_transactions"`
    CreatedAt           string `json:"created_at"`
}

type AssetsResponse struct {
    Assets []Asset `json:"assets"`
}