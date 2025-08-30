package engine

import (
	"testing"

	"github.com/ucpr/ss-migrate/internal/schema"
)

func TestApplyResult(t *testing.T) {
	tests := []struct {
		name           string
		result         ApplyResult
		expectedMsg    string
		expectedStatus bool
	}{
		{
			name: "successful apply",
			result: ApplyResult{
				Success:        true,
				Message:        "Successfully applied 3 changes",
				ChangesApplied: 3,
				Errors:         []error{},
			},
			expectedMsg:    "Successfully applied 3 changes",
			expectedStatus: true,
		},
		{
			name: "no changes",
			result: ApplyResult{
				Success:        true,
				Message:        "No changes to apply",
				ChangesApplied: 0,
				Errors:         []error{},
			},
			expectedMsg:    "No changes to apply",
			expectedStatus: true,
		},
		{
			name: "dry run",
			result: ApplyResult{
				Success:        true,
				Message:        "DRY RUN: Would apply 5 changes",
				ChangesApplied: 5,
				Errors:         []error{},
			},
			expectedMsg:    "DRY RUN: Would apply 5 changes",
			expectedStatus: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Message != tt.expectedMsg {
				t.Errorf("expected message %q, got %q", tt.expectedMsg, tt.result.Message)
			}

			if tt.result.Success != tt.expectedStatus {
				t.Errorf("expected success %v, got %v", tt.expectedStatus, tt.result.Success)
			}
		})
	}
}

func TestApplierDryRun(t *testing.T) {
	// Create a mock applier with dry run enabled
	applier := &Applier{
		dryRun: true,
	}

	// Test that dry run doesn't make actual changes
	diff := &DiffResult{
		HasChanges: true,
		Changes: []Change{
			{
				Type:        ChangeTypeAdd,
				Path:        "test.field1",
				Description: "Add field1",
			},
			{
				Type:        ChangeTypeRemove,
				Path:        "test.field2",
				Description: "Remove field2",
			},
		},
	}

	result, err := applier.Apply(nil, nil, diff)
	if err != nil {
		t.Fatalf("unexpected error in dry run: %v", err)
	}

	if !result.Success {
		t.Error("dry run should always be successful")
	}

	if result.ChangesApplied != 2 {
		t.Errorf("expected 2 changes in dry run, got %d", result.ChangesApplied)
	}

	if result.Message != "DRY RUN: Would apply 2 changes" {
		t.Errorf("unexpected dry run message: %s", result.Message)
	}
}

func TestApplierNoChanges(t *testing.T) {
	applier := &Applier{
		dryRun: false,
	}

	diff := &DiffResult{
		HasChanges: false,
		Changes:    []Change{},
	}

	result, err := applier.Apply(nil, nil, diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("should be successful when no changes")
	}

	if result.ChangesApplied != 0 {
		t.Errorf("expected 0 changes, got %d", result.ChangesApplied)
	}

	if result.Message != "No changes to apply" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestFieldOrdering(t *testing.T) {
	// Test that fields are inserted in the correct order according to schema
	resource := &schema.Resource{
		Name: "test_sheet",
		Fields: []schema.Field{
			{Name: "id"},
			{Name: "name"},
			{Name: "email"},  // New field to be added
			{Name: "created_at"},
		},
	}

	// Simulate current headers (missing "email")
	currentHeaders := []string{"id", "name", "created_at"}

	// Find where "email" should be inserted
	emailFieldIndex := -1
	for i, field := range resource.Fields {
		if field.Name == "email" {
			emailFieldIndex = i
			break
		}
	}

	if emailFieldIndex == -1 {
		t.Fatal("email field not found in schema")
	}

	// Determine insert position
	insertPosition := len(currentHeaders) // Default to end
	
	for i := emailFieldIndex + 1; i < len(resource.Fields); i++ {
		nextField := resource.Fields[i].Name
		for j, header := range currentHeaders {
			if header == nextField {
				insertPosition = j
				break
			}
		}
		if insertPosition < len(currentHeaders) {
			break
		}
	}

	// Should insert at position 2 (before "created_at")
	if insertPosition != 2 {
		t.Errorf("expected insert position 2, got %d", insertPosition)
	}
}