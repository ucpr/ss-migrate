package engine

import (
	"strings"
	"testing"
)

func TestDiffOutputInSchemaFieldOrder(t *testing.T) {
	tests := []struct {
		name           string
		currentFields  []FieldInfo
		schemaFields   []FieldInfo
		expectedOrder  []string // Expected order of field names in output
	}{
		{
			name: "mixed changes output in schema order",
			currentFields: []FieldInfo{
				{Name: "email", Type: "string"},
				{Name: "id", Type: "string"}, // Wrong type
				{Name: "oldfield", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1},
				{Name: "email", Type: "string", Hidden: true, Position: 2},
				{Name: "created_at", Type: "datetime", Position: 3},
			},
			expectedOrder: []string{"id", "name", "email", "created_at", "oldfield"},
		},
		{
			name: "all changes types in schema order",
			currentFields: []FieldInfo{
				{Name: "created_at", Type: "string"}, // Wrong type
				{Name: "email", Type: "string"},
				{Name: "deprecated", Type: "string"}, // To be removed
				{Name: "id", Type: "integer"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer", Position: 0},
				{Name: "name", Type: "string", Position: 1}, // To be added
				{Name: "email", Type: "string", Position: 2},
				{Name: "created_at", Type: "datetime", Position: 3}, // Type change
				{Name: "updated_at", Type: "datetime", Position: 4}, // To be added
			},
			// Only fields with changes appear in output, in schema order
			// id: no change, email: no change, so they don't appear
			expectedOrder: []string{"name", "created_at", "updated_at", "deprecated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare fields
			diff := CompareFields(tt.currentFields, tt.schemaFields)
			
			// Convert to result with schema field order
			result := ConvertDiffToResultWithOrder(diff, "TestSheet", tt.schemaFields)
			
			// Extract field names from changes (excluding reorder changes)
			var outputOrder []string
			for _, change := range result.Changes {
				if change.Type != ChangeTypeReorder {
					// Extract field name from path (format: "TestSheet.fieldname")
					parts := strings.Split(change.Path, ".")
					if len(parts) == 2 {
						fieldName := parts[1]
						// Only add if not already in list (to handle multiple changes for same field)
						found := false
						for _, name := range outputOrder {
							if name == fieldName {
								found = true
								break
							}
						}
						if !found {
							outputOrder = append(outputOrder, fieldName)
						}
					}
				}
			}
			
			// Verify the order
			if len(outputOrder) != len(tt.expectedOrder) {
				t.Errorf("expected %d fields in output, got %d", len(tt.expectedOrder), len(outputOrder))
				t.Errorf("expected: %v", tt.expectedOrder)
				t.Errorf("got: %v", outputOrder)
				return
			}
			
			for i, expectedField := range tt.expectedOrder {
				if outputOrder[i] != expectedField {
					t.Errorf("position %d: expected field '%s', got '%s'", i, expectedField, outputOrder[i])
					t.Errorf("full expected order: %v", tt.expectedOrder)
					t.Errorf("full actual order: %v", outputOrder)
					break
				}
			}
		})
	}
}

func TestDiffOutputWithoutSchemaOrder(t *testing.T) {
	// Test that the backward-compatible ConvertDiffToResult still works
	currentFields := []FieldInfo{
		{Name: "email", Type: "string"},
		{Name: "id", Type: "integer"},
	}
	
	schemaFields := []FieldInfo{
		{Name: "id", Type: "integer", Position: 0},
		{Name: "name", Type: "string", Position: 1},
		{Name: "email", Type: "string", Position: 2},
	}
	
	diff := CompareFields(currentFields, schemaFields)
	
	// Use the original function (without schema order)
	result := ConvertDiffToResult(diff, "TestSheet")
	
	// Should still have changes
	if !result.HasChanges {
		t.Error("expected changes in result")
	}
	
	// Check that we have an add change for "name"
	hasNameAdd := false
	for _, change := range result.Changes {
		if change.Type == ChangeTypeAdd && strings.Contains(change.Path, "name") {
			hasNameAdd = true
			break
		}
	}
	
	if !hasNameAdd {
		t.Error("expected add change for 'name' field")
	}
}