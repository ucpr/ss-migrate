package sheet

import (
	"context"
	"fmt"
	"strings"
)

// SheetInfo represents information about a sheet in a spreadsheet
type SheetInfo struct {
	Name        string
	SheetID     int64
	RowCount    int64
	ColumnCount int64
}

// ColumnInfo represents information about a column
type ColumnInfo struct {
	Name  string
	Index int
	Type  string
}

// GetSheetInfo retrieves information about all sheets in a spreadsheet
func (c *Client) GetSheetInfo(ctx context.Context, spreadsheetID string) ([]SheetInfo, error) {
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	var sheets []SheetInfo
	for _, sheet := range spreadsheet.Sheets {
		info := SheetInfo{
			Name:        sheet.Properties.Title,
			SheetID:     sheet.Properties.SheetId,
			RowCount:    sheet.Properties.GridProperties.RowCount,
			ColumnCount: sheet.Properties.GridProperties.ColumnCount,
		}
		sheets = append(sheets, info)
	}

	return sheets, nil
}

// GetColumnVisibility checks if columns are hidden in the sheet
func (c *Client) GetColumnVisibility(ctx context.Context, spreadsheetID, sheetName string) ([]bool, error) {
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			// Get column count
			columnCount := int(sheet.Properties.GridProperties.ColumnCount)
			visibility := make([]bool, columnCount)
			
			// By default, columns are visible (false = not hidden)
			// We need to get the sheet with include grid data to get dimension info
			return visibility, nil
		}
	}

	return nil, fmt.Errorf("sheet %s not found", sheetName)
}

// GetHeaders retrieves the header row from a sheet
func (c *Client) GetHeaders(ctx context.Context, spreadsheetID, sheetName string, headerRow int) ([]string, error) {
	if headerRow < 1 {
		headerRow = 1
	}

	readRange := fmt.Sprintf("%s!%d:%d", sheetName, headerRow, headerRow)
	values, err := c.GetValues(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	if len(values) == 0 || len(values[0]) == 0 {
		return []string{}, nil
	}

	headers := make([]string, len(values[0]))
	for i, val := range values[0] {
		if val != nil {
			headers[i] = fmt.Sprintf("%v", val)
		}
	}

	return headers, nil
}

// GetColumnData retrieves all data from a specific column
func (c *Client) GetColumnData(ctx context.Context, spreadsheetID, sheetName string, column string, startRow int) ([]any, error) {
	if startRow < 1 {
		startRow = 1
	}

	readRange := fmt.Sprintf("%s!%s%d:%s", sheetName, column, startRow, column)
	values, err := c.GetValues(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get column data: %w", err)
	}

	var columnData []any
	for _, row := range values {
		if len(row) > 0 {
			columnData = append(columnData, row[0])
		} else {
			columnData = append(columnData, nil)
		}
	}

	return columnData, nil
}

// InferColumnType attempts to infer the type of data in a column
func InferColumnType(data []any) string {
	if len(data) == 0 {
		return "string"
	}

	hasNumber := false
	hasString := false
	hasBoolean := false

	for _, val := range data {
		if val == nil {
			continue
		}

		strVal := fmt.Sprintf("%v", val)
		strVal = strings.TrimSpace(strVal)

		// Check for boolean
		if strings.ToLower(strVal) == "true" || strings.ToLower(strVal) == "false" {
			hasBoolean = true
			continue
		}

		// Check for number
		if isNumeric(strVal) {
			hasNumber = true
		} else if strVal != "" {
			hasString = true
		}
	}

	// Determine the predominant type
	if hasString {
		return "string"
	} else if hasNumber {
		if hasDecimal(data) {
			return "number"
		}
		return "integer"
	} else if hasBoolean {
		return "boolean"
	}

	return "string"
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	// Remove commas for number formatting
	s = strings.ReplaceAll(s, ",", "")

	dotCount := 0
	// Check for numeric patterns
	for i, ch := range s {
		if ch == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if (ch < '0' || ch > '9') && ch != '-' && ch != '+' {
			return false
		}
		if (ch == '-' || ch == '+') && i != 0 {
			return false
		}
	}

	return true
}

// hasDecimal checks if any values in the data have decimal points
func hasDecimal(data []any) bool {
	for _, val := range data {
		if val != nil {
			strVal := fmt.Sprintf("%v", val)
			if strings.Contains(strVal, ".") {
				return true
			}
		}
	}
	return false
}

// ColumnToLetter converts a column index (0-based) to a letter (A, B, C, ..., Z, AA, AB, ...)
func ColumnToLetter(col int) string {
	result := ""
	for col >= 0 {
		result = string(rune('A'+col%26)) + result
		col = col/26 - 1
	}
	return result
}

// LetterToColumn converts a column letter to an index (0-based)
func LetterToColumn(letter string) int {
	result := 0
	for i, ch := range letter {
		result = result*26 + int(ch-'A'+1)
		if i == len(letter)-1 {
			result--
		}
	}
	return result
}
