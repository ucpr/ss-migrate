package engine

import (
	"fmt"
	"strings"
)

// ChangeType represents the type of change detected
type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "ADD"
	ChangeTypeRemove ChangeType = "REMOVE"
	ChangeTypeModify ChangeType = "MODIFY"
	ChangeTypeNone   ChangeType = "NONE"
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
	Description string
}

// SheetDiff represents differences in sheet structure
type SheetDiff struct {
	SheetName      string
	FieldsToAdd    []FieldInfo
	FieldsToRemove []FieldInfo
	FieldsToModify []FieldDiff
}

// FieldInfo represents basic field information
type FieldInfo struct {
	Name     string
	Type     string
	Format   string
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
	default:
		return fmt.Sprintf("    %s: %s", c.Path, c.Description)
	}
}

// CompareFields compares two sets of fields and returns the differences
func CompareFields(currentFields, schemaFields []FieldInfo) *SheetDiff {
	diff := &SheetDiff{
		FieldsToAdd:    []FieldInfo{},
		FieldsToRemove: []FieldInfo{},
		FieldsToModify: []FieldDiff{},
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
			if currentField.Type != schemaField.Type || currentField.Format != schemaField.Format {
				diff.FieldsToModify = append(diff.FieldsToModify, FieldDiff{
					Name:      schemaField.Name,
					Type:      ChangeTypeModify,
					OldType:   currentField.Type,
					NewType:   schemaField.Type,
					OldFormat: currentField.Format,
					NewFormat: schemaField.Format,
					Description: fmt.Sprintf("Type change from %s to %s", 
						formatFieldType(currentField.Type, currentField.Format),
						formatFieldType(schemaField.Type, schemaField.Format)),
				})
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
	result := &DiffResult{
		Changes:    []Change{},
		HasChanges: false,
	}

	// Sort fields to add by position to maintain schema order
	sortedFieldsToAdd := make([]FieldInfo, len(diff.FieldsToAdd))
	copy(sortedFieldsToAdd, diff.FieldsToAdd)
	for i := 0; i < len(sortedFieldsToAdd)-1; i++ {
		for j := i + 1; j < len(sortedFieldsToAdd); j++ {
			if sortedFieldsToAdd[i].Position > sortedFieldsToAdd[j].Position {
				sortedFieldsToAdd[i], sortedFieldsToAdd[j] = sortedFieldsToAdd[j], sortedFieldsToAdd[i]
			}
		}
	}

	// Add field additions in schema order
	for _, field := range sortedFieldsToAdd {
		positionInfo := ""
		if field.Position >= 0 {
			positionInfo = fmt.Sprintf(" at position %d", field.Position+1)
		}
		result.Changes = append(result.Changes, Change{
			Type:        ChangeTypeAdd,
			Path:        fmt.Sprintf("%s.%s", sheetName, field.Name),
			Description: fmt.Sprintf("Add new field '%s' of type %s%s", field.Name, formatFieldType(field.Type, field.Format), positionInfo),
			NewValue:    field,
		})
		result.HasChanges = true
	}

	// Add field removals
	for _, field := range diff.FieldsToRemove {
		result.Changes = append(result.Changes, Change{
			Type:        ChangeTypeRemove,
			Path:        fmt.Sprintf("%s.%s", sheetName, field.Name),
			Description: fmt.Sprintf("Remove field '%s'", field.Name),
			OldValue:    field,
		})
		result.HasChanges = true
	}

	// Add field modifications
	for _, field := range diff.FieldsToModify {
		result.Changes = append(result.Changes, Change{
			Type:        ChangeTypeModify,
			Path:        fmt.Sprintf("%s.%s", sheetName, field.Name),
			Description: field.Description,
			OldValue:    fmt.Sprintf("%s(%s)", field.OldType, field.OldFormat),
			NewValue:    fmt.Sprintf("%s(%s)", field.NewType, field.NewFormat),
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

	if len(parts) == 0 {
		return fmt.Sprintf("Sheet '%s' is up to date", sheetName)
	}

	return fmt.Sprintf("Sheet '%s': %s", sheetName, strings.Join(parts, ", "))
}