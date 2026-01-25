// Package dds provides functions for working with DDS (DirectDraw Surface) files.
package dds

const (
	Magic = "DDS "

	HeaderSize      = 124 // Size of DDS_HEADER structure
	PixelFormatSize = 32  // Size of DDS_PIXELFORMAT structure

	// DDS_HEADER flags
	DCaps        = 0x1
	DHeight      = 0x2
	DWidth       = 0x4
	DPitch       = 0x8
	DPixelFormat = 0x1000
	DMipMapCount = 0x20000
	DLinearSize  = 0x80000
	DDepth       = 0x800000

	// DDS_PIXELFORMAT flags
	PFAlphaPixels = 0x1
	PFAlpha       = 0x2
	PFFourCC      = 0x4
	PFRGB         = 0x40
	PFYUV         = 0x200
	PFLuminance   = 0x20000

	// DDS_CAPS flags
	CapsComplex  = 0x8
	CapsTexture  = 0x1000
	CapsMipMap   = 0x400000
	Caps2Cubemap = 0x200

	HeaderFlagsTexture    = DCaps | DHeight | DWidth | DPixelFormat
	HeaderFlagsMipMap     = DMipMapCount
	HeaderFlagsVolume     = DDepth
	HeaderFlagsPitch      = DPitch
	HeaderFlagsLinearSize = DLinearSize

	// DX10 fourCC
	FourCCDX10 = 0x30315844 // "DX10" in little-endian
)

// PixelFormat represents DDS_PIXELFORMAT structure.
type PixelFormat struct {
	Size        uint32
	Flags       uint32
	FourCC      uint32
	RGBBitCount uint32
	RBitMask    uint32
	GBitMask    uint32
	BBitMask    uint32
	ABitMask    uint32
}

// Header represents DDS_HEADER structure.
type Header struct {
	Size              uint32
	Flags             uint32
	Height            uint32
	Width             uint32
	PitchOrLinearSize uint32
	Depth             uint32
	MipMapCount       uint32
	Reserved1         [11]uint32
	PixelFormat       PixelFormat
	Caps              uint32
	Caps2             uint32
	Caps3             uint32
	Caps4             uint32
	Reserved2         uint32
}

// HeaderDx10 represents DDS_HEADER_DXT10 structure.
type HeaderDx10 struct {
	DXGIFormat        uint32
	ResourceDimension uint32
	MiscFlag          uint32
	ArraySize         uint32
	MiscFlags2        uint32
}
