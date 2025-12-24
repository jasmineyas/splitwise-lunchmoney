package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jasmineyas/splitwise-lunchmoney/config"
	"github.com/jasmineyas/splitwise-lunchmoney/lunchmoney"
	"github.com/jasmineyas/splitwise-lunchmoney/models"
	"github.com/jasmineyas/splitwise-lunchmoney/splitwise"
	syncengine "github.com/jasmineyas/splitwise-lunchmoney/syncEngine"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting Splitwise-LunchMoney Sync")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	logger.Info("Config loaded successfully", "config", cfg)

	// initialize clients and sync engine here
	swClient := splitwise.NewClient(cfg.SplitwiseBearerToken)
	lmClient := lunchmoney.NewClient(cfg.UserALunchMoney.BearerToken)
	engine := syncengine.New(swClient, lmClient, cfg)

	// wrap everything in a loop

	// 1. fetch data - all expenses and associated comments with a friend
	expenses, err := swClient.GetAllExpenses(cfg.UserBSplitwiseID)
	if err != nil {
		logger.Error("Error fetching expenses with friend", "error", err)
		return
	}

	logger.Info("Fetched expenses with friend", "count", len(expenses))
	commentsMap := make(map[int64][]models.SplitwiseComment)
	for _, expense := range expenses {
		comments, err := swClient.GetExpenseComments(expense.ID)
		if err != nil {
			logger.Error("Error fetching comments for expense", "expenseID", expense.ID, "error", err)
			return
		}
		commentsMap[expense.ID] = comments
	}

	logger.Info("Fetched comments for expenses", "count", len(commentsMap))

	// 2. detect changes
	// toCreate, toUpdate, toDelete := detector.DetectChanges()

	// 3. execute sync data

}
