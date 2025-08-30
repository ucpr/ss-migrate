package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/ucpr/ss-migrate/internal/schema"
	"github.com/ucpr/ss-migrate/internal/sheet"
)

// Applier handles applying schema changes to sheets
type Applier struct {
	sheetClient *sheet.Client
	dryRun      bool
}

// NewApplier creates a new applier instance
func NewApplier(sheetClient *sheet.Client, dryRun bool) *Applier {
	return &Applier{
		sheetClient: sheetClient,
		dryRun:      dryRun,
	}
}

// ApplyResult represents the result of applying changes
type ApplyResult struct {
	Success        bool
	Message        string
	ChangesApplied int
	Errors         []error
}

// Apply applies the schema changes to the sheet
func (a *Applier) Apply(ctx context.Context, schemaConfig *schema.Schema, diff *DiffResult) (*ApplyResult, error) {
	if !diff.HasChanges {
		return &ApplyResult{
			Success: true,
			Message: "No changes to apply",
		}, nil
	}

	if a.dryRun {
		return &ApplyResult{
			Success:        true,
			Message:        fmt.Sprintf("DRY RUN: Would apply %d changes", len(diff.Changes)),
			ChangesApplied: len(diff.Changes),
		}, nil
	}

	result := &ApplyResult{
		Success: true,
		Errors:  []error{},
	}

	// Apply changes
	for _, change := range diff.Changes {
		err := a.applyChange(ctx, schemaConfig, change)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to apply %s: %w", change.Path, err))
			result.Success = false
		} else {
			result.ChangesApplied++
		}
	}

	if result.Success {
		result.Message = fmt.Sprintf("Successfully applied %d changes", result.ChangesApplied)
	} else {
		result.Message = fmt.Sprintf("Applied %d changes with %d errors", result.ChangesApplied, len(result.Errors))
	}

	return result, nil
}

// applyChange applies a single change to the sheet
func (a *Applier) applyChange(ctx context.Context, schemaConfig *schema.Schema, change Change) error {
	// Parse the path to get sheet name and field name
	parts := strings.Split(change.Path, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid change path: %s", change.Path)
	}

	sheetName := parts[0]
	
	// Find the resource for this sheet
	var resource *schema.Resource
	for i := range schemaConfig.Resources {
		if schemaConfig.Resources[i].Name == sheetName {
			resource = &schemaConfig.Resources[i]
			break
		}
	}

	if resource == nil {
		return fmt.Errorf("resource not found for sheet: %s", sheetName)
	}

	// Extract spreadsheet ID
	spreadsheetID, err := sheet.ExtractSpreadsheetID(resource.Path)
	if err != nil {
		return fmt.Errorf("failed to extract spreadsheet ID: %w", err)
	}

	switch change.Type {
	case ChangeTypeAdd:
		return a.addField(ctx, spreadsheetID, sheetName, change, resource)
	case ChangeTypeRemove:
		return a.removeField(ctx, spreadsheetID, sheetName, change, resource.HeaderRow)
	case ChangeTypeModify:
		// For now, we'll log modify changes but not actually change data types
		// as this could cause data loss
		fmt.Printf("Warning: Field type modification for %s requires manual intervention\n", change.Path)
		return nil
	default:
		return fmt.Errorf("unsupported change type: %s", change.Type)
	}
}

// addField adds a new field to the sheet in the correct position according to schema order
func (a *Applier) addField(ctx context.Context, spreadsheetID, sheetName string, change Change, resource *schema.Resource) error {
	headerRow := resource.HeaderRow
	if headerRow == 0 {
		headerRow = 1
	}

	// Get current headers
	headers, err := a.sheetClient.GetHeaders(ctx, spreadsheetID, sheetName, headerRow)
	if err != nil {
		return fmt.Errorf("failed to get headers: %w", err)
	}

	// Find the field info from the change
	fieldInfo, ok := change.NewValue.(FieldInfo)
	if !ok {
		return fmt.Errorf("invalid field info in change")
	}

	// Check if field already exists
	for _, header := range headers {
		if header == fieldInfo.Name {
			return fmt.Errorf("field %s already exists", fieldInfo.Name)
		}
	}

	// Find the correct position based on schema order
	schemaFieldIndex := -1
	for i, field := range resource.Fields {
		if field.Name == fieldInfo.Name {
			schemaFieldIndex = i
			break
		}
	}

	if schemaFieldIndex == -1 {
		return fmt.Errorf("field %s not found in schema", fieldInfo.Name)
	}

	// Determine the insert position
	insertColumnIndex := len(headers) // Default to end

	// Find where to insert based on schema order
	// Look for the first field after this one that exists in the current headers
	for i := schemaFieldIndex + 1; i < len(resource.Fields); i++ {
		nextFieldName := resource.Fields[i].Name
		// Find this field in current headers
		for j, header := range headers {
			if header == nextFieldName {
				// Insert before this field
				insertColumnIndex = j
				break
			}
		}
		// If we found a position, stop searching
		if insertColumnIndex < len(headers) {
			break
		}
	}

	// If we need to insert in the middle, we need to shift existing columns
	if insertColumnIndex < len(headers) {
		// For now, we'll insert at the position by using InsertColumn
		err = a.sheetClient.InsertColumn(ctx, spreadsheetID, sheetName, insertColumnIndex)
		if err != nil {
			return fmt.Errorf("failed to insert column: %w", err)
		}
	}

	// Add the header at the correct position
	columnLetter := sheet.ColumnToLetter(insertColumnIndex)
	cellRange := fmt.Sprintf("%s!%s%d", sheetName, columnLetter, headerRow)

	// Update the header cell
	values := [][]interface{}{
		{fieldInfo.Name},
	}

	err = a.sheetClient.UpdateValues(ctx, spreadsheetID, cellRange, values)
	if err != nil {
		return fmt.Errorf("failed to add field header: %w", err)
	}

	fmt.Printf("Added field '%s' to column %s\n", fieldInfo.Name, columnLetter)
	return nil
}

// removeField removes a field from the sheet
func (a *Applier) removeField(ctx context.Context, spreadsheetID, sheetName string, change Change, headerRow int) error {
	if headerRow == 0 {
		headerRow = 1
	}

	// Get current headers
	headers, err := a.sheetClient.GetHeaders(ctx, spreadsheetID, sheetName, headerRow)
	if err != nil {
		return fmt.Errorf("failed to get headers: %w", err)
	}

	// Find the field info from the change
	fieldInfo, ok := change.OldValue.(FieldInfo)
	if !ok {
		return fmt.Errorf("invalid field info in change")
	}

	// Find the column index
	columnIndex := -1
	for i, header := range headers {
		if header == fieldInfo.Name {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return fmt.Errorf("field %s not found", fieldInfo.Name)
	}

	// For safety, we'll just clear the header instead of deleting the entire column
	// This preserves data in case of mistakes
	columnLetter := sheet.ColumnToLetter(columnIndex)
	cellRange := fmt.Sprintf("%s!%s%d", sheetName, columnLetter, headerRow)

	err = a.sheetClient.ClearValues(ctx, spreadsheetID, cellRange)
	if err != nil {
		return fmt.Errorf("failed to clear field header: %w", err)
	}

	fmt.Printf("Cleared header for field '%s' in column %s (data preserved)\n", fieldInfo.Name, columnLetter)
	return nil
}

// ApplyAll applies changes for all resources in the schema
func (a *Applier) ApplyAll(ctx context.Context, schemaConfig *schema.Schema) ([]*ApplyResult, error) {
	// First, create a planner to get the diffs
	planner := NewPlanner(a.sheetClient)
	
	// Get all diffs
	diffs, err := planner.PlanAll(ctx, schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Apply each diff
	results := []*ApplyResult{}
	for i, diff := range diffs {
		if !diff.HasChanges {
			results = append(results, &ApplyResult{
				Success: true,
				Message: fmt.Sprintf("No changes for %s", schemaConfig.Resources[i].Name),
			})
			continue
		}

		result, err := a.Apply(ctx, schemaConfig, diff)
		if err != nil {
			return nil, fmt.Errorf("failed to apply changes for %s: %w", schemaConfig.Resources[i].Name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CreateSheetIfNotExists creates a new sheet if it doesn't exist
func (a *Applier) CreateSheetIfNotExists(ctx context.Context, spreadsheetID, sheetName string) error {
	// Check if sheet already exists
	exists, err := a.sheetClient.CheckSheetExists(ctx, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("failed to check sheet existence: %w", err)
	}

	if exists {
		return nil // Sheet already exists
	}

	if a.dryRun {
		fmt.Printf("DRY RUN: Would create sheet '%s'\n", sheetName)
		return nil
	}

	// Create new sheet using the sheet client
	err = a.sheetClient.CreateSheet(ctx, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	fmt.Printf("Created new sheet '%s'\n", sheetName)
	return nil
}