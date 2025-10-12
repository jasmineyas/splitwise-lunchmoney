# Splitwise to Lunch Money Handshake

A Go application that automatically syncs expenses from Splitwise to Lunch Money. When you create a new expense in Splitwise, this service will automatically create a corresponding transaction in Lunch Money.

### Future Vision

Long-term, this could be expanded into a multi-user service where users can securely provide their own credentials (TODO: how...) and have their Splitwise and Lunch Money accounts sync automatically. For now, this is a single-user application for personal use.

## Design Architecture

### Core Components

1. **Splitwise API Client** (`splitwise/client.go`)

   - Handles authentication with Splitwise API
   - Fetches expenses from Splitwise
   - Uses the `/api/v3.0/get_expenses`, `/api/v3.0/get_current_user` endpoints
   - Filters for expenses updated since last sync

2. **Lunch Money API Client** (`lunchmoney/client.go`)

   - Handles authentication with Lunch Money API
   - Creates new transactions in Lunch Money
   - Uses the `/v1/transactions` endpoint
   - Maps Splitwise expense data to Lunch Money transaction format

3. **Sync Engine** (`sync/engine.go`, `sync/state.go`)

   - Main orchestration logic
   - Maintains state of last synced expense (stored in `sync_state.json`)
   - Polls Splitwise at regular intervals using Go's ticker
   - Processes new expenses and creates corresponding Lunch Money transactions
   - Handles duplicate detection (won't sync the same expense twice)

4. **Main Application** (`main.go`)
   - Entry point for the application
   - Loads configuration from environment variables
   - Runs continuous sync loop with configurable interval
   - Includes error handling and structured logging
   - Graceful shutdown on interrupt signals

### Data Flow

```
┌─────────────┐
│  Splitwise  │
│   New       │
│  Expense    │
└──────┬──────┘
       │
       ▼
┌──────────────────────────────┐
│  Sync Service (main.go)      │
│                              │
│  ┌────────────────────────┐  │
│  │ 1. Poll Splitwise API  │  │
│  │    for new expenses    │  │
│  └──────────┬─────────────┘  │
│             │                │
│             ▼                │
│  ┌────────────────────────┐  │
│  │ 2. Check sync_state    │  │
│  │    to avoid dupes      │  │
│  └──────────┬─────────────┘  │
│             │                │
│             ▼                │
│  ┌────────────────────────┐  │
│  │ 3. Transform data      │  │
│  │    Splitwise format    │  │
│  │    → Lunch Money       │  │
│  └──────────┬─────────────┘  │
│             │                │
│             ▼                │
│  ┌────────────────────────┐  │
│  │ 4. Create transaction  │  │
│  │    via LM API          │  │
│  └──────────┬─────────────┘  │
│             │                │
│             ▼                │
│  ┌────────────────────────┐  │
│  │ 5. Update sync_state   │  │
│  └────────────────────────┘  │
└──────────────────────────────┘
       │
       ▼
┌─────────────┐
│ Lunch Money │
│   New       │
│ Transaction │
└─────────────┘
```

### Key Design Decisions

**Binary vs Web App**: This compiles to a standalone Go binary, not a web application. It runs continuously with a configurable polling interval (default: 5 minutes) using Go's `time.Ticker`.

**State Management**: Uses a simple JSON file (`sync_state.json`) to track:

- Timestamp of last successful sync
- IDs of expenses already synced (to prevent duplicates)
- Thread-safe access using Go's sync.Mutex

**Configuration**: Uses environment variables (loaded via `godotenv`) for:

- API credentials (Splitwise API key, Lunch Money access token)
- Lunch Money asset ID (optional - for dedicated Splitwise account)
- Transaction status (cleared vs uncleared - configurable)
- Sync interval
- Logging preferences

**Error Handling**:

- Graceful handling of API failures (network issues, rate limits)
- Continues running even if individual sync attempts fail
- Structured logging using Go's `log/slog` package
- Graceful shutdown on SIGINT/SIGTERM signals

**Concurrency**:

- Single-threaded sync loop for simplicity
- Future-proof for multi-user support using goroutines

**Data Transformation Logic**: This is the core business logic for converting Splitwise expenses to Lunch Money transactions.

### Business Logic: Splitwise → Lunch Money Transformation

#### Step 1: Filter Expenses from Splitwise

When fetching expenses from Splitwise API (`GET /api/v3.0/get_expenses`):

**Skip expense if:**

- `deleted_at` is NOT null (expense was deleted)
- `deleted_by` is NOT null (expense was deleted)

**Process expense if:**

- `deleted_at` is null AND `deleted_by` is null

#### Step 2: Identify Current User

Make a one-time API call to `GET /api/v3.0/get_current_user` to get the authenticated user's ID. This tells us whose perspective we're operating from (your user ID).

Store this as `currentUserID` for the session.

#### Step 3: Analyze User Shares

For each expense, examine the `users` array to find YOUR involvement:

```json
"users": [
  {
    "user_id": 123456,        // Your ID
    "paid_share": "25.46",     // Amount you paid
    "owed_share": "0.0",       // Amount you owe
    "net_balance": "25.46"     // Net: you're owed 25.46
  },
  {
    "user_id": 23456,       // Wesley's ID
    "paid_share": "0.0",       // Amount Wesley paid
    "owed_share": "25.46",     // Amount Wesley owes
    "net_balance": "-25.46"    // Net: Wesley owes 25.46
  }
]
```

#### Step 4: Create Lunch Money Transaction(s)

**Case A: You paid for something (paid_share > 0)**

Create 1 or 2 transactions depending on whether you're owed money:

1. **Debit Transaction** (your expense):

   - `date`: Splitwise `date` (convert from `2025-10-11T06:58:32Z` to `2025-10-11`)
   - `amount`: Your `paid_share` (e.g., `25.46`)
   - `payee`: Splitwise `description` (e.g., "save on foods")
   - `currency`: Splitwise `currency_code` lowercased (e.g., `cad`)
   - `asset_id`: From env var (optional - assigns to dedicated Splitwise account in Lunch Money)
   - `notes`: Include details:
     ```
     Splitwise expense (ID: 4096668238)
     Split with: Wesley Finck
     You paid: $25.46, They owe: $25.46
     [Receipt: <receipt_url>] (if available)
     Category: Groceries (if available)
     ```
   - `status`: `"uncleared"` (uncleared means unreviewed, users typically review transactions in lunchmoney and mark them reviewed "cleared")
   - `external_id`: `"splitwise-{expense_id}"` (e.g., `"splitwise-4096668238"`)
   - `tags`: `["Splitwise-lunchmoney-sync"]` (auto-creates tag if it doesn't exist)

2. **Credit Transaction** (if others owe you):
   - Only create if total `owed_share` from others > 0
   - `date`: Same as debit
   - `amount`: Sum of other users' `owed_share` (e.g., `25.46`) - **negative value** to indicate income
   - `payee`: `"Wesley Finck"` (or multiple names if split)
   - `currency`: Same as debit
   - `asset_id`: Same as debit (from env var)
   - `notes`:
     ```
     Splitwise reimbursement (Expense ID: 4096668238)
     Original expense: save on foods
     Amount owed to you: $25.46
     ```
   - `status`: `"uncleared"`
   - `external_id`: `"splitwise-{expense_id}-reimbursement"`
   - `tags`: `["Splitwise-lunchmoney-sync", "Splitwise-reimbursement"]`(auto-creates tag if it doesn't exist)

**Case B: You owe someone (owed_share > 0)**

Create 1 transaction:

1. **Debit Transaction** (your debt):
   - `date`: Splitwise `date`
   - `amount`: Your `owed_share` (e.g., `17.86`)
   - `payee`: Splitwise `description` + " (via {payer_name})"
   - `currency`: Splitwise `currency_code` lowercased
   - `asset_id`: From env var (optional)
   - `notes`:
     ```
     Splitwise expense (ID: 4096669090)
     Paid by: Wesley Finck
     Your share: $17.86
     [Receipt: <receipt_url>] (if available)
     Category: Groceries (if available)
     ```
   - `status`: `"uncleared"`
   - `external_id`: `"splitwise-{expense_id}"`
   - `tags`: `["Splitwise-lunchmoney-sync"]`

#### Step 5: Handle Special Fields

**Dedicated Account (asset_id):**

- **Setup**: Create a manual account in Lunch Money (e.g., "Splitwise Expenses")
- **Get asset_id**:
  - Option 1: Provide asset_id in `.env` file (you find this manually in Lunch Money)
  - Option 2: Future enhancement - fetch via `GET /v1/assets` API
- **Use**: Include `asset_id` in every transaction to group all Splitwise expenses in one account
- **Optional**: If not provided, transactions go to default account

**Category Mapping:**

- If Splitwise `category.name` exists (e.g., "Groceries"), include it in `notes`
- Future enhancement: Map Splitwise categories to Lunch Money category IDs

**Receipt Image:**

- If `receipt.original` is NOT null, include the URL in `notes` as `[Receipt: {url}]`

**Currency:**

- Splitwise provides uppercase (e.g., "CAD")
- Lunch Money requires lowercase (e.g., "cad")
- Convert: `strings.ToLower(splitwise.CurrencyCode)`

**Date Conversion:**

- Splitwise: `"2025-10-11T07:00:04Z"` (ISO 8601 with time)
- Lunch Money: `"2025-10-11"` (ISO 8601 date only)
- Parse and format: `time.Parse()` then `Format("2006-01-02")`

#### Step 6: API Request to Lunch Money

Send to `POST https://dev.lunchmoney.app/v1/transactions`:

```json
{
  "transactions": [
    {
      "date": "2025-10-11",
      "amount": "-25.46",
      "payee": "save on foods",
      "currency": "cad",
      "asset_id": 12345,
      "notes": "Splitwise expense (ID: 4096668238)...",
      "status": "uncleared",
      "external_id": "splitwise-4096668238",
      "tags": ["Splitwise-lunchmoney-sync"]
    }
  ],
  "apply_rules": false,
  "skip_duplicates": true,
  "debit_as_negative": true
}
```

**Important flags:**

- `skip_duplicates: true` - Lunch Money will auto-dedupe by `external_id`, date, payee, and amount
- `apply_rules: false` - Don't apply Lunch Money's rules (keep raw data)
- `debit_as_negative: true` - debt is negative and credit is positive.

**Key fields:**

- `asset_id`: Optional, assigns to dedicated Splitwise account in Lunch Money
- `status`: "uncleared" means unreviewed, "cleared" means reviewed, users typically review transactions in lunchmoney and mark them reviewed "cleared"

### Future Enhancements (Out of Scope for MVP)

1. **Handle Updated/Deleted Expenses:**

   - Track when Splitwise expenses are deleted (`deleted_at` changes from null)
   - Delete corresponding Lunch Money transactions
   - Requires storing mapping of Splitwise ID → Lunch Money transaction ID(s)

2. **Payment Tracking:**
   - When Splitwise marks a debt as "paid" (via payments), track the same e transfer transaction in lunch money and merge the duplicate? can use "Splitwise-reimbursement" tag to look up the transactions matching? that means we will need to check what the state was before and what's the state now..... that sounds hard.

## Project Structure

```
splitwise-lunchmoney/
├── README.md                 # This file
├── go.mod                   # Go module definition
├── go.sum                   # Go dependency checksums
├── .env.example             # Example environment variables
├── .env                     # Your actual credentials (gitignored)
├── .gitignore               # Git ignore file
├── main.go                  # Entry point - runs the sync loop
├── config/
│   └── config.go            # Configuration loader
├── splitwise/
│   └── client.go            # Splitwise API client
├── lunchmoney/
│   └── client.go            # Lunch Money API client
├── sync/
│   ├── engine.go            # Sync orchestration logic
│   └── state.go             # State management with persistence
└── sync_state.json          # Sync state (created automatically)
```

## Building and Running

### Build

```bash
# Build the binary
go build -o splitwise-lunchmoney

# Or build for specific platforms
GOOS=linux GOARCH=amd64 go build -o splitwise-lunchmoney-linux
```

### Run Locally

```bash
# Set up your .env file first (copy from .env.example)
cp .env.example .env
# Edit .env with your API credentials

# Run the binary
./splitwise-lunchmoney
```

## Running Options

### Option 1: Simple Terminal (Testing)

The easiest way to test - just run it in a terminal:

```bash
./splitwise-lunchmoney
```

**Pros**: Simple, immediate feedback
**Cons**: Stops when you close the terminal or put your Mac to sleep

### Option 2: Keep Running in Background on Mac (Recommended for Personal Use)

#### Using launchd (Mac's built-in service manager)

Create a launch agent to keep it running even after restarts:

1. Create a plist file at `~/Library/LaunchAgents/com.yourusername.splitwise-sync.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yourusername.splitwise-sync</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/yourusername/Documents/Projects/splitwise-lunchmoney/splitwise-lunchmoney</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/Users/yourusername/Documents/Projects/splitwise-lunchmoney</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/Users/yourusername/Documents/Projects/splitwise-lunchmoney/logs/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/yourusername/Documents/Projects/splitwise-lunchmoney/logs/stderr.log</string>
</dict>
</plist>
```

2. Create logs directory and load the service:

```bash
mkdir -p logs
launchctl load ~/Library/LaunchAgents/com.yourusername.splitwise-sync.plist
```

3. Manage the service:

```bash
# Check if it's running
launchctl list | grep splitwise-sync

# Stop it
launchctl unload ~/Library/LaunchAgents/com.yourusername.splitwise-sync.plist

# Start it again
launchctl load ~/Library/LaunchAgents/com.yourusername.splitwise-sync.plist

# View logs
tail -f logs/stdout.log
```

**Pros**: Runs in background, auto-starts on login, restarts on crash
**Cons**: Only runs when your Mac is on (stops during sleep unless you configure Mac to prevent sleep)

#### Using screen or tmux (Simpler alternative)

```bash
# Install tmux if you don't have it
brew install tmux

# Start a tmux session
tmux new -s splitwise-sync

# Run the binary
./splitwise-lunchmoney

# Detach from session: Press Ctrl+B, then D
# Reattach later: tmux attach -t splitwise-sync
```

**Pros**: Very simple, easy to check logs
**Cons**: Stops when Mac sleeps or restarts

### Option 3: Remote Server (For True 24/7 Operation)

When you're ready to move to a server, here are your options:

#### VPS/Cloud Server (Recommended)

Services like DigitalOcean ($6/month), AWS EC2, Google Cloud, or Linode:

1. Build for Linux:

   ```bash
   GOOS=linux GOARCH=amd64 go build -o splitwise-lunchmoney-linux
   ```

2. Copy to server:

   ```bash
   scp splitwise-lunchmoney-linux your-server:/home/youruser/
   scp .env your-server:/home/youruser/
   ```

3. Set up systemd service:

   ```bash
   sudo nano /etc/systemd/system/splitwise-sync.service
   ```

   Example service file:

   ```ini
   [Unit]
   Description=Splitwise to Lunch Money Sync
   After=network.target

   [Service]
   Type=simple
   User=youruser
   WorkingDirectory=/home/youruser
   ExecStart=/home/youruser/splitwise-lunchmoney-linux
   Restart=always
   RestartSec=10

   [Install]
   WantedBy=multi-user.target
   ```

   Enable and start:

   ```bash
   sudo systemctl enable splitwise-sync
   sudo systemctl start splitwise-sync
   sudo systemctl status splitwise-sync
   ```

**Pros**: True 24/7 operation, very reliable, low cost
**Cons**: Requires basic server management knowledge, ongoing cost

#### Raspberry Pi

If you have a Raspberry Pi at home:

```bash
# Cross-compile for ARM
GOOS=linux GOARCH=arm64 go build -o splitwise-lunchmoney-pi

# Copy and run similar to VPS setup above
```

**Pros**: One-time hardware cost, runs at home
**Cons**: Depends on your home internet, power outages affect it

## Future Enhancements

- Web UI for configuration and monitoring
- Multi-user support with secure credential storage (using goroutines for concurrent syncing)
- Webhook-based sync (instead of polling) if Splitwise supports it
- More sophisticated expense categorization
- Support for splitting rules and custom mappings
- Docker containerization for easier deployment
- Prometheus metrics for monitoring sync health
- GraphQL API for programmatic access
