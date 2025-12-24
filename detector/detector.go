package detector

import (
	"log"
	"strings"

	"github.com/jasmineyas/splitwise-lunchmoney/models"
)

// func DetectChanges(expenses []models.SplitwiseExpense, commentsMap map[int64][]models.SplitwiseComment) (toCreate []models.SplitwiseExpense, toUpdate []models.UpdateAction, toDelete []models.DeleteAction, err error) {

func DetectChanges(expenses []models.SplitwiseExpense, commentsMap map[int64][]models.SplitwiseComment) (toCreate []models.SplitwiseExpense, err error) {
	for _, expense := range expenses {
		comments := commentsMap[expense.ID]

		// Check for legacy tag
		if hasLegacyTag(comments) {
			continue
		}

		// Check for sync comment
		// TODO - update to "Synced-to-LM" when in production
		// okay this is a really bad practice in managing env and variables lol
		syncComment := findSyncComment(comments, "test")

		if syncComment == nil {
			// No sync comment found, mark for creation
			toCreate = append(toCreate, expense)
		} else {
			// skip these for now
			log.Println("Found existing sync comment, skipping update/delete detection for now")
			continue
			// // Sync comment found, check for updates or deletions
			// syncedData, parseErr := parseSyncComment(syncComment.Content)
			// if parseErr != nil {
			// 	return nil, nil, nil, parseErr
			// }

			// // Compare current expense with synced data to determine if update is needed
			// if isExpenseUpdated(expense, syncedData) {
			// 	changedFields := getChangedFields(expense, syncedData)
			// 	updateAction := models.UpdateAction{
			// 		ExpenseID:       expense.ID,
			// 		Expense:         expense,
			// 		ChangedFields:   changedFields,
			// 		OldHash:         syncedData.SnapshotHash,
			// 		NewHash:         computeExpenseHash(expense, comments),
			// 		LunchMoneyTxnID: getLMTxnIDFromSW(comments) // Assuming UserA is the one who synced
			// 	}
			// 	toUpdate = append(toUpdate, updateAction)
			// }

			// // Additional logic can be added here to determine if deletion is needed
		}
	}

	return toCreate, nil
}

func hasLegacyTag(comments []models.SplitwiseComment) bool {
	for _, comment := range comments {
		if comment.Content == "pre-lunchmoney-sync" {
			return true
		}
	}
	return false
}

func findSyncComment(comments []models.SplitwiseComment, tag string) *models.SplitwiseComment {
	for _, comment := range comments {
		if strings.Contains(comment.Content, tag) {
			return &comment
		}
	}
	return nil
}
