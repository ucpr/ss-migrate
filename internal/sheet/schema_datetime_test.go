package sheet

import (
	"testing"
)

func TestInferColumnType_DateTime(t *testing.T) {
	tests := []struct {
		name string
		data []any
		want string
	}{
		{
			name: "ISO 8601 date",
			data: []any{"2024-01-15", "2024-02-20", "2024-03-25"},
			want: "datetime",
		},
		{
			name: "ISO 8601 datetime",
			data: []any{"2024-01-15T10:30:00", "2024-02-20T14:45:00", "2024-03-25T09:15:00"},
			want: "datetime",
		},
		{
			name: "datetime with space",
			data: []any{"2024-01-15 10:30:00", "2024-02-20 14:45:00", "2024-03-25 09:15:00"},
			want: "datetime",
		},
		{
			name: "US format date",
			data: []any{"01/15/2024", "02/20/2024", "03/25/2024"},
			want: "datetime",
		},
		{
			name: "RFC3339 with timezone",
			data: []any{"2024-01-15T10:30:00Z", "2024-02-20T14:45:00+09:00", "2024-03-25T09:15:00-05:00"},
			want: "datetime",
		},
		{
			name: "mixed datetime and nil",
			data: []any{"2024-01-15", nil, "2024-03-25", nil},
			want: "datetime",
		},
		{
			name: "mixed datetime and other types",
			data: []any{"2024-01-15", "not a date", "2024-03-25"},
			want: "string",
		},
		{
			name: "integer that looks like date",
			data: []any{"20240115", "20240220", "20240325"},
			want: "integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferColumnType(tt.data); got != tt.want {
				t.Errorf("InferColumnType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDateTime(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"ISO date", "2024-01-15", true},
		{"ISO datetime", "2024-01-15T10:30:00", true},
		{"datetime with space", "2024-01-15 10:30:00", true},
		{"US format", "01/15/2024", true},
		{"US format short", "1/2/2024", true},
		{"RFC3339 UTC", "2024-01-15T10:30:00Z", true},
		{"RFC3339 timezone", "2024-01-15T10:30:00+09:00", true},
		{"not a date", "hello world", false},
		{"just numbers", "12345", false},
		{"empty string", "", false},
		{"year only", "2024", false},
		{"partial date", "2024-01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDateTime(tt.input); got != tt.want {
				t.Errorf("isDateTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}