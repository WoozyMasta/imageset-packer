package imageio

import (
	"fmt"
	"strings"

	"github.com/woozymasta/bcn"
)

// EncodeSettings configures DDS/EDDS output encoding.
type EncodeSettings struct {
	// Format selects output pixel format. Zero value means BGRA8 default.
	Format bcn.Format
	// Quality controls BCn quality: 0 = library default, 1..10 = explicit levels.
	Quality int
	// Mipmaps limits written mip levels for EDDS: 0 = full chain, 1 = base only.
	Mipmaps int
}

// ParseOutputFormat parses a textual output format alias.
func ParseOutputFormat(s string) (bcn.Format, error) {
	v := normalizeFormatAlias(s)
	if v == "" {
		return bcn.FormatBGRA8, nil
	}

	switch v {
	case "bgra8", "b8g8r8a8unorm":
		return bcn.FormatBGRA8, nil
	case "dxt1", "bc1":
		return bcn.FormatDXT1, nil
	case "dxt5", "bc3":
		return bcn.FormatDXT5, nil
	default:
		return bcn.FormatUnknown, fmt.Errorf(
			"unknown format %q (supported: bgra8, dxt1, dxt5)",
			s,
		)
	}
}

// ValidateQualityLevel validates BCn quality.
func ValidateQualityLevel(q int) error {
	if q < 0 || q > 10 {
		return fmt.Errorf("quality must be in range 0..10, got %d", q)
	}

	return nil
}

// normalizeFormatAlias normalizes a format alias.
func normalizeFormatAlias(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.TrimPrefix(s, "dxgiformat")
	return s
}

// effectiveEncodeSettings returns the effective encode settings.
func effectiveEncodeSettings(opts *EncodeSettings) EncodeSettings {
	e := EncodeSettings{
		Format:  bcn.FormatBGRA8,
		Quality: 0,
		Mipmaps: 0,
	}
	if opts == nil {
		return e
	}

	if opts.Format != bcn.FormatUnknown {
		e.Format = opts.Format
	}
	e.Quality = opts.Quality
	e.Mipmaps = opts.Mipmaps

	return e
}
