package imageio

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/woozymasta/bcn"
	"github.com/woozymasta/edds"
)

func TestWriteWithOptionsEDDSCompressed(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 16), //nolint:gosec // bounded 0..112
				G: uint8(y * 16), //nolint:gosec // bounded 0..112
				B: 128,
				A: 255,
			})
		}
	}

	path := filepath.Join(t.TempDir(), "atlas.edds")
	err := WriteWithOptions(path, img, &EncodeSettings{
		Format:  mustParseFormat(t, "dxt1"),
		Quality: 8,
		Mipmaps: 1,
	})
	if err != nil {
		t.Fatalf("WriteWithOptions error: %v", err)
	}

	cfg, err := edds.ReadConfig(path)
	if err != nil {
		t.Fatalf("ReadConfig error: %v", err)
	}
	if cfg.Width != 8 || cfg.Height != 8 {
		t.Fatalf("ReadConfig size = %dx%d, want 8x8", cfg.Width, cfg.Height)
	}
}

func mustParseFormat(t *testing.T, s string) bcn.Format {
	t.Helper()
	f, err := ParseOutputFormat(s)
	if err != nil {
		t.Fatalf("ParseOutputFormat(%q): %v", s, err)
	}
	return f
}
