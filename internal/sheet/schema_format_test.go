package sheet

import (
	"testing"
)

func TestInferTypeFromFormat(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		// Integer patterns
		{"integer no decimal", "0", "integer"},
		{"integer with thousands", "#,##0", "integer"},
		{"integer with hash", "#,###", "integer"},
		{"integer with optional decimals", "0.#####", "integer"},
		
		// Number patterns
		{"number two decimals", "0.00", "number"},
		{"number with thousands", "#,##0.00", "number"},
		{"number one decimal", "0.0", "number"},
		{"percentage", "0.00%", "number"},
		{"currency USD", "$#,##0.00", "number"},
		{"currency JPY", "¥#,##0", "number"},
		{"currency EUR", "€#,##0.00", "number"},
		
		// String patterns
		{"text format", "@", "string"},
		
		// DateTime patterns
		{"date ISO", "yyyy-mm-dd", "datetime"},
		{"datetime full", "yyyy-mm-dd hh:mm:ss", "datetime"},
		{"date US", "mm/dd/yyyy", "datetime"},
		{"time only", "hh:mm:ss", "datetime"},
		{"date with time", "dd/mm/yyyy hh:mm", "datetime"},
		
		// Empty or unknown
		{"empty pattern", "", ""},
		{"unknown pattern", "###", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferTypeFromFormat(tt.pattern); got != tt.want {
				t.Errorf("InferTypeFromFormat(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}