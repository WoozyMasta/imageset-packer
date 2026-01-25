// Package bcn provides BC2 (DXT2/DXT3) codec.
package bcn

import "fmt"

// DecodeBC2 decodes BC2 data to RGBA (BC2 uses explicit alpha, BC1 color).
// BC2: 16 bytes - 4-bit alpha per pixel (64 bits) + BC1 color (8 bytes)
func DecodeBC2(data []byte, width, height int) ([]byte, error) {
	blocksW := (width + 3) / 4
	blocksH := (height + 3) / 4
	expectedSize := blocksW * blocksH * 16

	if len(data) < expectedSize {
		return nil, fmt.Errorf("BC2 data too short: expected %d bytes, got %d", expectedSize, len(data))
	}

	result := make([]byte, width*height*4)

	for y := 0; y < blocksH; y++ {
		for x := 0; x < blocksW; x++ {
			offset := (y*blocksW + x) * 16

			// Decode 4-bit alpha values (first 8 bytes)
			var alphas [16]uint8
			for i := 0; i < 8; i++ {
				byteVal := data[offset+i]
				alphas[i*2] = (byteVal & 0x0F) * 17 // Scale 4-bit to 8-bit
				alphas[i*2+1] = (byteVal >> 4) * 17
			}

			// Decode color from BC1 (last 8 bytes)
			colorBlock := decodeBlockBC1(data[offset+8 : offset+16])

			// Combine
			for i := range colorBlock {
				colorBlock[i].A = alphas[i]
			}

			// Write block to result
			for row := 0; row < 4; row++ {
				for col := 0; col < 4; col++ {
					px := x*4 + col
					py := y*4 + row
					if px < width && py < height {
						idx := (py*width + px) * 4
						c := colorBlock[row*4+col]
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
