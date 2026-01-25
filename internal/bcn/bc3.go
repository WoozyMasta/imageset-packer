// Package bcn provides BC3 (DXT4/DXT5) codec.
package bcn

import "fmt"

// BC3 Block structure: 16 bytes
// - alpha_block: BC4 (8 bytes)
// - color_block: BC1 (8 bytes)

// encodeBlockBC3 encodes a 4x4 block to BC3 format.
func encodeBlockBC3(block [16]ColorRGBA) [16]byte {
	// Encode alpha with BC4
	alphaBlock := encodeBlockBC4(block)

	// Encode color with BC1 (ignore alpha in BC1 encoding)
	colorBlock := encodeBlockBC1(block)

	// Combine: alpha block (8 bytes) + color block (8 bytes)
	var result [16]byte
	copy(result[0:8], alphaBlock[:])
	copy(result[8:16], colorBlock[:])

	return result
}

// decodeBlockBC3 decodes a BC3 block (16 bytes) to 4x4 RGBA.
func decodeBlockBC3(data []byte) [16]ColorRGBA {
	if len(data) < 16 {
		panic("BC3 block must be 16 bytes")
	}

	// Decode alpha from BC4
	alphas := decodeBlockBC4(data[0:8])

	// Decode color from BC1
	colorBlock := decodeBlockBC1(data[8:16])

	// Combine alpha with colors
	for i := range colorBlock {
		colorBlock[i].A = alphas[i]
	}

	return colorBlock
}

// EncodeBC3 encodes RGBA image to BC3 format.
func EncodeBC3(rgba []byte, width, height int) ([]byte, error) {
	blocksW := (width + 3) / 4
	blocksH := (height + 3) / 4
	result := make([]byte, blocksW*blocksH*16)

	for y := 0; y < blocksH; y++ {
		for x := 0; x < blocksW; x++ {
			block := fetchBlock(rgba, x*4, y*4, width, height)
			encoded := encodeBlockBC3(block)
			offset := (y*blocksW + x) * 16
			copy(result[offset:], encoded[:])
		}
	}

	return result, nil
}

// DecodeBC3 decodes BC3 data to RGBA.
func DecodeBC3(data []byte, width, height int) ([]byte, error) {
	blocksW := (width + 3) / 4
	blocksH := (height + 3) / 4
	expectedSize := blocksW * blocksH * 16

	if len(data) < expectedSize {
		return nil, fmt.Errorf("BC3 data too short: expected %d bytes, got %d", expectedSize, len(data))
	}

	result := make([]byte, width*height*4)

	for y := 0; y < blocksH; y++ {
		for x := 0; x < blocksW; x++ {
			offset := (y*blocksW + x) * 16
			block := decodeBlockBC3(data[offset : offset+16])

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
