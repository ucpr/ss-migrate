package engine

import (
	"fmt"
	"strings"
)

// ChangeType represents the type of change detected
type ChangeType string

const (
	ChangeTypeAdd     ChangeType = "ADD"
	ChangeTypeRemove  ChangeType = "REMOVE"
	ChangeTypeModify  ChangeType = "MODIFY"
	ChangeTypeReorder ChangeType = "REORDER"
	ChangeTypeNone    ChangeType = "NONE"
)

// Change represents a single change detected between schemas
type Change struct {
	Type        ChangeType
	Path        string
	Description string
	OldValue    interface{}
	NewValue    interface{}
}

// DiffResult represents the complete diff between sheet and schema
type DiffResult struct {
	Changes     []Change
	HasChanges  bool
	Summary     string
}

// FieldDiff represents differences in a field
type FieldDiff struct {
	Name        string
	Type        ChangeType
	OldType     string
	NewType     string
	OldFormat   string
	NewFormat   string
	OldHidden   bool
	NewHidden   bool
	Description string
}

// SheetDiff represents differences in sheet structure
type SheetDiff struct {
	SheetName       string
	FieldsToAdd     []FieldInfo
	FieldsToRemove  []FieldInfo
	FieldsToModify  []FieldDiff
	FieldsToReorder bool     // Indicates if fields need reordering
	ExpectedOrder   []string // Expected field order from schema
}

// FieldInfo represents basic field information
type FieldInfo struct {
	Name     string
	Type     string
	Format   string
	Hidden   bool
	Position int // Position in schema for ordering
}

// FormatDiff formats the diff result for display
func (d *DiffResult) Format() string {
	if !d.HasChanges {
		return "No changes detected. Sheet matches the schema."
	}

	var sb strings.Builder
	sb.WriteString("=== Schema Migration Plan ===\n\n")
	sb.WriteString(d.Summary)
	sb.WriteString("\n\nChanges to be applied:\n")

	for _, change := range d.Changes {
		sb.WriteString(formatChange(change))
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatChange(c Change) string {
	switch c.Type {
	case ChangeTypeAdd:
		return fmt.Sprintf("  + %s: %s", c.Path, c.Description)
	case ChangeTypeRemove:
		return fmt.Sprintf("  - %s: %s", c.Path, c.Description)
	case ChangeTypeModify:
		return fmt.Sprintf("  ~ %s: %s", c.Path, c.Description)
	case ChangeTypeReorder:
		return fmt.Sprintf("  ↔ %s: %s", c.Path, c.Description)
	default:
		return fmt.Sprintf("    %s: %s", c.Path, c.Description)
	}
}

// CompareFields compares two sets of fields and returns the differences
func CompareFields(currentFields, schemaFields []FieldInfo) *SheetDiff {
	diff := &SheetDiff{
		FieldsToAdd:     []FieldInfo{},
		FieldsToRemove:  []FieldInfo{},
		FieldsToModify:  []FieldDiff{},
		FieldsToReorder: false,
		ExpectedOrder:   []string{},
	}

	// Create maps for easier lookup
	currentMap := make(map[string]FieldInfo)
	for _, field := range currentFields {
		currentMap[field.Name] = field
	}

	schemaMap := make(map[string]FieldInfo)
	for _, field := range schemaFields {
		schemaMap[field.Name] = field
	}

	// Find fields to add (in schema but not in sheet)
	// Preserve the position information from schema
	for _, schemaField := range schemaFields {
		if _, exists := currentMap[schemaField.Name]; !exists {
			// Keep the position from the schema for ordering
			diff.FieldsToAdd = append(diff.FieldsToAdd, schemaField)
		}
	}

	// Find fields to remove (in sheet but not in schema)
	for _, currentField := range currentFields {
		if _, exists := schemaMap[currentField.Name]; !exists {
			diff.FieldsToRemove = append(diff.FieldsToRemove, currentField)
		}
	}

	// Find fields to modify (exist in both but different)
	for _, schemaField := range schemaFields {
		if currentField, exists := currentMap[schemaField.Name]; exists {
			hasChanges := false
			fieldDiff := FieldDiff{
				Name:      schemaField.Name,
				Type:      ChangeTypeModify,
				OldType:   currentField.Type,
				NewType:   schemaField.Type,
				OldFormat: currentField.Format,
				NewFormat: schemaField.Format,
				OldHidden: currentField.Hidden,
				NewHidden: schemaField.Hidden,
			}
			
			var changes []string
			
			// Check for type/format changes
			if currentField.Type != schemaField.Type || currentField.Format != schemaField.Format {
				hasChanges = true
				changes = append(changes, fmt.Sprintf("type from %s to %s",
					formatFieldType(currentField.Type, currentField.Format),
					formatFieldType(schemaField.Type, schemaField.Format)))
			}
			
			// Check for hidden status changes
			if currentField.Hidden != schemaField.Hidden {
				hasChanges = true
				if schemaField.Hidden {
					changes = append(changes, "hide column")
				} else {
					changes = append(changes, "show column")
				}
			}
			
			if hasChanges {
				fieldDiff.Description = strings.Join(changes, ", ")
				diff.FieldsToModify = append(diff.FieldsToModify, fieldDiff)
			}
		}
	}

	// Check for field order changes
	// Build expected order from schema (only fields that exist in current)
	for _, schemaField := range schemaFields {
		if _, exists := currentMap[schemaField.Name]; exists {
			diff.ExpectedOrder = append(diff.ExpectedOrder, schemaField.Name)
		}
	}

	// Build current order
	currentOrder := []string{}
	for _, field := range currentFields {
		// Only include fields that are also in schema (ignore fields to be removed)
		if _, exists := schemaMap[field.Name]; exists {
			currentOrder = append(currentOrder, field.Name)
		}
	}

	// Compare orders
	if len(currentOrder) == len(diff.ExpectedOrder) {
		for i := range currentOrder {
			if currentOrder[i] != diff.ExpectedOrder[i] {
				diff.FieldsToReorder = true
				break
			}
		}
	}

	return diff
}

func formatFieldType(fieldType, format string) string {
	if format != "" {
		return fmt.Sprintf("%s(%s)", fieldType, format)
	}
	return fieldType
}

// ConvertDiffToResult converts a SheetDiff to a DiffResult
func ConvertDiffToResult(diff *SheetDiff, sheetName string) *DiffResult {
	return ConvertDiffToResultWithOrder(diff, sheetName, nil)
}

// ConvertDiffToResultWithOrder converts a SheetDiff to a DiffResult with optional field ordering
func ConvertDiffToResultWithOrder(diff *SheetDiff, sheetName string, schemaFields []FieldInfo) *DiffResult {
	result := &DiffResult{
		Changes:    []Change{},
		HasChanges: false,
	}

	// Create a map to store all changes by field name
	changesByField := make(map[string][]Change)
	
	// Collect all field additions
	for _, field := range diff.FieldsToAdd {
		positionInfo := ""
		if field.Position >= 0 {
			positionInfo = fmt.Sprintf(" at position %d", field.Position+1)
		}
		change := Change{
			Type:        ChangeTypeAdd,
			Path:        fmt.Sprintf("%s.%s", sheetName, field.Name),
			Description: fmt.Sprintf("Add new field '%s' of type %s%s", field.Name, formatFieldType(field.Type, field.Format), positionInfo),
			NewValue:    field,
		}
		changesByField[field.Name] = append(changesByField[field.Name], change)
		result.HasChanges = true
	}

	// Collect all field removals
	for _, field := range diff.FieldsToRemove {
		change := Change{
			Type:        ChangeTypeRemove,
			Path:        fmt.Sprintf("%s.%s", sheetName, field.Name),
			Description: fmt.Sprintf("Remove field '%s'", field.Name),
			OldValue:    field,
		}
		changesByField[field.Name] = append(changesByField[field.Name], change)
		result.HasChanges = true
	}

	// Collect all field modifications
	for _, field := range diff.FieldsToModify {
		change := Change{
			Type:        ChangeTypeModify,
			Path:        fmt.Sprintf("%s.%s", sheetName, field.Name),
			Description: field.Description,
			OldValue:    field, // Pass the entire FieldDiff object
			NewValue:    field, // Pass the entire FieldDiff object
		}
		changesByField[field.Name] = append(changesByField[field.Name], change)
		result.HasChanges = true
	}

	// If schemaFields is provided, output changes in schema field order
	if schemaFields != nil && len(schemaFields) > 0 {
		// First add changes for fields in schema order
		for _, schemaField := range schemaFields {
			if changes, exists := changesByField[schemaField.Name]; exists {
				result.Changes = append(result.Changes, changes...)
				delete(changesByField, schemaField.Name)
			}
		}
		// Then add any remaining changes (fields being removed that aren't in schema)
		for _, changes := range changesByField {
			result.Changes = append(result.Changes, changes...)
		}
	} else {
		// No schema field order provided, use original logic
		// First add additions (sorted by position)
		sortedFieldsToAdd := make([]FieldInfo, len(diff.FieldsToAdd))
		copy(sortedFieldsToAdd, diff.FieldsToAdd)
		for i := 0; i < len(sortedFieldsToAdd)-1; i++ {
			for j := i + 1; j < len(sortedFieldsToAdd); j++ {
				if sortedFieldsToAdd[i].Position > sortedFieldsToAdd[j].Position {
					sortedFieldsToAdd[i], sortedFieldsToAdd[j] = sortedFieldsToAdd[j], sortedFieldsToAdd[i]
				}
			}
		}
		for _, field := range sortedFieldsToAdd {
			if changes, exists := changesByField[field.Name]; exists {
				for _, change := range changes {
					if change.Type == ChangeTypeAdd {
						result.Changes = append(result.Changes, change)
					}
				}
			}
		}
		// Then add removals and modifications
		for _, field := range diff.FieldsToRemove {
			if changes, exists := changesByField[field.Name]; exists {
				for _, change := range changes {
					if change.Type == ChangeTypeRemove {
						result.Changes = append(result.Changes, change)
					}
				}
			}
		}
		for _, field := range diff.FieldsToModify {
			if changes, exists := changesByField[field.Name]; exists {
				for _, change := range changes {
					if change.Type == ChangeTypeModify {
						result.Changes = append(result.Changes, change)
					}
				}
			}
		}
	}

	// Add field reordering if needed (always at the end)
	if diff.FieldsToReorder {
		result.Changes = append(result.Changes, Change{
			Type:        ChangeTypeReorder,
			Path:        sheetName,
			Description: fmt.Sprintf("Reorder fields to match schema: %s", strings.Join(diff.ExpectedOrder, ", ")),
			NewValue:    diff.ExpectedOrder,
		})
		result.HasChanges = true
	}

	// Generate summary
	result.Summary = generateSummary(diff, sheetName)

	return result
}

func generateSummary(diff *SheetDiff, sheetName string) string {
	parts := []string{}

	if len(diff.FieldsToAdd) > 0 {
		parts = append(parts, fmt.Sprintf("%d field(s) to add", len(diff.FieldsToAdd)))
	}
	if len(diff.FieldsToRemove) > 0 {
		parts = append(parts, fmt.Sprintf("%d field(s) to remove", len(diff.FieldsToRemove)))
	}
	if len(diff.FieldsToModify) > 0 {
		parts = append(parts, fmt.Sprintf("%d field(s) to modify", len(diff.FieldsToModify)))
	}
	if diff.FieldsToReorder {
		parts = append(parts, "fields need reordering")
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Sheet '%s' is up to date", sheetName)
	}

	return fmt.Sprintf("Sheet '%s': %s", sheetName, strings.Join(parts, ", "))
}