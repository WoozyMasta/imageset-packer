package imageio

import (
	"testing"

	"github.com/woozymasta/bcn"
)

func TestParseOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bcn.Format
	}{
		{name: "default-empty", input: "", want: bcn.FormatBGRA8},
		{name: "bgra8", input: "bgra8", want: bcn.FormatBGRA8},
		{name: "dxgi-alias", input: "DXGI_FORMAT_B8G8R8A8_UNORM", want: bcn.FormatBGRA8},
		{name: "dxt1", input: "dxt1", want: bcn.FormatDXT1},
		{name: "bc3", input: "bc3", want: bcn.FormatDXT5},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseOutputFormat(tc.input)
			if err != nil {
				t.Fatalf("ParseOutputFormat(%q) error = %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("ParseOutputFormat(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseOutputFormatUnknown(t *testing.T) {
	t.Parallel()

	_, err := ParseOutputFormat("foo")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestValidateQualityLevel(t *testing.T) {
	t.Parallel()

	valid := []int{0, 1, 6, 10}
	for _, q := range valid {
		if err := ValidateQualityLevel(q); err != nil {
			t.Fatalf("ValidateQualityLevel(%d) unexpected error: %v", q, err)
		}
	}

	invalid := []int{-1, 11}
	for _, q := range invalid {
		if err := ValidateQualityLevel(q); err == nil {
			t.Fatalf("ValidateQualityLevel(%d) expected error", q)
		}
	}
}
