// Package edds provides EDDS reading functionality.
package edds

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

	"github.com/pierrec/lz4/v4"

	"github.com/woozymasta/imageset-packer/internal/bcn"
	"github.com/woozymasta/imageset-packer/internal/dds"
)

// ReadEDDSConfig reads EDDS file configuration without decoding the image data.
func ReadEDDSConfig(path string) (image.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return image.Config{}, fmt.Errorf("failed to open EDDS file %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	header, _, err := readEDDSHeaders(f)
	if err != nil {
		return image.Config{}, err
	}

	return image.Config{
		Width:      int(header.Width),
		Height:     int(header.Height),
		ColorModel: color.RGBAModel,
	}, nil
}

// ReadEDDS reads and decodes an EDDS file.
func ReadEDDS(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open EDDS file %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	header, dx10, err := readEDDSHeaders(f)
	if err != nil {
		return nil, err
	}

	// Detect format
	format, _ := bcn.DetectFormat(header, dx10)

	// Determine mipmap count
	mipMapCount := uint32(1)
	if (header.Caps&dds.CapsMipMap) != 0 && header.MipMapCount > 0 {
		mipMapCount = header.MipMapCount
	}

	// Read blocks for each mipmap (we'll use the largest one)
	var mipData []byte
	var mipWidth, mipHeight int

	mipData, mipWidth, mipHeight, err = readLargestMipFromBlocks(f, header, format, mipMapCount)
	if err != nil {
		mipData, mipWidth, mipHeight, err = readLegacySingleBlock(f, header, dx10, format)
		if err != nil {
			return nil, err
		}
	}

	// Convert to RGBA
	rgbaData, err := bcn.ConvertToRGBA(mipData, format, mipWidth, mipHeight)
	if err != nil {
		return nil, fmt.Errorf("converting to RGBA: %w", err)
	}

	// DDS payload is typically bottom-up, but most tools expect top-down.
	img := &image.NRGBA{
		Pix:    rgbaData,
		Stride: mipWidth * 4,
		Rect:   image.Rect(0, 0, mipWidth, mipHeight),
	}

	return img, nil
}

func readLargestMipFromBlocks(r io.ReadSeeker, header *dds.Header, format bcn.Format, mipMapCount uint32) ([]byte, int, int, error) {
	if mipMapCount == 0 {
		mipMapCount = 1
	}

	table, err := readBlockTable(r, mipMapCount)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("reading block table: %w", err)
	}

	for i := uint32(0); i < mipMapCount; i++ {
		mipLevel := mipMapCount - i - 1
		if mipLevel != 0 {
			if _, err := r.Seek(int64(table[i].Size), io.SeekCurrent); err != nil {
				return nil, 0, 0, fmt.Errorf("skipping block body for mipmap %d: %w", i, err)
			}
			continue
		}

		block, err := readBlockBody(r, table[i])
		if err != nil {
			return nil, 0, 0, fmt.Errorf("reading block body for mipmap %d: %w", i, err)
		}

		mipW := mipDimension(int(header.Width), int(mipLevel))
		mipH := mipDimension(int(header.Height), int(mipLevel))

		expectedSize := bcn.ExpectedDataLength(format, mipW, mipH)
		if expectedSize <= 0 {
			return nil, 0, 0, fmt.Errorf("unknown/invalid format %s for mipmap %d", format, i)
		}

		decompressed, err := decompressBlock(block, expectedSize)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("decompressing block for mipmap %d: %w", i, err)
		}
		if len(decompressed) != expectedSize {
			return nil, 0, 0, fmt.Errorf("largest mip size mismatch: expected %d, got %d", expectedSize, len(decompressed))
		}

		return decompressed, mipW, mipH, nil
	}

	return nil, 0, 0, fmt.Errorf("failed to pick largest mip: mipmaps=%d", mipMapCount)
}

func readLegacySingleBlock(r io.ReadSeeker, header *dds.Header, dx10 *dds.HeaderDx10, format bcn.Format) ([]byte, int, int, error) {
	headerSize := int64(4 + dds.HeaderSize)
	if dx10 != nil {
		headerSize += 20
	}
	if _, err := r.Seek(headerSize, io.SeekStart); err != nil {
		return nil, 0, 0, fmt.Errorf("seeking to data start: %w", err)
	}

	remainingData, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("reading remaining data: %w", err)
	}

	expectedSize := bcn.ExpectedDataLength(format, int(header.Width), int(header.Height))
	if expectedSize <= 0 {
		return nil, 0, 0, fmt.Errorf("unknown format %s", format)
	}

	decompressed := make([]byte, expectedSize)
	n, lz4Err := lz4.UncompressBlock(remainingData, decompressed)
	if lz4Err == nil && n == expectedSize {
		return decompressed, int(header.Width), int(header.Height), nil
	}
	if len(remainingData) == expectedSize {
		return remainingData, int(header.Width), int(header.Height), nil
	}
	return nil, 0, 0, fmt.Errorf("failed to parse as single block: LZ4 error=%v, size mismatch (expected %d, got %d)", lz4Err, expectedSize, len(remainingData))
}

// readEDDSHeaders reads DDS magic, header, and optional DX10 header.
func readEDDSHeaders(r io.Reader) (*dds.Header, *dds.HeaderDx10, error) {
	header, err := dds.ReadHeader(r)
	if err != nil {
		return nil, nil, err
	}

	dx10, err := dds.ReadHeaderDx10(r, header)
	if err != nil {
		return nil, nil, err
	}

	return header, dx10, nil
}

// mipDimension calculates mipmap dimension.
func mipDimension(base, level int) int {
	result := base >> level
	if result < 1 {
		return 1
	}
	return result
}
