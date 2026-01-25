// Package dds provides DDS format creation functions.
package dds

// CreateHeaderRGBA8 creates a DDS header for RGBA8 format.
func CreateHeaderRGBA8(width, height uint32, mipMapCount uint32) *Header {
	flags := uint32(HeaderFlagsTexture)
	// Always set mipmap flag if mipMapCount > 0.
	if mipMapCount > 0 {
		flags |= HeaderFlagsMipMap
	}
	flags |= HeaderFlagsPitch

	// Base flag for texture: DDSCAPS_TEXTURE (0x1000).
	caps := uint32(CapsTexture)

	// If mipmaps are present, add COMPLEX and MIPMAP.
	if mipMapCount > 0 {
		caps |= CapsComplex | CapsMipMap
	}

	// Match Workbench format exactly from reference file.
	// Reserved1[1] = 0x31464e45 ("ENF1" in little-endian).
	// RGB masks: R=0x00ff0000, G=0x0000ff00, B=0x000000ff, A=0xff000000 (standard RGBA).
	// Reserved2 = 0x00000000.
	reserved1 := [11]uint32{
		0,
		0x31464e45, // "ENF1" at index 1
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
	}

	return &Header{
		Size:              HeaderSize,
		Flags:             flags,
		Height:            height,
		Width:             width,
		PitchOrLinearSize: width * 4, // Pitch for RGBA8 (bytes per row).
		Depth:             0,
		MipMapCount:       mipMapCount,
		Reserved1:         reserved1,
		PixelFormat: PixelFormat{
			Size:        PixelFormatSize,
			Flags:       PFAlphaPixels | PFRGB,
			FourCC:      0x00000000, // Standard RGBA, not 0x20.
			RGBBitCount: 32,
			RBitMask:    0x00ff0000, // Standard RGBA masks.
			GBitMask:    0x0000ff00,
			BBitMask:    0x000000ff,
			ABitMask:    0xff000000, // Standard RGBA8 alpha mask.
		},
		Caps:      caps,
		Caps2:     0,
		Caps3:     0,
		Caps4:     0,
		Reserved2: 0x00000000, // Not "COPY".
	}
}
