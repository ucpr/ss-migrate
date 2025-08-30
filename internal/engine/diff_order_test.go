package engine

import (
	"strings"
	"testing"
)

func TestCompareFieldsWithOrder(t *testing.T) {
	tests := []struct {
		name          string
		currentFields []FieldInfo
		schemaFields  []FieldInfo
		expectedOrder []string // Expected order of field names in FieldsToAdd
	}{
		{
			name: "fields added in schema order",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "created_at", Type: "datetime"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
				{Name: "created_at", Type: "datetime", Position: 3},
			},
			expectedOrder: []string{"name", "email"},
		},
		{
			name:          "all fields new - maintain schema order",
			currentFields: []FieldInfo{},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
			},
			expectedOrder: []string{"id", "name", "email"},
		},
		{
			name: "fields added with gaps",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "email", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Position: 2},
				{Name: "phone", Type: "string", Position: 3},
			},
			expectedOrder: []string{"name", "phone"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := CompareFields(tt.currentFields, tt.schemaFields)

			// Check that fields to add are in the expected order
			if len(diff.FieldsToAdd) != len(tt.expectedOrder) {
				t.Errorf("expected %d fields to add, got %d", len(tt.expectedOrder), len(diff.FieldsToAdd))
				return
			}

			for i, expectedName := range tt.expectedOrder {
				if diff.FieldsToAdd[i].Name != expectedName {
					t.Errorf("at position %d: expected field '%s', got '%s'", i, expectedName, diff.FieldsToAdd[i].Name)
				}
			}
		})
	}
}

func TestConvertDiffToResultWithOrder(t *testing.T) {
	diff := &SheetDiff{
		SheetName: "test_sheet",
		FieldsToAdd: []FieldInfo{
			{Name: "email", Type: "string", Position: 2},
			{Name: "name", Type: "string", Position: 1},
			{Name: "phone", Type: "string", Position: 3},
		},
	}

	result := ConvertDiffToResult(diff, "test_sheet")

	// Check that changes are sorted by position
	if len(result.Changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(result.Changes))
	}

	// Expected order after sorting: name (pos 1), email (pos 2), phone (pos 3)
	expectedOrder := []string{"name", "email", "phone"}
	for i, change := range result.Changes {
		// Extract field name from path (format: "sheet.fieldname")
		parts := strings.Split(change.Path, ".")
		if len(parts) != 2 {
			t.Errorf("unexpected path format: %s", change.Path)
			continue
		}
		fieldName := parts[1]
		
		if fieldName != expectedOrder[i] {
			t.Errorf("at position %d: expected field '%s', got '%s'", i, expectedOrder[i], fieldName)
		}

		// Check that position is mentioned in description
		if !strings.Contains(change.Description, "at position") {
			t.Errorf("expected position info in description, got: %s", change.Description)
		}
	}
}

func TestFieldPositionInPlanOutput(t *testing.T) {
	diff := &SheetDiff{
		SheetName: "test_sheet",
		FieldsToAdd: []FieldInfo{
			{Name: "middle_field", Type: "string", Position: 1},
		},
	}

	result := ConvertDiffToResult(diff, "test_sheet")
	
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}

	change := result.Changes[0]
	// Position should be shown as 1-based index (position 2 for index 1)
	if !strings.Contains(change.Description, "at position 2") {
		t.Errorf("expected 'at position 2' in description, got: %s", change.Description)
	}
}