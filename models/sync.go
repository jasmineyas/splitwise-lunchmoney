package models

import "time"

type SyncMetadata struct {
	SplitwiseExpenseID int64     `json:"splitwise_expense_id"`
	SnapshotHash       string    `json:"snapshot_hash"` // For change detection
	SyncedAt           time.Time `json:"synced_at"`
	SyncedBy           int64     `json:"synced_by_user_id"` // Who posted this comment

	UserA UserSyncData `json:"user_a"`
	UserB UserSyncData `json:"user_b"`
}

type UserSyncData struct {
	SplitwiseUserID int64  `json:"splitwise_user_id"`
	LMTransactionID int64  `json:"lm_transaction_id,omitempty"` // 0 if not synced
	LMAssetID       int64  `json:"lm_asset_id"`
	LMRequestBody   string `json:"lm_request_body"`
	LMResponseBody  string `json:"lm_response_body"`
	LastSyncedAt    int64  `json:"last_synced_at"`
	Error           string `json:"error,omitempty"`
}

type DeletionMetadata struct {
	DeletedFromLMAt time.Time `json:"deleted_from_lm_at"`
	DeletedBy       int64     `json:"deleted_by_user_id"`
}
