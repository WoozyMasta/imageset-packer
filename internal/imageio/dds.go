package imageio

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"

	"github.com/woozymasta/imageset-packer/internal/bcn"
	"github.com/woozymasta/imageset-packer/internal/dds"
)

// readDDS reads a DDS file and returns an image.Image.
func readDDS(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	h, err := dds.ReadHeader(f)
	if err != nil {
		return nil, err
	}
	dx10, err := dds.ReadHeaderDx10(f, h)
	if err != nil {
		return nil, err
	}

	format, _ := bcn.DetectFormat(h, dx10)
	data, err := readAll(f)
	if err != nil {
		return nil, err
	}

	width := int(h.Width)
	height := int(h.Height)
	expected := bcn.ExpectedDataLength(format, width, height)
	if expected < 0 || len(data) < expected {
		return nil, fmt.Errorf("dds data size mismatch: expected >= %d, got %d", expected, len(data))
	}

	rgba, err := bcn.ConvertToRGBA(data[:expected], format, width, height)
	if err != nil {
		return nil, err
	}

	return &image.RGBA{
		Pix:    rgba,
		Stride: width * 4,
		Rect:   image.Rect(0, 0, width, height),
	}, nil
}

// readDDSSize reads the size of the DDS file.
func readDDSSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = f.Close() }()

	h, err := dds.ReadHeader(f)
	if err != nil {
		return 0, 0, err
	}

	return int(h.Width), int(h.Height), nil
}

// writeDDSRGBA8 writes a DDS file with RGBA8 format.
func writeDDSRGBA8(path string, img image.Image) error {
	// Write uncompressed RGBA8 DDS (BGRA byte order in payload).
	b := img.Bounds()
	width := uint32(b.Dx())  //nolint:gosec // Dimensions come from image bounds.
	height := uint32(b.Dy()) //nolint:gosec // Dimensions come from image bounds.

	rgba := image.NewRGBA(b)
	drawToRGBA(rgba, img)

	payload := make([]byte, len(rgba.Pix))
	for i := 0; i+3 < len(payload); i += 4 {
		payload[i] = rgba.Pix[i+2]   // B
		payload[i+1] = rgba.Pix[i+1] // G
		payload[i+2] = rgba.Pix[i]   // R
		payload[i+3] = rgba.Pix[i+3] // A
	}

	header := dds.CreateHeaderRGBA8(width, height, 1)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if err := dds.WriteMagic(f); err != nil {
		return err
	}
	if err := dds.WriteHeader(f, header); err != nil {
		return err
	}
	_, err = f.Write(payload)
	return err
}

// readAll reads all data from the reader.
func readAll(r io.Reader) ([]byte, error) {
	buf := make([]byte, 0, 1<<20)
	tmp := make([]byte, 32*1024)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				return buf, nil
			}
			return nil, err
		}
	}
}

// drawToRGBA draws the source image to the destination image.
func drawToRGBA(dst *image.RGBA, src image.Image) {
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
}
