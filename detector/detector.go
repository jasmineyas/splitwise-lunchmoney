package detector

import (
	"github.com/jasmineyas/splitwise-lunchmoney/models"
)

// likely return a new type of struct... like deleteStruct, updateStruct, createStruct
func DetectChanges(expenses []models.SplitwiseExpense, commentsMap map[int64][]models.SplitwiseComment) (toCreate, toUpdate, toDelete []models.SplitwiseExpense, err error) {
	// placeholder implementation
	return nil, nil, nil, nil
}
