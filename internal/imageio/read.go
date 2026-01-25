package imageio

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "image/png"

	_ "github.com/schwarzlichtbezirk/tga"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/woozymasta/imageset-packer/internal/edds"
)

// Read loads an image from a supported file format.
func Read(path string) (image.Image, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "png", "bmp", "tga", "tiff":
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		img, _, err := image.Decode(f)
		if err != nil {
			return nil, err
		}
		return img, nil

	case "dds":
		return readDDS(path)

	case "edds":
		return edds.ReadEDDS(path)

	default:
		return nil, fmt.Errorf("unsupported input format: %q", ext)
	}
}

// GetImageSize reads only image dimensions without decoding full pixel data.
func GetImageSize(path string) (width, height int, err error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "png", "bmp", "tga", "tiff":
		f, err := os.Open(path)
		if err != nil {
			return 0, 0, err
		}
		defer func() { _ = f.Close() }()

		cfg, _, err := image.DecodeConfig(f)
		if err != nil {
			return 0, 0, err
		}
		return cfg.Width, cfg.Height, nil

	case "dds":
		return readDDSSize(path)

	case "edds":
		cfg, err := edds.ReadEDDSConfig(path)
		if err != nil {
			return 0, 0, err
		}
		return cfg.Width, cfg.Height, nil

	default:
		return 0, 0, fmt.Errorf("unsupported input format: %q", ext)
	}
}
