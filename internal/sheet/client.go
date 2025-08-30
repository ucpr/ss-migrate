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
	service *sheets.Service
}

// NewClient creates a new Google Sheets client using Application Default Credentials (ADC)
func NewClient(ctx context.Context) (*Client, error) {
	service, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &Client{
		service: service,
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
	spreadsheet, err := c.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}
	return spreadsheet, nil
}

// GetValues retrieves values from a specific range
func (c *Client) GetValues(ctx context.Context, spreadsheetID, readRange string) ([][]any, error) {
	resp, err := c.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
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

	_, err := c.service.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
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

	_, err := c.service.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to batch update values: %w", err)
	}

	return nil
}

// ClearValues clears values in a specific range
func (c *Client) ClearValues(ctx context.Context, spreadsheetID, clearRange string) error {
	clearRequest := &sheets.ClearValuesRequest{}

	_, err := c.service.Spreadsheets.Values.Clear(spreadsheetID, clearRange, clearRequest).Context(ctx).Do()
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

	_, err := c.service.Spreadsheets.Values.Append(spreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to append values: %w", err)
	}

	return nil
}
