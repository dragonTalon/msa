package utils

import (
	"testing"
)

// TestPrettyJSON tests the PrettyJSON function with various inputs
func TestPrettyJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid object",
			input:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "complex nested structure",
			input:   map[string]interface{}{"a": []interface{}{1, 2, "b"}, "c": map[string]int{"d": 3}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("PrettyJSON() returned empty string for valid input")
			}
		})
	}
}

// TestCompactJSON tests the CompactJSON function
func TestCompactJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid object",
			input:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompactJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompactJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("CompactJSON() returned empty string for valid input")
			}
		})
	}
}

// TestValidateJSON tests the ValidateJSON function
func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid JSON object",
			input: `{"key": "value"}`,
			want:  true,
		},
		{
			name:  "valid JSON array",
			input: `[1, 2, 3]`,
			want:  true,
		},
		{
			name:  "valid JSON string",
			input: `"hello"`,
			want:  true,
		},
		{
			name:  "valid JSON number",
			input: `42`,
			want:  true,
		},
		{
			name:  "valid JSON boolean",
			input: `true`,
			want:  true,
		},
		{
			name:  "valid JSON null",
			input: `null`,
			want:  true,
		},
		{
			name:  "invalid JSON - missing quote",
			input: `{key: "value"}`,
			want:  false,
		},
		{
			name:  "invalid JSON - missing closing brace",
			input: `{"key": "value"`,
			want:  false,
		},
		{
			name:  "invalid JSON - random string",
			input: `not json at all`,
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateJSON(tt.input); got != tt.want {
				t.Errorf("ValidateJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseJSON tests the ParseJSON function
func TestParseJSON(t *testing.T) {
	type TestStruct struct {
		Key string `json:"key"`
	}
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON string",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{not valid}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := ParseJSON(tt.input, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestToJSONString tests the ToJSONString function
func TestToJSONString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "simple object",
			input: map[string]string{"key": "value"},
			want:  `{"key":"value"}`,
		},
		{
			name:  "array",
			input: []int{1, 2, 3},
			want:  `[1,2,3]`,
		},
		{
			name:  "nil",
			input: nil,
			want:  `null`,
		},
		{
			name:  "empty object",
			input: map[string]string{},
			want:  `{}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToJSONString(tt.input); got != tt.want {
				t.Errorf("ToJSONString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToJSONBytes tests the ToJSONBytes function
func TestToJSONBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "simple object",
			input:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "array",
			input:   []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToJSONBytes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSONBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("ToJSONBytes() returned empty bytes for valid input")
			}
		})
	}
}

// TestTruncateString tests the TruncateString function
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "string shorter than max",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "string equal to max",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "string longer than max",
			input:  "hello world",
			maxLen: 5,
			want:   "hello...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 5,
			want:   "",
		},
		{
			name:   "single character",
			input:  "a",
			maxLen: 1,
			want:   "a",
		},
		{
			name:   "unicode characters",
			input:  "你好世界",
			maxLen: 6, // Each Chinese character is 3 bytes in UTF-8
			want:   "你好...",
		},
		{
			name:   "zero max length",
			input:  "hello",
			maxLen: 0,
			want:   "...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateString(tt.input, tt.maxLen); got != tt.want {
				t.Errorf("TruncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatArray tests the FormatArray function
func TestFormatArray(t *testing.T) {
	tests := []struct {
		name      string
		arr       []string
		separator string
		want      string
	}{
		{
			name:      "comma separated",
			arr:       []string{"a", "b", "c"},
			separator: ",",
			want:      "a,b,c",
		},
		{
			name:      "dash separated",
			arr:       []string{"x", "y", "z"},
			separator: "-",
			want:      "x-y-z",
		},
		{
			name:      "empty array",
			arr:       []string{},
			separator: ",",
			want:      "",
		},
		{
			name:      "nil array",
			arr:       nil,
			separator: ",",
			want:      "",
		},
		{
			name:      "single element",
			arr:       []string{"only"},
			separator: ",",
			want:      "only",
		},
		{
			name:      "empty separator",
			arr:       []string{"a", "b"},
			separator: "",
			want:      "ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatArray(tt.arr, tt.separator); got != tt.want {
				t.Errorf("FormatArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatObject tests the FormatObject function
func TestFormatObject(t *testing.T) {
	tests := []struct {
		name    string
		obj     interface{}
		pretty  bool
		wantErr bool
	}{
		{
			name:    "pretty format",
			obj:     map[string]string{"key": "value"},
			pretty:  true,
			wantErr: false,
		},
		{
			name:    "compact format",
			obj:     map[string]string{"key": "value"},
			pretty:  false,
			wantErr: false,
		},
		{
			name:    "nil object",
			obj:     nil,
			pretty:  true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatObject(tt.obj, tt.pretty)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("FormatObject() returned empty string for valid input")
			}
		})
	}
}

// TestPrettyJSONBytes tests the PrettyJSONBytes function
func TestPrettyJSONBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid JSON bytes",
			input:   []byte(`{"key":"value"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON bytes",
			input:   []byte(`{invalid}`),
			wantErr: true,
		},
		{
			name:    "nil bytes",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "empty bytes",
			input:   []byte{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyJSONBytes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJSONBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("PrettyJSONBytes() returned empty string for valid input")
			}
		})
	}
}
