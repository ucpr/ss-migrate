package engine

import (
	"context"
	"fmt"

	"github.com/ucpr/ss-migrate/internal/schema"
	"github.com/ucpr/ss-migrate/internal/sheet"
)

// Planner handles planning migrations between sheet and schema
type Planner struct {
	sheetClient *sheet.Client
}

// NewPlanner creates a new planner instance
func NewPlanner(sheetClient *sheet.Client) *Planner {
	return &Planner{
		sheetClient: sheetClient,
	}
}

// Plan generates a migration plan by comparing sheet with schema
func (p *Planner) Plan(ctx context.Context, schemaConfig *schema.Schema) (*DiffResult, error) {
	if len(schemaConfig.Resources) == 0 {
		return nil, fmt.Errorf("no resources defined in schema")
	}

	// For now, we'll handle the first resource
	resource := schemaConfig.Resources[0]

	// Extract spreadsheet ID from URL
	spreadsheetID, err := sheet.ExtractSpreadsheetID(resource.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to extract spreadsheet ID: %w", err)
	}

	// Get current sheet structure
	currentFields, err := p.analyzeSheet(ctx, spreadsheetID, resource.Name, resource.HeaderRow)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze sheet: %w", err)
	}

	// Convert schema fields to FieldInfo
	schemaFields := convertSchemaFields(resource.Fields)

	// Compare fields
	diff := CompareFields(currentFields, schemaFields)
	diff.SheetName = resource.Name

	// Convert to result
	result := ConvertDiffToResult(diff, resource.Name)

	return result, nil
}

// analyzeSheet analyzes the current structure of a sheet
func (p *Planner) analyzeSheet(ctx context.Context, spreadsheetID, sheetName string, headerRow int) ([]FieldInfo, error) {
	if headerRow == 0 {
		headerRow = 1
	}

	// Get headers from the sheet
	headers, err := p.sheetClient.GetHeaders(ctx, spreadsheetID, sheetName, headerRow)
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	// For each header, analyze the column data to infer type
	fields := []FieldInfo{}
	for i, header := range headers {
		if header == "" {
			continue
		}

		// Get sample data from the column
		column := sheet.ColumnToLetter(i)
		columnData, err := p.sheetClient.GetColumnData(ctx, spreadsheetID, sheetName, column, headerRow+1)
		if err != nil {
			// If we can't get data, assume string type
			fields = append(fields, FieldInfo{
				Name: header,
				Type: "string",
			})
			continue
		}

		// Infer type from data
		inferredType := sheet.InferColumnType(columnData)
		fields = append(fields, FieldInfo{
			Name: header,
			Type: inferredType,
		})
	}

	return fields, nil
}

// convertSchemaFields converts schema fields to FieldInfo with position information
func convertSchemaFields(fields []schema.Field) []FieldInfo {
	result := []FieldInfo{}
	for i, field := range fields {
		info := FieldInfo{
			Name:     field.Name,
			Type:     field.Type,
			Format:   field.Format,
			Hidden:   field.Hidden,
			Position: i, // Store the position in the schema
		}
		result = append(result, info)
	}
	return result
}

// PlanAll generates migration plans for all resources in the schema
func (p *Planner) PlanAll(ctx context.Context, schemaConfig *schema.Schema) ([]*DiffResult, error) {
	results := []*DiffResult{}

	for _, resource := range schemaConfig.Resources {
		// Extract spreadsheet ID from URL
		spreadsheetID, err := sheet.ExtractSpreadsheetID(resource.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to extract spreadsheet ID for resource %s: %w", resource.Name, err)
		}

		// Get current sheet structure
		currentFields, err := p.analyzeSheet(ctx, spreadsheetID, resource.Name, resource.HeaderRow)
		if err != nil {
			// If sheet doesn't exist, treat as all fields need to be added
			currentFields = []FieldInfo{}
		}

		// Convert schema fields to FieldInfo
		schemaFields := convertSchemaFields(resource.Fields)

		// Compare fields
		diff := CompareFields(currentFields, schemaFields)
		diff.SheetName = resource.Name

		// Convert to result
		result := ConvertDiffToResult(diff, resource.Name)
		results = append(results, result)
	}

	return results, nil
}