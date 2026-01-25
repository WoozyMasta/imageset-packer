// Package edds provides conversion functions for EDDS files.
package edds

import (
	"fmt"
	"image/png"
	"os"

	"github.com/woozymasta/imageset-packer/internal/bcn"
	"github.com/woozymasta/imageset-packer/internal/dds"
)

// ToDDS converts EDDS file to DDS file (uncompressed).
func ToDDS(eddsPath, ddsPath string) error {
	f, err := os.Open(eddsPath)
	if err != nil {
		return fmt.Errorf("failed to open EDDS file %q: %w", eddsPath, err)
	}
	defer func() { _ = f.Close() }()

	header, dx10, err := readEDDSHeaders(f)
	if err != nil {
		return fmt.Errorf("reading EDDS headers: %w", err)
	}

	// Determine mipmap count
	mipMapCount := uint32(1)
	if (header.Caps&dds.CapsMipMap) != 0 && header.MipMapCount > 0 {
		mipMapCount = header.MipMapCount
	}

	// Detect format
	format, _ := bcn.DetectFormat(header, dx10)

	// Read and decompress all blocks
	table, err := readBlockTable(f, mipMapCount)
	if err != nil {
		return fmt.Errorf("reading block table: %w", err)
	}

	var allData []byte
	for i := uint32(0); i < mipMapCount; i++ {
		block, err := readBlockBody(f, table[i])
		if err != nil {
			return fmt.Errorf("reading block body for mipmap %d: %w", i, err)
		}

		mipLevel := mipMapCount - i - 1
		mipW := mipDimension(int(header.Width), int(mipLevel))
		mipH := mipDimension(int(header.Height), int(mipLevel))

		expectedSize := bcn.ExpectedDataLength(format, mipW, mipH)
		if expectedSize < 0 {
			return fmt.Errorf("unknown format %s", format)
		}

		decompressed, err := decompressBlock(block, expectedSize)
		if err != nil {
			return fmt.Errorf("decompressing block for mipmap %d: %w", i, err)
		}

		allData = append(allData, decompressed...)
	}

	// Write DDS file
	out, err := os.Create(ddsPath)
	if err != nil {
		return fmt.Errorf("failed to create DDS file %q: %w", ddsPath, err)
	}
	defer func() { _ = out.Close() }()

	// Write DDS magic
	if err := dds.WriteMagic(out); err != nil {
		return fmt.Errorf("writing DDS magic: %w", err)
	}

	// Write DDS header
	if err := dds.WriteHeader(out, header); err != nil {
		return fmt.Errorf("writing DDS header: %w", err)
	}

	// Write DX10 header if present
	if dx10 != nil {
		if err := dds.WriteHeaderDx10(out, dx10); err != nil {
			return fmt.Errorf("writing DX10 header: %w", err)
		}
	}

	// Write decompressed data
	if _, err := out.Write(allData); err != nil {
		return fmt.Errorf("writing DDS data: %w", err)
	}

	return nil
}

// ToPNG converts EDDS file to PNG file.
func ToPNG(eddsPath, pngPath string) error {
	img, err := ReadEDDS(eddsPath)
	if err != nil {
		return fmt.Errorf("reading EDDS file: %w", err)
	}

	f, err := os.Create(pngPath)
	if err != nil {
		return fmt.Errorf("failed to create PNG file %q: %w", pngPath, err)
	}
	defer func() { _ = f.Close() }()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}

	return nil
}
