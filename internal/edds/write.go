// Package edds provides EDDS writing functionality.
package edds

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"os"

	"github.com/woozymasta/imageset-packer/internal/dds"
)

// WriteEDDS writes an image as EDDS file.
func WriteEDDS(img image.Image, path string) error {
	return WriteEDDSWithMipmaps(img, path, 0)
}

// WriteEDDSWithMipmaps writes an image as EDDS file with a mipmap limit.
// maxMipMaps=0 means full chain. If maxMipMaps exceeds the possible count,
// the extra levels are ignored.
func WriteEDDSWithMipmaps(img image.Image, path string, maxMipMaps int) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert to RGBA
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Generate mipmaps
	mipMapCount := int(calculateMipMapCount(width, height))
	if maxMipMaps > 0 && maxMipMaps < mipMapCount {
		mipMapCount = maxMipMaps
	}
	if mipMapCount < 1 {
		mipMapCount = 1
	}

	// Create DDS header
	header := dds.CreateHeaderRGBA8(uint32(width), uint32(height), uint32(mipMapCount)) //nolint:gosec // Dimensions come from image bounds.

	// Generate mipmap data
	mipmaps := generateMipmaps(rgba, mipMapCount)

	// Compress each mipmap
	// mipmaps[0] is largest, mipmaps[N] is smallest.
	blocks := make([]*Block, mipMapCount)
	for i := 0; i < mipMapCount; i++ {
		mipData := mipmaps[i]

		// Convert RGBA to BGRA (DDS expectation)
		dataToCompress := make([]byte, len(mipData.data))
		for k := 0; k < len(mipData.data); k += 4 {
			dataToCompress[k] = mipData.data[k+2]   // B
			dataToCompress[k+1] = mipData.data[k+1] // G
			dataToCompress[k+2] = mipData.data[k]   // R
			dataToCompress[k+3] = mipData.data[k+3] // A
		}

		block, err := compressBlock(dataToCompress)
		if err != nil {
			return fmt.Errorf("failed to compress mipmap %d: %w", i, err)
		}
		blocks[i] = block
	}

	// Create Output File
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create EDDS file %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	// 1. Write DDS Magic
	if err := dds.WriteMagic(f); err != nil {
		return fmt.Errorf("writing DDS magic: %w", err)
	}

	// 2. Write DDS Header
	if err := dds.WriteHeader(f, header); err != nil {
		return fmt.Errorf("writing DDS header: %w", err)
	}

	// 3. Write Block Header Table (Magic + Size)
	// Written from Smallest Mipmap (index Count-1) to Largest Mipmap (index 0)
	for i := mipMapCount - 1; i >= 0; i-- {
		block := blocks[i]

		// Format: [Magic (4 bytes)] [Size (4 bytes)]
		if _, err := f.Write([]byte(block.Magic)); err != nil {
			return fmt.Errorf("writing block magic for mipmap %d: %w", i, err)
		}

		if err := binary.Write(f, binary.LittleEndian, block.Size); err != nil {
			return fmt.Errorf("writing block size for mipmap %d: %w", i, err)
		}
	}

	// 4. Write Block Data Body
	// Written from Smallest to Largest
	for i := mipMapCount - 1; i >= 0; i-- {
		if err := writeBlockData(f, blocks[i]); err != nil {
			return fmt.Errorf("writing block data for mipmap %d: %w", i, err)
		}
	}

	return nil
}

// calculateMipMapCount calculates the number of mipmaps needed.
func calculateMipMapCount(width, height int) uint32 {
	count := uint32(1)
	w, h := uint32(width), uint32(height) //nolint:gosec // Dimensions come from image bounds.
	for w > 1 || h > 1 {
		count++
		if w > 1 {
			w /= 2
		}
		if h > 1 {
			h /= 2
		}
	}
	if count > 11 {
		count = 11
	}
	return count
}

type mipmapData struct {
	data   []byte
	width  int
	height int
}

// generateMipmaps generates mipmap chain from RGBA image.
func generateMipmaps(rgba *image.RGBA, maxMipMaps int) []mipmapData {
	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()

	mipCount := 1
	w, h := width, height
	for mipCount < maxMipMaps && (w > 1 || h > 1) {
		mipCount++
		if w > 1 {
			w /= 2
		}
		if h > 1 {
			h /= 2
		}
	}

	result := make([]mipmapData, mipCount)

	// Level 0
	firstMipData := make([]byte, len(rgba.Pix))
	copy(firstMipData, rgba.Pix)
	result[0] = mipmapData{
		data:   firstMipData,
		width:  width,
		height: height,
	}

	// Subsequent levels
	current := rgba.Pix
	currentWidth := width
	currentHeight := height

	for i := 1; i < mipCount; i++ {
		newWidth := max(1, currentWidth>>1)
		newHeight := max(1, currentHeight>>1)

		newData := resizeToHalf(current, currentWidth, currentHeight, newWidth, newHeight)

		result[i] = mipmapData{
			data:   newData,
			width:  newWidth,
			height: newHeight,
		}

		current = newData
		currentWidth = newWidth
		currentHeight = newHeight
	}

	return result
}

func resizeToHalf(src []byte, srcWidth, srcHeight, dstWidth, dstHeight int) []byte {
	dst := make([]byte, dstWidth*dstHeight*4)

	// Helper to clamp coordinates
	getIdx := func(x, y int) int {
		if x >= srcWidth {
			x = srcWidth - 1
		}
		if y >= srcHeight {
			y = srcHeight - 1
		}
		return (y*srcWidth + x) * 4
	}

	for y := 0; y < dstHeight; y++ {
		for x := 0; x < dstWidth; x++ {
			srcX := x * 2
			srcY := y * 2

			// Box filter (average 4 pixels)
			idx0 := getIdx(srcX, srcY)
			idx1 := getIdx(srcX+1, srcY)
			idx2 := getIdx(srcX, srcY+1)
			idx3 := getIdx(srcX+1, srcY+1)

			dstIdx := (y*dstWidth + x) * 4

			for c := 0; c < 4; c++ {
				sum := int(src[idx0+c]) + int(src[idx1+c]) + int(src[idx2+c]) + int(src[idx3+c])
				dst[dstIdx+c] = byte(sum / 4)
			}
		}
	}
	return dst
}
