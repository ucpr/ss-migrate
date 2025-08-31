package sheet

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Client struct {
	Service *sheets.Service
}

// NewClient creates a new Google Sheets client using Application Default Credentials (ADC)
func NewClient(ctx context.Context) (*Client, error) {
	service, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &Client{
		Service: service,
	}, nil
}

// ExtractSpreadsheetID extracts the spreadsheet ID from a Google Sheets URL
func ExtractSpreadsheetID(sheetURL string) (string, error) {
	parsedURL, err := url.Parse(sheetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a Google Sheets URL
	if !strings.Contains(parsedURL.Host, "docs.google.com") {
		return "", fmt.Errorf("not a Google Sheets URL")
	}

	// Extract ID from path: /spreadsheets/d/{ID}/...
	parts := strings.Split(parsedURL.Path, "/")
	for i, part := range parts {
		if part == "d" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return "", fmt.Errorf("spreadsheet ID not found in URL")
}

// GetSpreadsheet retrieves spreadsheet metadata
func (c *Client) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	spreadsheet, err := c.Service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}
	return spreadsheet, nil
}

// GetValues retrieves values from a specific range
func (c *Client) GetValues(ctx context.Context, spreadsheetID, readRange string) ([][]any, error) {
	resp, err := c.Service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get values: %w", err)
	}
	return resp.Values, nil
}

// UpdateValues updates values in a specific range
func (c *Client) UpdateValues(ctx context.Context, spreadsheetID, writeRange string, values [][]any) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.Service.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to update values: %w", err)
	}

	return nil
}

// BatchUpdateValues updates multiple ranges at once
func (c *Client) BatchUpdateValues(ctx context.Context, spreadsheetID string, data []*sheets.ValueRange) error {
	batchUpdateRequest := &sheets.BatchUpdateValuesRequest{
		Data:             data,
		ValueInputOption: "USER_ENTERED",
	}

	_, err := c.Service.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to batch update values: %w", err)
	}

	return nil
}

// ClearValues clears values in a specific range
func (c *Client) ClearValues(ctx context.Context, spreadsheetID, clearRange string) error {
	clearRequest := &sheets.ClearValuesRequest{}

	_, err := c.Service.Spreadsheets.Values.Clear(spreadsheetID, clearRange, clearRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to clear values: %w", err)
	}

	return nil
}

// AppendValues appends values to the end of a sheet
func (c *Client) AppendValues(ctx context.Context, spreadsheetID, appendRange string, values [][]any) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.Service.Spreadsheets.Values.Append(spreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to append values: %w", err)
	}

	return nil
}

// CreateSheet creates a new sheet in the spreadsheet
func (c *Client) CreateSheet(ctx context.Context, spreadsheetID, sheetName string) error {
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: sheetName,
			},
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err := c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	return nil
}

// CheckSheetExists checks if a sheet exists in the spreadsheet
func (c *Client) CheckSheetExists(ctx context.Context, spreadsheetID, sheetName string) (bool, error) {
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return false, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return true, nil
		}
	}

	return false, nil
}

// DeleteColumn deletes a column at the specified index
func (c *Client) DeleteColumn(ctx context.Context, spreadsheetID, sheetName string, columnIndex int) error {
	// Get sheet ID
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	if sheetID == -1 {
		return fmt.Errorf("sheet %s not found", sheetName)
	}

	// Create delete dimension request
	req := &sheets.Request{
		DeleteDimension: &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(columnIndex),
				EndIndex:   int64(columnIndex + 1),
			},
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}

	return nil
}

// MoveColumn moves a column from one position to another
func (c *Client) MoveColumn(ctx context.Context, spreadsheetID, sheetName string, sourceIndex, destinationIndex int) error {
	// Get sheet ID
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	if sheetID == -1 {
		return fmt.Errorf("sheet %s not found", sheetName)
	}

	// Adjust destination index based on Google Sheets API behavior:
	// When moving a column to the right (sourceIndex < destinationIndex),
	// the API expects the destination index to be one more than the visual position
	// because the source column will be removed first.
	adjustedDestination := destinationIndex
	if sourceIndex < destinationIndex {
		adjustedDestination = destinationIndex + 1
	}

	// Create move dimension request
	req := &sheets.Request{
		MoveDimension: &sheets.MoveDimensionRequest{
			Source: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(sourceIndex),
				EndIndex:   int64(sourceIndex + 1),
			},
			DestinationIndex: int64(adjustedDestination),
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to move column: %w", err)
	}

	return nil
}

// HideColumn hides a column at the specified index
func (c *Client) HideColumn(ctx context.Context, spreadsheetID, sheetName string, columnIndex int) error {
	// Get sheet ID
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	if sheetID == -1 {
		return fmt.Errorf("sheet %s not found", sheetName)
	}

	// Create update dimension properties request to hide the column
	req := &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(columnIndex),
				EndIndex:   int64(columnIndex + 1),
			},
			Properties: &sheets.DimensionProperties{
				HiddenByUser: true,
			},
			Fields: "hiddenByUser",
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to hide column: %w", err)
	}

	return nil
}

// ShowColumn shows a hidden column at the specified index
func (c *Client) ShowColumn(ctx context.Context, spreadsheetID, sheetName string, columnIndex int) error {
	// Get sheet ID
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	if sheetID == -1 {
		return fmt.Errorf("sheet %s not found", sheetName)
	}

	// Create update dimension properties request to show the column
	req := &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(columnIndex),
				EndIndex:   int64(columnIndex + 1),
			},
			Properties: &sheets.DimensionProperties{
				HiddenByUser: false,
			},
			Fields: "hiddenByUser",
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to show column: %w", err)
	}

	return nil
}

// InsertColumn inserts a new column at the specified index
func (c *Client) InsertColumn(ctx context.Context, spreadsheetID, sheetName string, columnIndex int) error {
	// Get sheet ID
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	if sheetID == -1 {
		return fmt.Errorf("sheet %s not found", sheetName)
	}

	// Create insert dimension request
	req := &sheets.Request{
		InsertDimension: &sheets.InsertDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(columnIndex),
				EndIndex:   int64(columnIndex + 1),
			},
			InheritFromBefore: false,
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to insert column: %w", err)
	}

	return nil
}

// GetColumnFormat retrieves the number format pattern of a column
func (c *Client) GetColumnFormat(ctx context.Context, spreadsheetID, sheetName string, columnIndex int) (string, error) {
	// Get spreadsheet with cell format data
	spreadsheet, err := c.Service.Spreadsheets.Get(spreadsheetID).
		Ranges(fmt.Sprintf("%s!%s2:%s2", sheetName, ColumnToLetter(columnIndex), ColumnToLetter(columnIndex))).
		IncludeGridData(true).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	// Find the sheet
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			// Check if we have grid data
			if len(sheet.Data) > 0 && len(sheet.Data[0].RowData) > 0 && len(sheet.Data[0].RowData[0].Values) > 0 {
				cellData := sheet.Data[0].RowData[0].Values[0]
				if cellData.UserEnteredFormat != nil && cellData.UserEnteredFormat.NumberFormat != nil {
					return cellData.UserEnteredFormat.NumberFormat.Pattern, nil
				}
				if cellData.EffectiveFormat != nil && cellData.EffectiveFormat.NumberFormat != nil {
					return cellData.EffectiveFormat.NumberFormat.Pattern, nil
				}
			}
			break
		}
	}

	return "", nil // No format found
}

// FormatColumn applies number formatting to a column based on the data type
func (c *Client) FormatColumn(ctx context.Context, spreadsheetID, sheetName string, columnIndex int, dataType, format string) error {
	// Get sheet ID
	spreadsheet, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	var maxRows int64 = 1000 // Default max rows
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			if sheet.Properties.GridProperties != nil && sheet.Properties.GridProperties.RowCount > 0 {
				maxRows = sheet.Properties.GridProperties.RowCount
			}
			break
		}
	}

	if sheetID == -1 {
		return fmt.Errorf("sheet %s not found", sheetName)
	}

	// Determine the number format pattern based on type
	var pattern string
	switch dataType {
	case "integer":
		pattern = "0" // No decimal places
	case "number":
		pattern = "0.00" // Two decimal places
	case "datetime":
		if format == "default" || format == "" {
			pattern = "yyyy-mm-dd hh:mm:ss"
		} else if format == "date" {
			pattern = "yyyy-mm-dd"
		} else if format == "time" {
			pattern = "hh:mm:ss"
		} else {
			pattern = "yyyy-mm-dd hh:mm:ss"
		}
	case "boolean":
		// Boolean doesn't need number formatting, skip
		return nil
	case "string":
		// String doesn't need number formatting, but we'll clear any existing format
		pattern = "@" // Text format
	default:
		// No specific formatting needed
		return nil
	}

	// Create format request for the entire column (excluding header)
	req := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    1, // Skip header row
				EndRowIndex:      maxRows,
				StartColumnIndex: int64(columnIndex),
				EndColumnIndex:   int64(columnIndex + 1),
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					NumberFormat: &sheets.NumberFormat{
						Type:    "NUMBER",
						Pattern: pattern,
					},
				},
			},
			Fields: "userEnteredFormat.numberFormat",
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = c.Service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to format column: %w", err)
	}

	return nil
}
