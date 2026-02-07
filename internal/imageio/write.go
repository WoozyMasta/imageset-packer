package imageio

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/schwarzlichtbezirk/tga"
	"github.com/woozymasta/bcn"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/woozymasta/edds"
)

// Write saves an image to the given path based on its extension.
func Write(path string, img image.Image) error {
	return WriteWithOptions(path, img, nil)
}

// WriteWithOptions saves an image using optional DDS/EDDS encoding settings.
func WriteWithOptions(path string, img image.Image, opts *EncodeSettings) error {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "png":
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return png.Encode(f, img)

	case "bmp":
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return bmp.Encode(f, img)

	case "tga":
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return tga.Encode(f, img)

	case "tiff":
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return tiff.Encode(f, img, &tiff.Options{Compression: tiff.Deflate})

	case "dds":
		cfg := effectiveEncodeSettings(opts)
		if err := ValidateQualityLevel(cfg.Quality); err != nil {
			return err
		}

		encOpts := &bcn.EncodeOptions{
			QualityLevel: cfg.Quality,
			Workers:      0,
		}

		dds, err := bcn.EncodeDDSWithOptions([]image.Image{img}, cfg.Format, encOpts)
		if err != nil {
			return err
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		return dds.Write(f)

	case "edds":
		cfg := effectiveEncodeSettings(opts)
		if cfg.Mipmaps < 0 {
			return fmt.Errorf("mipmaps must be >= 0")
		}
		if err := ValidateQualityLevel(cfg.Quality); err != nil {
			return err
		}

		return edds.WriteWithOptions(img, path, &edds.WriteOptions{
			Format:     cfg.Format,
			MaxMipMaps: cfg.Mipmaps,
			Compress:   true,
			EncodeOptions: &bcn.EncodeOptions{
				QualityLevel: cfg.Quality,
				Workers:      0,
			},
		})

	default:
		return fmt.Errorf("unsupported output format: %q", ext)
	}
}
