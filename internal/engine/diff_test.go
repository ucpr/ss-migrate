package engine

import (
	"strings"
	"testing"
)

func TestCompareFields(t *testing.T) {
	tests := []struct {
		name           string
		currentFields  []FieldInfo
		schemaFields   []FieldInfo
		expectedAdd    int
		expectedRemove int
		expectedModify int
	}{
		{
			name: "no changes",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
			},
			expectedAdd:    0,
			expectedRemove: 0,
			expectedModify: 0,
		},
		{
			name: "add new fields",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
				{Name: "email", Type: "string"},
			},
			expectedAdd:    2,
			expectedRemove: 0,
			expectedModify: 0,
		},
		{
			name: "remove fields",
			currentFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
				{Name: "old_field", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
			},
			expectedAdd:    0,
			expectedRemove: 1,
			expectedModify: 0,
		},
		{
			name: "modify field types",
			currentFields: []FieldInfo{
				{Name: "id", Type: "string"},
				{Name: "count", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "count", Type: "number"},
			},
			expectedAdd:    0,
			expectedRemove: 0,
			expectedModify: 2,
		},
		{
			name: "mixed changes",
			currentFields: []FieldInfo{
				{Name: "id", Type: "string"},
				{Name: "old_field", Type: "string"},
			},
			schemaFields: []FieldInfo{
				{Name: "id", Type: "integer"},
				{Name: "new_field", Type: "boolean"},
			},
			expectedAdd:    1,
			expectedRemove: 1,
			expectedModify: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := CompareFields(tt.currentFields, tt.schemaFields)

			if len(diff.FieldsToAdd) != tt.expectedAdd {
				t.Errorf("expected %d fields to add, got %d", tt.expectedAdd, len(diff.FieldsToAdd))
			}

			if len(diff.FieldsToRemove) != tt.expectedRemove {
				t.Errorf("expected %d fields to remove, got %d", tt.expectedRemove, len(diff.FieldsToRemove))
			}

			if len(diff.FieldsToModify) != tt.expectedModify {
				t.Errorf("expected %d fields to modify, got %d", tt.expectedModify, len(diff.FieldsToModify))
			}
		})
	}
}

func TestDiffResultFormat(t *testing.T) {
	tests := []struct {
		name     string
		result   DiffResult
		contains []string
	}{
		{
			name: "no changes",
			result: DiffResult{
				HasChanges: false,
			},
			contains: []string{"No changes detected"},
		},
		{
			name: "with changes",
			result: DiffResult{
				HasChanges: true,
				Summary:    "Sheet 'test': 1 field(s) to add",
				Changes: []Change{
					{
						Type:        ChangeTypeAdd,
						Path:        "test.new_field",
						Description: "Add new field 'new_field' of type string",
					},
				},
			},
			contains: []string{
				"Schema Migration Plan",
				"Sheet 'test': 1 field(s) to add",
				"+ test.new_field",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.result.Format()
			for _, expected := range tt.contains {
				if !strings.Contains(formatted, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, formatted)
				}
			}
		})
	}
}

func TestConvertDiffToResult(t *testing.T) {
	diff := &SheetDiff{
		SheetName: "test_sheet",
		FieldsToAdd: []FieldInfo{
			{Name: "new_field", Type: "string"},
		},
		FieldsToRemove: []FieldInfo{
			{Name: "old_field", Type: "integer"},
		},
		FieldsToModify: []FieldDiff{
			{
				Name:        "modified_field",
				Type:        ChangeTypeModify,
				OldType:     "string",
				NewType:     "integer",
				Description: "Type change from string to integer",
			},
		},
	}

	result := ConvertDiffToResult(diff, "test_sheet")

	if !result.HasChanges {
		t.Error("expected HasChanges to be true")
	}

	if len(result.Changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(result.Changes))
	}

	// Check change types
	changeTypes := map[ChangeType]int{
		ChangeTypeAdd:    0,
		ChangeTypeRemove: 0,
		ChangeTypeModify: 0,
	}

	for _, change := range result.Changes {
		changeTypes[change.Type]++
	}

	if changeTypes[ChangeTypeAdd] != 1 {
		t.Errorf("expected 1 ADD change, got %d", changeTypes[ChangeTypeAdd])
	}

	if changeTypes[ChangeTypeRemove] != 1 {
		t.Errorf("expected 1 REMOVE change, got %d", changeTypes[ChangeTypeRemove])
	}

	if changeTypes[ChangeTypeModify] != 1 {
		t.Errorf("expected 1 MODIFY change, got %d", changeTypes[ChangeTypeModify])
	}
}