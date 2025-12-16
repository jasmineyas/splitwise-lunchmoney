# Implementation TODO

- [x] finish this implementation blueprint
  - clearly state responsibility of each function of each file.

## Phase 0: Script to update all existing splitwise transaction

- [ ] add a comment to all existing transaction with "pre LM <> SW handshake"

## Phase 1: Project Setup

- [x] Initialize Go module
  - Verify Splitwise asset exists in LM before sync
  - Validate current user has access to all expenses
- [x] Create .gitignore
- [x] Create .env.example
- [ ] Set up project directory structure
- [x] define model.go
- [x] create test friend account and test lunch money friend account for testing

## Phase 2: Configuration

- [ ] Implement config/config.go to load environment variables
  - Splitwise API key
  - Lunch Money access token
  - Lunch Money asset ID

## Phase 3: API Clients

NTS: Need to think about rate limits.

- [ ] Implement splitwise/client.go
      Responsibility: this file contains all the things that splitwise does... add a comment to a transation, get current user, get all expense,

  - GetCurrentUser() - fetch authenticated user ID
  - GetComments(expenseID) - fetch comments to check sync status using Get Comment end point
  - CreateComment(expenseID, content) - post sync metadata using Create Comment end point
  - GetExpense(expenseID) - fetch single expense for verification
  - GetExpensesForInsert - how to quickly parse through all transactions (it's quite intensive)
  - GetExpensesForUpdate - compare hashes (purpose: if there's a typo)
  - GetExpensesForDelete -
  - NOTE: Compare the current Splitwise transaction with the metadata in comments to detect changes (this comparison doesn't need to consider which user posted it). However, to keep track of which Lunch Money transaction matches a given Splitwise transaction, we need to check which user posted the "synced-to-LM" tag, as the Lunch Money transaction ID is user-specific. Multiple users could be using this sync service.
  - Define Expense struct matching Splitwise API response

- [ ] Implement lunchmoney/client.go
  - InsertTransactions(transactions []Transaction) - create transactions
  - Define Transaction struct matching Lunch Money API request
  - UpdateTransactions - if the splitwise transaction is modified
  - DeleteTransactions - if the splitwise transaction is deleted
  - NOTE: LM transaction notes - max 350 characters

## Phase 4: Sync Logic

Add log/slog for each session.

- [ ] Implement sync-engine.go
      Responsibility: this file contains all the syncing functions - transform splitwise transaction to lunch money transaction;
  - TransformSWToLMTransaction() - convert Splitwise expense to Lunch Money transaction(s)
    - "creation_method": "payment" - payment type
    - vs regular expenses
  - Post a new comment to Splitwise with the current user:
    - snapshot_hash (to enable quick change detection)
    - `"synced-to-LM"` tag.
    - "user-name" - I guess
    - Lunch Money transaction ID.
    - Snapshot of the Lunch Money API request body and LM response when the LM transaction was created for debugging.
  - Handle update: Create an update to LM.
  - Handle deletions: If a Splitwise transaction is deleted, remove the corresponding Lunch Money transaction.
    TransformSWToLMTransaction(expense Expense, currentUserID int) []Transaction
  - GenerateSnapshotHash(expense Expense) string
  - ParseSyncComment(comment Comment) (lmTxnID, hash, userName string)
  - ShouldSync(expense Expense, comments- []Comment) (action string, reason string)
  - main 4 sync logiq:
    1. User A owes money (from perspective, negative for user A)
    2. User B owes money (to perspective, positive for user A)
    3. User B pays back User A (settlement, negative amount)
    4. User A pays back User B (settlement, positive amount)

## Phase 5: Main Application

- [ ] Implement main.go
  - Initialize config
  - Initialize API clients
  - Get current user ID from Splitwise
  - Run continuous sync loop with ticker
    - Polling interval (every 5 min? 1 hour?)
    - Rate limiting strategy (SW has limits)
    - Retry logic with exponential backoff
    - Graceful shutdown handling
  - check splitwise transactions, analyze for transactions for udpates,

## Phase 6: Testing

- [ ] run the script first (only need to run)
- [ ] Manual testing with real API credentials
- [ ] Verify correct amount signs
- [ ] verify comments are posted correctly
- [ ] Verify tags are created correctly
- [ ] Test error handling
- Tests:
  - Transform unit tests: the 4 cases above, with 1 and N counterparties, plus currency variants and with/without receipt.original.
  - Idempotency: post same expense twice → no second LM write.
  - Update detection: change an amount/participant/date → update LM and overwrite comment (new hash).
  - Deletion: expense & payment deletions remove LM txn(s).
  - Paging & backoff: simulate >1 page; simulate 429/5xx.
  - Legacy tagger (Phase 0): posts the handshake comment only once.
  - Verify that a transaction added for a past date get captured correctly
