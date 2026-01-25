package imageio

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/schwarzlichtbezirk/tga"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/woozymasta/imageset-packer/internal/edds"
)

// Write saves an image to the given path based on its extension.
func Write(path string, img image.Image) error {
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
		return writeDDSRGBA8(path, img)

	case "edds":
		return edds.WriteEDDS(img, path)

	default:
		return fmt.Errorf("unsupported output format: %q", ext)
	}
}
