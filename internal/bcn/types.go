// Package bcn provides Block Compression (BCn/DXT) codecs.
package bcn

// Format represents the pixel format.
type Format string

const (
	FormatBC1     Format = "BC1"
	FormatBC2     Format = "BC2"
	FormatBC3     Format = "BC3"
	FormatBC4     Format = "BC4"
	FormatBC5     Format = "BC5"
	FormatBC6     Format = "BC6"
	FormatBC7     Format = "BC7"
	FormatRGBA8   Format = "RGBA8"
	FormatBGRA8   Format = "BGRA8"
	FormatUnknown Format = "UNKNOWN"
)

// ColorRGBA represents an RGBA color.
type ColorRGBA struct {
	R, G, B, A uint8
}
