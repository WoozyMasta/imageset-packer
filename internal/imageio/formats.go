package imageio

import (
	"fmt"
	"strconv"
	"strings"
)

// RGB stores an 8-bit per channel color.
type RGB struct{ R, G, B uint8 }

// ParseHexRGB parses a 6-digit hex RGB string (with or without leading '#').
func ParseHexRGB(s string) (RGB, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return RGB{}, fmt.Errorf("expected 6 hex chars, got %q", s)
	}

	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return RGB{}, err
	}

	return RGB{
		R: uint8((v >> 16) & 0xff), //nolint:gosec // Masked to 8 bits.
		G: uint8((v >> 8) & 0xff),  //nolint:gosec // Masked to 8 bits.
		B: uint8(v & 0xff),         //nolint:gosec // Masked to 8 bits.
	}, nil
}
