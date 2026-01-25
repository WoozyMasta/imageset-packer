// Package bcn provides format detection and conversion functions.
package bcn

import (
	"fmt"

	"github.com/woozymasta/imageset-packer/internal/dds"
)

// DetectFormat detects the format from DDS header.
func DetectFormat(header *dds.Header, dx10 *dds.HeaderDx10) (Format, string) {
	if dx10 != nil {
		format := mapDxgiFormat(dx10.DXGIFormat)
		return format, fmt.Sprintf("DXGI %d", dx10.DXGIFormat)
	}

	pf := header.PixelFormat
	if (pf.Flags & dds.PFFourCC) != 0 {
		fourCCStr := intToFourCC(pf.FourCC)
		switch fourCCStr {
		case "DXT1":
			return FormatBC1, fourCCStr
		case "DXT2", "DXT3":
			return FormatBC2, fourCCStr
		case "DXT4", "DXT5":
			return FormatBC3, fourCCStr
		case "ATI1", "BC4U", "BC4S":
			return FormatBC4, fourCCStr
		case "ATI2", "BC5U", "BC5S":
			return FormatBC5, fourCCStr
		default:
			return FormatUnknown, fourCCStr
		}
	}

	if (pf.Flags & dds.PFRGB) != 0 {
		if (pf.Flags&dds.PFAlphaPixels != 0) && pf.RGBBitCount == 32 {
			if pf.RBitMask == 0x000000ff && pf.GBitMask == 0x0000ff00 &&
				pf.BBitMask == 0x00ff0000 && pf.ABitMask == 0xff000000 {
				return FormatRGBA8, "RGBA8"
			}
			if pf.RBitMask == 0x00ff0000 && pf.GBitMask == 0x0000ff00 &&
				pf.BBitMask == 0x000000ff && pf.ABitMask == 0xff000000 {
				return FormatBGRA8, "BGRA8"
			}
		}
	}

	if (pf.Flags&dds.PFLuminance) != 0 && pf.RGBBitCount == 8 {
		return FormatRGBA8, "LUMINANCE8"
	}

	return FormatUnknown, "UNKNOWN"
}

// mapDxgiFormat maps DXGI format to Format.
func mapDxgiFormat(dxgiFormat uint32) Format {
	switch dxgiFormat {
	case 71: // DXGI_FORMAT_BC1_UNORM
		return FormatBC1
	case 74: // DXGI_FORMAT_BC2_UNORM
		return FormatBC2
	case 77: // DXGI_FORMAT_BC3_UNORM
		return FormatBC3
	case 80: // DXGI_FORMAT_BC4_UNORM
		return FormatBC4
	case 83: // DXGI_FORMAT_BC5_UNORM
		return FormatBC5
	case 95: // DXGI_FORMAT_BC6H_UF16
		return FormatBC6
	case 98: // DXGI_FORMAT_BC7_UNORM
		return FormatBC7
	case 87: // DXGI_FORMAT_B8G8R8A8_UNORM
		return FormatBGRA8
	case 28: // DXGI_FORMAT_R8G8B8A8_UNORM
		return FormatRGBA8
	default:
		return FormatUnknown
	}
}

// intToFourCC converts uint32 to fourCC string.
func intToFourCC(value uint32) string {
	return string([]byte{
		byte(value & 0xff),
		byte((value >> 8) & 0xff),
		byte((value >> 16) & 0xff),
		byte((value >> 24) & 0xff),
	})
}

// ExpectedDataLength calculates expected data length for a format.
func ExpectedDataLength(format Format, width, height int) int {
	blocksW := (width + 3) / 4
	blocksH := (height + 3) / 4

	switch format {
	case FormatBC1, FormatBC4:
		return blocksW * blocksH * 8
	case FormatBC2, FormatBC3, FormatBC5, FormatBC6, FormatBC7:
		return blocksW * blocksH * 16
	case FormatRGBA8, FormatBGRA8:
		return width * height * 4
	default:
		return -1 // Unknown
	}
}

// ConvertToRGBA converts data to RGBA format.
func ConvertToRGBA(data []byte, format Format, width, height int) ([]byte, error) {
	switch format {
	case FormatBC1:
		return DecodeBC1(data, width, height)
	case FormatBC2:
		return DecodeBC2(data, width, height)
	case FormatBC3:
		return DecodeBC3(data, width, height)
	case FormatBC6:
		return nil, fmt.Errorf("BC6 (HDR) conversion is not yet implemented")
	case FormatBC7:
		return nil, fmt.Errorf("BC7 conversion is not yet implemented")
	case FormatRGBA8:
		// Already RGBA, just copy
		result := make([]byte, len(data))
		copy(result, data)
		return result, nil
	case FormatBGRA8:
		// Convert BGRA to RGBA
		result := make([]byte, len(data))
		for i := 0; i < len(data); i += 4 {
			result[i] = data[i+2]   // R
			result[i+1] = data[i+1] // G
			result[i+2] = data[i]   // B
			result[i+3] = data[i+3] // A
		}
		return result, nil
	default:
		return nil, fmt.Errorf("RGBA conversion is not implemented for format %s", format)
	}
}
