package engine

import (
	"testing"
)

func TestFieldOrderDetection(t *testing.T) {
	tests := []struct {
		name          string
		currentFields []FieldInfo
		schemaFields  []FieldInfo
		expectReorder bool
		expectedOrder []string
	}{
		{
			name: "no reorder needed - same order",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
				{Name: "email", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
			},
			expectReorder: false,
			expectedOrder: []string{"id", "name", "email"},
		},
		{
			name: "reorder needed - fields swapped",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "email", Type: "string"},
				{Name: "name", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
			},
			expectReorder: true,
			expectedOrder: []string{"id", "name", "email"},
		},
		{
			name: "reorder needed - completely reversed",
			currentFields: []FieldInfo{
				{Name: "email", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "id", Type: "integer"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
			},
			expectReorder: true,
			expectedOrder: []string{"id", "name", "email"},
		},
		{
			name: "no reorder when fields to be removed",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "oldfield", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "email", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
			},
			expectReorder: false,
			expectedOrder: []string{"id", "name", "email"},
		},
		{
			name: "reorder needed with field to be removed",
			currentFields: []FieldInfo{
				{Name: "email", Type: "string"},
				{Name: "oldfield", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "id", Type: "integer"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
			},
			expectReorder: true,
			expectedOrder: []string{"id", "name", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := CompareFields(tt.currentFields, tt.schemaFields)
			
			if diff.FieldsToReorder != tt.expectReorder {
				t.Errorf("expected FieldsToReorder=%v, got %v", tt.expectReorder, diff.FieldsToReorder)
			}

			if tt.expectReorder {
				// Check expected order
				if len(diff.ExpectedOrder) != len(tt.expectedOrder) {
					t.Errorf("expected order length %d, got %d", len(tt.expectedOrder), len(diff.ExpectedOrder))
				}
				for i := range tt.expectedOrder {
					if i < len(diff.ExpectedOrder) && diff.ExpectedOrder[i] != tt.expectedOrder[i] {
						t.Errorf("expected order[%d]=%s, got %s", i, tt.expectedOrder[i], diff.ExpectedOrder[i])
					}
				}
			}
		})
	}
}

func TestReorderChangeInDiffResult(t *testing.T) {
	// Test that reorder changes are properly included in DiffResult
	schemaFields := []FieldInfo{
		{Name: "id", Type: "integer", Position: 0},
		{Name: "name", Type: "string", Position: 1},
		{Name: "email", Type: "string", Position: 2},
	}

	currentFields := []FieldInfo{
		{Name: "email", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
	}

	diff := CompareFields(currentFields, schemaFields)
	result := ConvertDiffToResult(diff, "TestSheet")

	// Should have a reorder change
	if !result.HasChanges {
		t.Error("expected HasChanges=true for reordered fields")
	}

	// Find the reorder change
	hasReorderChange := false
	for _, change := range result.Changes {
		if change.Type == ChangeTypeReorder {
			hasReorderChange = true
			if change.Path != "TestSheet" {
				t.Errorf("expected reorder change path='TestSheet', got %s", change.Path)
			}
			expectedOrder, ok := change.NewValue.([]string)
			if !ok {
				t.Error("reorder change NewValue should be []string")
			}
			if len(expectedOrder) != 3 || expectedOrder[0] != "id" || expectedOrder[1] != "name" || expectedOrder[2] != "email" {
				t.Errorf("unexpected order in reorder change: %v", expectedOrder)
			}
			break
		}
	}

	if !hasReorderChange {
		t.Error("expected a reorder change in result but found none")
	}

	// Check summary mentions reordering
	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
}