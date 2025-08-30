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

	// Create move dimension request
	req := &sheets.Request{
		MoveDimension: &sheets.MoveDimensionRequest{
			Source: &sheets.DimensionRange{
				SheetId:    sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(sourceIndex),
				EndIndex:   int64(sourceIndex + 1),
			},
			DestinationIndex: int64(destinationIndex),
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
