# Implementation TODO

## Phase 1: Project Setup
- [x] Initialize Go module
- [ ] Create .gitignore
- [ ] Create .env.example
- [ ] Set up project directory structure

## Phase 2: Configuration
- [ ] Implement config/config.go to load environment variables
  - Splitwise API key
  - Lunch Money access token
  - Lunch Money asset ID (optional)
  - Sync interval
  - Log level

## Phase 3: API Clients
- [ ] Implement splitwise/client.go
  - GetCurrentUser() - fetch authenticated user ID
  - GetExpenses(since time.Time) - fetch expenses updated since timestamp
  - Define Expense struct matching Splitwise API response

- [ ] Implement lunchmoney/client.go
  - InsertTransactions(transactions []Transaction) - create transactions
  - Define Transaction struct matching Lunch Money API request

## Phase 4: Sync Logic
- [ ] Implement sync/state.go
  - Load/Save sync state from/to JSON file
  - Track last sync timestamp
  - Track synced expense IDs
  - Thread-safe operations with mutex

- [ ] Implement sync/engine.go
  - TransformExpense() - convert Splitwise expense to Lunch Money transaction(s)
  - Handle Case A: You paid (create debit + optional credit)
  - Handle Case B: You owe (create debit)
  - Filter deleted expenses
  - Build notes with receipt URLs, categories, etc.

## Phase 5: Main Application
- [ ] Implement main.go
  - Initialize config
  - Initialize API clients
  - Get current user ID from Splitwise
  - Run continuous sync loop with ticker
  - Handle SIGINT/SIGTERM for graceful shutdown
  - Structured logging with slog

## Phase 6: Testing
- [ ] Manual testing with real API credentials
- [ ] Verify duplicate prevention
- [ ] Verify correct amount signs (negative for expenses, positive for income)
- [ ] Verify tags are created correctly
- [ ] Test error handling
