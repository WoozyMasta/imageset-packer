// Package bcn provides BC1 (DXT1) codec.
package bcn

import "fmt"

// BC1 Block structure: 8 bytes
// - max: u16 (RGB565)
// - min: u16 (RGB565)
// - color_table: u32 (16 indices, 2 bits each)

// encodeBlockBC1 encodes a 4x4 block to BC1 format.
func encodeBlockBC1(block [16]ColorRGBA) [8]byte {
	minColor, maxColor := minMaxLuminance(block)

	min565 := minColor.to565()
	max565 := maxColor.to565()

	// Determine if we need alpha (for BC1, if min > max, we use alpha mode)
	canContainAlpha := max565 <= min565
	if canContainAlpha {
		// Swap if needed
		min565, max565 = max565, min565
		minColor, maxColor = maxColor, minColor
	}

	// Generate reference colors
	var color2, color3 ColorRGBA
	if canContainAlpha {
		color2 = maxColor.mix11Over2Saturate(minColor)
		color3 = ColorRGBA{} // Black/transparent
	} else {
		color2 = maxColor.mix21Over3Saturate(minColor)
		color3 = maxColor.mix12Over3Saturate(minColor)
	}

	refColors := [4]ColorRGBA{maxColor, minColor, color2, color3}
	colorTable := encodeColorTableBC1BC3(block, refColors, canContainAlpha)

	// Pack into 8 bytes
	// BC1 spec: bytes 0-1 = color_0 (max), bytes 2-3 = color_1 (min)
	var result [8]byte
	result[0] = byte(max565) // color_0 (little-endian)
	result[1] = byte(max565 >> 8)
	result[2] = byte(min565) // color_1 (little-endian)
	result[3] = byte(min565 >> 8)
	result[4] = byte(colorTable)
	result[5] = byte(colorTable >> 8)
	result[6] = byte(colorTable >> 16)
	result[7] = byte(colorTable >> 24)

	return result
}

// decodeBlockBC1 decodes a BC1 block (8 bytes) to 4x4 RGBA.
//
//nolint:gosec // Fixed-size BC1 decoding indexes are safe.
func decodeBlockBC1(data []byte) [16]ColorRGBA {
	if len(data) < 8 {
		panic("BC1 block must be 8 bytes")
	}

	// BC1 spec: bytes 0-1 = color_0, bytes 2-3 = color_1
	color0_565 := uint16(data[0]) | (uint16(data[1]) << 8)
	color1_565 := uint16(data[2]) | (uint16(data[3]) << 8)
	colorTable := uint32(data[4]) | (uint32(data[5]) << 8) | (uint32(data[6]) << 16) | (uint32(data[7]) << 24)

	color0 := from565(color0_565)
	color1 := from565(color1_565)

	// Check if alpha mode (color_0 <= color_1 means alpha mode in BC1)
	hasAlpha := color0_565 <= color1_565
	var maxColor, minColor ColorRGBA
	if hasAlpha {
		// In alpha mode, swap: color_0 becomes min, color_1 becomes max
		minColor = color0
		maxColor = color1
	} else {
		// Normal mode: color_0 is max, color_1 is min
		maxColor = color0
		minColor = color1
	}

	// Generate reference colors
	var color2, color3 ColorRGBA
	if hasAlpha {
		color2 = maxColor.mix11Over2Saturate(minColor)
		color3 = ColorRGBA{} // Black/transparent
	} else {
		color2 = maxColor.mix21Over3Saturate(minColor)
		color3 = maxColor.mix12Over3Saturate(minColor)
	}

	refColors := [4]ColorRGBA{maxColor, minColor, color2, color3}

	// Decode indices and create block
	var block [16]ColorRGBA
	for i := 0; i < 16; i++ {
		idx := int((colorTable >> (i * 2)) & 0x3)
		block[i] = refColors[idx]
		if hasAlpha && idx == 3 {
			block[i].A = 0 // Transparent
		}
	}

	return block
}

// EncodeBC1 encodes RGBA image to BC1 format.
func EncodeBC1(rgba []byte, width, height int) ([]byte, error) {
	blocksW := (width + 3) / 4
	blocksH := (height + 3) / 4
	result := make([]byte, blocksW*blocksH*8)

	for y := 0; y < blocksH; y++ {
		for x := 0; x < blocksW; x++ {
			block := fetchBlock(rgba, x*4, y*4, width, height)
			encoded := encodeBlockBC1(block)
			offset := (y*blocksW + x) * 8
			copy(result[offset:], encoded[:])
		}
	}

	return result, nil
}

// DecodeBC1 decodes BC1 data to RGBA.
func DecodeBC1(data []byte, width, height int) ([]byte, error) {
	blocksW := (width + 3) / 4
	blocksH := (height + 3) / 4
	expectedSize := blocksW * blocksH * 8

	if len(data) < expectedSize {
		return nil, fmt.Errorf("BC1 data too short: expected %d bytes, got %d", expectedSize, len(data))
	}

	result := make([]byte, width*height*4)

	for y := 0; y < blocksH; y++ {
		for x := 0; x < blocksW; x++ {
			offset := (y*blocksW + x) * 8
			block := decodeBlockBC1(data[offset : offset+8])

			// Write block to result
			for row := 0; row < 4; row++ {
				for col := 0; col < 4; col++ {
					px := x*4 + col
					py := y*4 + row
					if px < width && py < height {
						idx := (py*width + px) * 4
						c := block[row*4+col]
						result[idx] = c.R
						result[idx+1] = c.G
						result[idx+2] = c.B
						result[idx+3] = c.A
					}
				}
			}
		}
	}

	return result, nil
}
