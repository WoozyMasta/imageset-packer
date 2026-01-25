// Package bcn provides Block Compression (BCn/DXT) codecs.
package bcn

// to565 converts RGB color to RGB565 format.
func (c ColorRGBA) to565() uint16 {
	return (uint16(c.R&0b11111000) << 8) | (uint16(c.G&0b11111100) << 3) | uint16(c.B>>3)
}

// from565 converts RGB565 to ColorRGBA.
func from565(v uint16) ColorRGBA {
	r := uint8((v >> 8) & 0b11111000) //nolint:gosec // Masked to 8 bits.
	g := uint8((v >> 3) & 0b11111100) //nolint:gosec // Masked to 8 bits.
	b := uint8((v << 3) & 0b11111000) //nolint:gosec // Masked to 8 bits.
	return ColorRGBA{R: r, G: g, B: b, A: 255}
}

// luminance calculates luminance of the color.
func (c ColorRGBA) luminance() int32 {
	return int32(c.R) + int32(c.G)*2 + int32(c.B)
}

// sqrDistance calculates squared distance between two colors.
func (c ColorRGBA) sqrDistance(other ColorRGBA) int32 {
	dr := int32(c.R) - int32(other.R)
	dg := int32(c.G) - int32(other.G)
	db := int32(c.B) - int32(other.B)
	return dr*dr + dg*dg + db*db
}

// mix21Over3Saturate mixes two colors: (2*self + other) / 3.
func (c ColorRGBA) mix21Over3Saturate(other ColorRGBA) ColorRGBA {
	return ColorRGBA{
		R: mix21Over3(c.R, other.R),
		G: mix21Over3(c.G, other.G),
		B: mix21Over3(c.B, other.B),
		A: 255,
	}
}

// mix12Over3Saturate mixes two colors: (self + 2*other) / 3.
func (c ColorRGBA) mix12Over3Saturate(other ColorRGBA) ColorRGBA {
	return ColorRGBA{
		R: mix12Over3(c.R, other.R),
		G: mix12Over3(c.G, other.G),
		B: mix12Over3(c.B, other.B),
		A: 255,
	}
}

// mix11Over2Saturate mixes two colors: (self + other) / 2.
func (c ColorRGBA) mix11Over2Saturate(other ColorRGBA) ColorRGBA {
	return ColorRGBA{
		R: mix11Over2(c.R, other.R),
		G: mix11Over2(c.G, other.G),
		B: mix11Over2(c.B, other.B),
		A: 255,
	}
}

// Mix functions
func mix21Over3(x, y uint8) uint8 {
	return uint8((2*uint16(x) + uint16(y)) / 3) //nolint:gosec // Result is within 0..255.
}

func mix12Over3(x, y uint8) uint8 {
	return uint8((uint16(x) + 2*uint16(y)) / 3) //nolint:gosec // Result is within 0..255.
}

func mix11Over2(x, y uint8) uint8 {
	return uint8((uint16(x) + uint16(y)) / 2) //nolint:gosec // Result is within 0..255.
}

// minMaxLuminance finds min and max colors by luminance in a 4x4 block.
func minMaxLuminance(block [16]ColorRGBA) (ColorRGBA, ColorRGBA) {
	maxLum := int32(-1)
	minLum := int32(0x7FFFFFFF)
	maxColor := block[0]
	minColor := block[0]

	for _, p := range block {
		lum := p.luminance()
		if lum > maxLum {
			maxLum = lum
			maxColor = p
		}
		if lum < minLum {
			minLum = lum
			minColor = p
		}
	}

	return minColor, maxColor
}

// fetchBlock extracts a 4x4 block from RGBA data.
func fetchBlock(rgba []byte, x, y, width, height int) [16]ColorRGBA {
	var block [16]ColorRGBA
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			px := x + col
			py := y + row
			idx := (py*width + px) * 4

			if px < width && py < height && idx+3 < len(rgba) {
				block[row*4+col] = ColorRGBA{
					R: rgba[idx],
					G: rgba[idx+1],
					B: rgba[idx+2],
					A: rgba[idx+3],
				}
			} else {
				block[row*4+col] = ColorRGBA{} // Default (black, transparent)
			}
		}
	}
	return block
}

// encodeColorTableBC1BC3 encodes color indices for BC1/BC3.
//
//nolint:gosec // Fixed-size tables with bounded indices.
func encodeColorTableBC1BC3(block [16]ColorRGBA, refColors [4]ColorRGBA, hasAlpha bool) uint32 {
	var colorIndices [16]uint8

	for i := range colorIndices {
		p := block[i]
		if hasAlpha && p.A < 128 {
			// Map transparent pixels to index 3 (black)
			colorIndices[i] = 3
		} else {
			// Find closest color
			minDistance := int32(0x7FFFFFFF)
			bestIdx := uint8(0)
			for j, refColor := range refColors {
				distance := p.sqrDistance(refColor)
				if distance < minDistance {
					minDistance = distance
					bestIdx = uint8(j) //nolint:gosec // j is 0..3.
				}
			}
			colorIndices[i] = bestIdx
		}
	}

	// Pack indices into 32-bit value (2 bits per index)
	var colorTable uint32
	for i, idx := range colorIndices {
		colorTable |= uint32(idx) << (i * 2)
	}

	return colorTable
}

func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
