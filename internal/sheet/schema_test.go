package sheet

import (
	"testing"
)

func TestInferColumnType(t *testing.T) {
	tests := []struct {
		name string
		data []any
		want string
	}{
		{
			name: "integer column",
			data: []any{1, 2, 3, 4, 5},
			want: "integer",
		},
		{
			name: "float column",
			data: []any{1.5, 2.0, 3.14},
			want: "number",
		},
		{
			name: "string column",
			data: []any{"apple", "banana", "cherry"},
			want: "string",
		},
		{
			name: "boolean column",
			data: []any{"true", "false", "TRUE", "FALSE"},
			want: "boolean",
		},
		{
			name: "mixed with strings",
			data: []any{1, "two", 3},
			want: "string",
		},
		{
			name: "empty data",
			data: []any{},
			want: "string",
		},
		{
			name: "with nil values",
			data: []any{1, nil, 3, nil, 5},
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

func TestColumnToLetter(t *testing.T) {
	tests := []struct {
		col  int
		want string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
		{701, "ZZ"},
		{702, "AAA"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := ColumnToLetter(tt.col); got != tt.want {
				t.Errorf("ColumnToLetter(%d) = %v, want %v", tt.col, got, tt.want)
			}
		})
	}
}

func TestLetterToColumn(t *testing.T) {
	tests := []struct {
		letter string
		want   int
	}{
		{"A", 0},
		{"B", 1},
		{"Z", 25},
		{"AA", 26},
		{"AB", 27},
		{"AZ", 51},
		{"BA", 52},
		{"ZZ", 701},
		{"AAA", 702},
	}

	for _, tt := range tests {
		t.Run(tt.letter, func(t *testing.T) {
			if got := LetterToColumn(tt.letter); got != tt.want {
				t.Errorf("LetterToColumn(%s) = %v, want %v", tt.letter, got, tt.want)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"123.45", true},
		{"-123", true},
		{"+123", true},
		{"1,234", true},
		{"abc", false},
		{"12a34", false},
		{"", false},
		{"1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isNumeric(tt.input); got != tt.want {
				t.Errorf("isNumeric(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHasDecimal(t *testing.T) {
	tests := []struct {
		name string
		data []any
		want bool
	}{
		{
			name: "has decimal",
			data: []any{1, 2.5, 3},
			want: true,
		},
		{
			name: "no decimal",
			data: []any{1, 2, 3},
			want: false,
		},
		{
			name: "decimal as string",
			data: []any{"1", "2.5", "3"},
			want: true,
		},
		{
			name: "with nil",
			data: []any{1, nil, 3},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasDecimal(tt.data); got != tt.want {
				t.Errorf("hasDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}
