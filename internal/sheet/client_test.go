package sheet

import (
	"testing"
)

func TestExtractSpreadsheetID(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid Google Sheets URL",
			url:     "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit",
			want:    "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			wantErr: false,
		},
		{
			name:    "valid URL with gid parameter",
			url:     "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit#gid=0",
			want:    "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			wantErr: false,
		},
		{
			name:    "invalid URL - not Google Sheets",
			url:     "https://example.com/spreadsheet",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid URL - no spreadsheet ID",
			url:     "https://docs.google.com/spreadsheets",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			url:     "not-a-url",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSpreadsheetID(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSpreadsheetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractSpreadsheetID() = %v, want %v", got, tt.want)
			}
		})
	}
}
