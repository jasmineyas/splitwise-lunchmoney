package detector

import (
	"testing"

	"github.com/jasmineyas/splitwise-lunchmoney/models"
)

func TestHasLegacyTag(t *testing.T) {
	tests := []struct {
		name     string
		comments []models.SplitwiseComment
		want     bool
	}{
		{
			name: "has exact legacy tag",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "pre-lunchmoney-sync"},
				{ID: 2, Content: "some other comment"},
			},
			want: true,
		},
		{
			name: "no legacy tag",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "synced-to-LM"},
				{ID: 2, Content: "some comment"},
			},
			want: false,
		},
		{
			name: "similar but not exact legacy tag",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "pre-lunchmoney-sync-v2"},
			},
			want: false,
		},
		{
			name:     "empty comments",
			comments: []models.SplitwiseComment{},
			want:     false,
		},
		{
			name:     "nil comments",
			comments: nil,
			want:     false,
		},
		{
			name: "legacy tag among multiple comments",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "first comment"},
				{ID: 2, Content: "second comment"},
				{ID: 3, Content: "pre-lunchmoney-sync"},
				{ID: 4, Content: "fourth comment"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasLegacyTag(tt.comments)
			if got != tt.want {
				t.Errorf("hasLegacyTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindSyncComment(t *testing.T) {
	tests := []struct {
		name     string
		comments []models.SplitwiseComment
		tag      string
		wantID   int64 // ID of the expected comment, 0 if nil expected
		wantNil  bool
	}{
		{
			name: "finds exact tag match",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "test"},
				{ID: 2, Content: "other"},
			},
			tag:     "test",
			wantID:  1,
			wantNil: false,
		},
		{
			name: "finds tag within content",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "This is a test comment"},
				{ID: 2, Content: "other"},
			},
			tag:     "test",
			wantID:  1,
			wantNil: false,
		},
		{
			name: "finds synced-to-LM tag with metadata",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "random comment"},
				{ID: 2, Content: "synced-to-LM: tx_12345, hash: abc123"},
			},
			tag:     "synced-to-LM",
			wantID:  2,
			wantNil: false,
		},
		{
			name: "returns first match when multiple exist",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "test first"},
				{ID: 2, Content: "test second"},
				{ID: 3, Content: "test third"},
			},
			tag:     "test",
			wantID:  1,
			wantNil: false,
		},
		{
			name: "tag not found",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "some comment"},
				{ID: 2, Content: "another comment"},
			},
			tag:     "nonexistent",
			wantNil: true,
		},
		{
			name:     "empty comments list",
			comments: []models.SplitwiseComment{},
			tag:      "test",
			wantNil:  true,
		},
		{
			name:     "nil comments",
			comments: nil,
			tag:      "test",
			wantNil:  true,
		},
		{
			name: "empty tag matches all (contains behavior)",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "any content"},
			},
			tag:     "",
			wantID:  1, // Empty string is contained in any string
			wantNil: false,
		},
		{
			name: "case sensitive search",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "TEST uppercase"},
				{ID: 2, Content: "test lowercase"},
			},
			tag:     "test",
			wantID:  2, // Should find lowercase, not uppercase
			wantNil: false,
		},
		{
			name: "finds tag at beginning of content",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "test at start"},
			},
			tag:     "test",
			wantID:  1,
			wantNil: false,
		},
		{
			name: "finds tag at end of content",
			comments: []models.SplitwiseComment{
				{ID: 1, Content: "content ends with test"},
			},
			tag:     "test",
			wantID:  1,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findSyncComment(tt.comments, tt.tag)

			if tt.wantNil {
				if got != nil {
					t.Errorf("findSyncComment() = %v (ID: %d), want nil", got.Content, got.ID)
				}
			} else {
				if got == nil {
					t.Error("findSyncComment() = nil, want non-nil")
					return
				}
				if got.ID != tt.wantID {
					t.Errorf("findSyncComment() returned comment ID %d with content %q, want ID %d", got.ID, got.Content, tt.wantID)
				}
			}
		})
	}
}

func TestDetectChanges(t *testing.T) {
	tests := []struct {
		name            string
		expenses        []models.SplitwiseExpense
		commentsMap     map[int64][]models.SplitwiseComment
		wantCreateCount int
		wantCreateIDs   []int64 // IDs of expenses that should be created
		wantErr         bool
	}{
		{
			name: "new expense without any comments",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "New expense"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {},
			},
			wantCreateCount: 1,
			wantCreateIDs:   []int64{1},
		},
		{
			name: "new expense with nil comments",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "New expense"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: nil,
			},
			wantCreateCount: 1,
			wantCreateIDs:   []int64{1},
		},
		{
			name: "expense with legacy tag - should skip",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "Old expense"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {
					{ID: 100, Content: "pre-lunchmoney-sync"},
				},
			},
			wantCreateCount: 0,
			wantCreateIDs:   []int64{},
		},
		{
			name: "expense with sync comment - should skip",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "Synced expense"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {
					{ID: 100, Content: "test comment synced"},
				},
			},
			wantCreateCount: 0,
			wantCreateIDs:   []int64{},
		},
		{
			name: "multiple expenses - all new",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "New expense 1"},
				{ID: 2, Description: "New expense 2"},
				{ID: 3, Description: "New expense 3"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {},
				2: {},
				3: {},
			},
			wantCreateCount: 3,
			wantCreateIDs:   []int64{1, 2, 3},
		},
		{
			name: "multiple expenses - mixed states",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "New expense 1"},
				{ID: 2, Description: "Legacy expense"},
				{ID: 3, Description: "New expense 2"},
				{ID: 4, Description: "Synced expense"},
				{ID: 5, Description: "New expense 3"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {},
				2: {
					{ID: 100, Content: "pre-lunchmoney-sync"},
				},
				3: {
					{ID: 101, Content: "random comment"},
				},
				4: {
					{ID: 102, Content: "test sync metadata"},
				},
				5: nil,
			},
			wantCreateCount: 3,
			wantCreateIDs:   []int64{1, 3, 5},
		},
		{
			name:            "empty expenses list",
			expenses:        []models.SplitwiseExpense{},
			commentsMap:     map[int64][]models.SplitwiseComment{},
			wantCreateCount: 0,
			wantCreateIDs:   []int64{},
		},
		{
			name:            "nil expenses",
			expenses:        nil,
			commentsMap:     map[int64][]models.SplitwiseComment{},
			wantCreateCount: 0,
			wantCreateIDs:   []int64{},
		},
		{
			name: "expense with multiple comments but no sync tag",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "Expense with comments"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {
					{ID: 100, Content: "first comment"},
					{ID: 101, Content: "second comment"},
					{ID: 102, Content: "third comment"},
				},
			},
			wantCreateCount: 1,
			wantCreateIDs:   []int64{1},
		},
		{
			name: "expense with both legacy and sync tags - legacy takes precedence",
			expenses: []models.SplitwiseExpense{
				{ID: 1, Description: "Conflicting tags"},
			},
			commentsMap: map[int64][]models.SplitwiseComment{
				1: {
					{ID: 100, Content: "pre-lunchmoney-sync"},
					{ID: 101, Content: "test sync"},
				},
			},
			wantCreateCount: 0, // Should skip due to legacy tag check first
			wantCreateIDs:   []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreate, err := DetectChanges(tt.expenses, tt.commentsMap)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check create count
			if len(gotCreate) != tt.wantCreateCount {
				t.Errorf("DetectChanges() created %d items, want %d", len(gotCreate), tt.wantCreateCount)
			}

			// Verify the correct expenses are marked for creation
			if len(tt.wantCreateIDs) > 0 {
				gotIDs := make(map[int64]bool)
				for _, exp := range gotCreate {
					gotIDs[exp.ID] = true
				}

				for _, wantID := range tt.wantCreateIDs {
					if !gotIDs[wantID] {
						t.Errorf("Expected expense ID %d to be in create list, but it wasn't found", wantID)
					}
				}

				// Check for unexpected IDs
				for gotID := range gotIDs {
					found := false
					for _, wantID := range tt.wantCreateIDs {
						if gotID == wantID {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Unexpected expense ID %d in create list", gotID)
					}
				}
			}
		})
	}
}
