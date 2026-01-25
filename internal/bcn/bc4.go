// Package bcn provides BC4 codec.
package bcn

// BC4 Block structure: 8 bytes
// - max_alpha: u8
// - min_alpha: u8
// - alpha_table: [6]u8 (packed indices)

// genAlphaRef generates 8 alpha reference values from alpha_0 and alpha_1.
// According to BC4 spec and BCnEncoder.NET: if alpha_0 > alpha_1, interpolate 6 values; otherwise 4 values + 0 and 255.
func genAlphaRef(alpha0, alpha1 uint8) [8]uint8 {
	if alpha0 > alpha1 {
		// 6 interpolated alpha values (BC4 spec) using InterpolateSeventh with correction=3
		return [8]uint8{
			alpha0,                                // bit code 000
			alpha1,                                // bit code 001
			interpolateSeventh(alpha0, alpha1, 1), // bit code 010: ((7-1)*a0 + 1*a1 + 3)/7
			interpolateSeventh(alpha0, alpha1, 2), // bit code 011: ((7-2)*a0 + 2*a1 + 3)/7
			interpolateSeventh(alpha0, alpha1, 3), // bit code 100: ((7-3)*a0 + 3*a1 + 3)/7
			interpolateSeventh(alpha0, alpha1, 4), // bit code 101: ((7-4)*a0 + 4*a1 + 3)/7
			interpolateSeventh(alpha0, alpha1, 5), // bit code 110: ((7-5)*a0 + 5*a1 + 3)/7
			interpolateSeventh(alpha0, alpha1, 6), // bit code 111: ((7-6)*a0 + 6*a1 + 3)/7
		}
	} else {
		// 4 interpolated alpha values + 0 and 255 (BC4 spec) using InterpolateFifth with correction=2
		return [8]uint8{
			alpha0,                              // bit code 000
			alpha1,                              // bit code 001
			interpolateFifth(alpha0, alpha1, 1), // bit code 010: ((5-1)*a0 + 1*a1 + 2)/5
			interpolateFifth(alpha0, alpha1, 2), // bit code 011: ((5-2)*a0 + 2*a1 + 2)/5
			interpolateFifth(alpha0, alpha1, 3), // bit code 100: ((5-3)*a0 + 3*a1 + 2)/5
			interpolateFifth(alpha0, alpha1, 4), // bit code 101: ((5-4)*a0 + 4*a1 + 2)/5
			0,                                   // bit code 110: 0.0f
			255,                                 // bit code 111: 1.0f (255)
		}
	}
}

// interpolateSeventh interpolates two values by seventh with correction (BCnEncoder.NET formula).
// Formula: ((7 - num) * a0 + num * a1 + 3) / 7
func interpolateSeventh(a0, a1 uint8, num int) uint8 {
	return uint8(((7-num)*int(a0) + num*int(a1) + 3) / 7) //nolint:gosec // Result is within 0..255.
}

// interpolateFifth interpolates two values by fifth with correction (BCnEncoder.NET formula).
// Formula: ((5 - num) * a0 + num * a1 + 2) / 5
func interpolateFifth(a0, a1 uint8, num int) uint8 {
	return uint8(((5-num)*int(a0) + num*int(a1) + 2) / 5) //nolint:gosec // Result is within 0..255.
}

// minMaxAlpha finds min and max alpha values in a block.
func minMaxAlpha(block [16]ColorRGBA) (minAlpha, maxAlpha uint8) {
	minAlpha = 255
	maxAlpha = 0
	for _, p := range block {
		if p.A < minAlpha {
			minAlpha = p.A
		}
		if p.A > maxAlpha {
			maxAlpha = p.A
		}
	}
	return minAlpha, maxAlpha
}

// encodeBlockBC4 encodes a 4x4 block's alpha channel to BC4 format.
func encodeBlockBC4(block [16]ColorRGBA) [8]byte {
	minAlpha, maxAlpha := minMaxAlpha(block)
	// BC4 spec: alpha_0 is max, alpha_1 is min (stored in this order)
	alpha0 := maxAlpha
	alpha1 := minAlpha
	alphaRef := genAlphaRef(alpha0, alpha1)

	// Find closest alpha for each pixel
	var indices [16]uint8
	for i, p := range block {
		minDelta := int32(0x7FFFFFFF)
		alpha := int32(p.A)
		for j, refAlpha := range alphaRef {
			delta := abs(int32(refAlpha) - alpha)
			if delta < minDelta {
				minDelta = delta
				indices[i] = uint8(j) //nolint:gosec // j is 0..7.
			}
		}
	}

	// Pack indices into 6 bytes (3 bits per index)
	alphaTable := [6]uint8{
		(indices[0] << 0) | (indices[1] << 3) | (indices[2] << 6),
		(indices[2] >> 2) | (indices[3] << 1) | (indices[4] << 4) | (indices[5] << 7),
		(indices[5] >> 1) | (indices[6] << 2) | (indices[7] << 5),
		(indices[8] << 0) | (indices[9] << 3) | (indices[10] << 6),
		(indices[10] >> 2) | (indices[11] << 1) | (indices[12] << 4) | (indices[13] << 7),
		(indices[13] >> 1) | (indices[14] << 2) | (indices[15] << 5),
	}

	return [8]byte{
		alpha0, // alpha_0 (max)
		alpha1, // alpha_1 (min)
		alphaTable[0],
		alphaTable[1],
		alphaTable[2],
		alphaTable[3],
		alphaTable[4],
		alphaTable[5],
	}
}

// decodeBlockBC4 decodes a BC4 block (8 bytes) to 4x4 alpha values.
//
//nolint:gosec // Fixed-size BC4 decoding indexes are safe.
func decodeBlockBC4(data []byte) [16]uint8 {
	if len(data) < 8 {
		panic("BC4 block must be 8 bytes")
	}

	alpha0 := data[0] // BC4 spec: alpha_0
	alpha1 := data[1] // BC4 spec: alpha_1
	alphaTable := [6]uint8{data[2], data[3], data[4], data[5], data[6], data[7]}

	alphaRef := genAlphaRef(alpha0, alpha1)

	// Unpack indices from 6 bytes
	var indices [16]uint8
	indices[0] = (alphaTable[0] >> 0) & 0x7
	indices[1] = (alphaTable[0] >> 3) & 0x7
	indices[2] = ((alphaTable[0] >> 6) & 0x3) | ((alphaTable[1] << 2) & 0x4)
	indices[3] = (alphaTable[1] >> 1) & 0x7
	indices[4] = (alphaTable[1] >> 4) & 0x7
	indices[5] = ((alphaTable[1] >> 7) & 0x1) | ((alphaTable[2] << 1) & 0x6)
	indices[6] = (alphaTable[2] >> 2) & 0x7
	indices[7] = (alphaTable[2] >> 5) & 0x7
	indices[8] = (alphaTable[3] >> 0) & 0x7
	indices[9] = (alphaTable[3] >> 3) & 0x7
	indices[10] = ((alphaTable[3] >> 6) & 0x3) | ((alphaTable[4] << 2) & 0x4)
	indices[11] = (alphaTable[4] >> 1) & 0x7
	indices[12] = (alphaTable[4] >> 4) & 0x7
	indices[13] = ((alphaTable[4] >> 7) & 0x1) | ((alphaTable[5] << 1) & 0x6)
	indices[14] = (alphaTable[5] >> 2) & 0x7
	indices[15] = (alphaTable[5] >> 5) & 0x7

	// Decode alpha values
	var alphas [16]uint8
	for i, idx := range indices {
		alphas[i] = alphaRef[idx]
	}

	return alphas
}
