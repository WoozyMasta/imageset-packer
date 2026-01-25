// Package dds provides DDS reading functionality.
package dds

import (
	"encoding/binary"
	"fmt"
	"io"
)

// readDWORD reads a 32-bit little-endian value.
func readDWORD(r io.Reader) (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// ReadHeader reads DDS header from reader (including magic).
func ReadHeader(r io.Reader) (*Header, error) {
	// Read magic
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, fmt.Errorf("reading magic: %w", err)
	}
	if string(magic) != Magic {
		return nil, fmt.Errorf("invalid magic: expected %q, got %q", Magic, string(magic))
	}

	// Read header size
	size, err := readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading header size: %w", err)
	}
	if size != HeaderSize {
		return nil, fmt.Errorf("invalid header size: expected %d, got %d", HeaderSize, size)
	}

	var h Header
	h.Size = size
	h.Flags, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading flags: %w", err)
	}
	h.Height, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading height: %w", err)
	}
	h.Width, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading width: %w", err)
	}
	h.PitchOrLinearSize, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pitchOrLinearSize: %w", err)
	}
	h.Depth, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading depth: %w", err)
	}
	h.MipMapCount, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading mipMapCount: %w", err)
	}

	// Read reserved1[11]
	for i := 0; i < 11; i++ {
		h.Reserved1[i], err = readDWORD(r)
		if err != nil {
			return nil, fmt.Errorf("reading reserved1[%d]: %w", i, err)
		}
	}

	// Read pixel format
	pfSize, err := readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format size: %w", err)
	}
	if pfSize != PixelFormatSize {
		return nil, fmt.Errorf("invalid pixel format size: expected %d, got %d", PixelFormatSize, pfSize)
	}

	h.PixelFormat.Size = pfSize
	h.PixelFormat.Flags, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format flags: %w", err)
	}
	h.PixelFormat.FourCC, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format fourCC: %w", err)
	}
	h.PixelFormat.RGBBitCount, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format rgbBitCount: %w", err)
	}
	h.PixelFormat.RBitMask, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format rBitMask: %w", err)
	}
	h.PixelFormat.GBitMask, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format gBitMask: %w", err)
	}
	h.PixelFormat.BBitMask, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format bBitMask: %w", err)
	}
	h.PixelFormat.ABitMask, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading pixel format aBitMask: %w", err)
	}

	// Read caps
	h.Caps, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading caps: %w", err)
	}
	h.Caps2, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading caps2: %w", err)
	}
	h.Caps3, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading caps3: %w", err)
	}
	h.Caps4, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading caps4: %w", err)
	}
	h.Reserved2, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading reserved2: %w", err)
	}
	if h.Reserved2 != 0 {
		return nil, fmt.Errorf("invalid header: reserved2 is not zero")
	}

	// Validate flags
	if h.Flags&HeaderFlagsTexture != HeaderFlagsTexture {
		return nil, fmt.Errorf("invalid header flags: required fields not set (flags: 0x%x)", h.Flags)
	}

	return &h, nil
}

// ReadHeaderDx10 reads DX10 header if present.
func ReadHeaderDx10(r io.Reader, header *Header) (*HeaderDx10, error) {
	if (header.PixelFormat.Flags&PFFourCC == 0) || header.PixelFormat.FourCC != FourCCDX10 {
		return nil, nil // No DX10 header
	}

	var dx10 HeaderDx10
	var err error

	dx10.DXGIFormat, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading dxgiFormat: %w", err)
	}
	dx10.ResourceDimension, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading resourceDimension: %w", err)
	}
	dx10.MiscFlag, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading miscFlag: %w", err)
	}
	dx10.ArraySize, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading arraySize: %w", err)
	}
	dx10.MiscFlags2, err = readDWORD(r)
	if err != nil {
		return nil, fmt.Errorf("reading miscFlags2: %w", err)
	}

	return &dx10, nil
}
