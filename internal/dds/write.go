// Package dds provides DDS writing functionality.
package dds

import (
	"encoding/binary"
	"io"
)

// writeDWORD writes a 32-bit little-endian value.
func writeDWORD(w io.Writer, v uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

// WriteMagic writes DDS magic to writer.
func WriteMagic(w io.Writer) error {
	_, err := w.Write([]byte(Magic))
	return err
}

// WriteHeader writes DDS header to writer (without magic).
func WriteHeader(w io.Writer, h *Header) error {
	if err := writeDWORD(w, h.Size); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Flags); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Height); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Width); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PitchOrLinearSize); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Depth); err != nil {
		return err
	}
	if err := writeDWORD(w, h.MipMapCount); err != nil {
		return err
	}

	// Write reserved1
	for i := 0; i < 11; i++ {
		if err := writeDWORD(w, h.Reserved1[i]); err != nil {
			return err
		}
	}

	// Write pixel format
	if err := writeDWORD(w, h.PixelFormat.Size); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.Flags); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.FourCC); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.RGBBitCount); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.RBitMask); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.GBitMask); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.BBitMask); err != nil {
		return err
	}
	if err := writeDWORD(w, h.PixelFormat.ABitMask); err != nil {
		return err
	}

	// Write caps
	if err := writeDWORD(w, h.Caps); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Caps2); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Caps3); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Caps4); err != nil {
		return err
	}
	if err := writeDWORD(w, h.Reserved2); err != nil {
		return err
	}

	return nil
}

// WriteHeaderDx10 writes DX10 header to writer.
func WriteHeaderDx10(w io.Writer, dx10 *HeaderDx10) error {
	if err := writeDWORD(w, dx10.DXGIFormat); err != nil {
		return err
	}
	if err := writeDWORD(w, dx10.ResourceDimension); err != nil {
		return err
	}
	if err := writeDWORD(w, dx10.MiscFlag); err != nil {
		return err
	}
	if err := writeDWORD(w, dx10.ArraySize); err != nil {
		return err
	}
	if err := writeDWORD(w, dx10.MiscFlags2); err != nil {
		return err
	}
	return nil
}
