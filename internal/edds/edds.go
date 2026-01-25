// Package edds provides EDDS functions.
package edds

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pierrec/lz4/v4"
)

const (
	// BlockMagicCOPY is the magic string for uncompressed blocks.
	BlockMagicCOPY = "COPY"
	// BlockMagicLZ4 is the magic string for LZ4 compressed blocks.
	BlockMagicLZ4 = "LZ4 "

	// ChunkSize defines the uncompressed data size per chunk.
	// 64KB is the standard Enfusion chunk size.
	ChunkSize = 64 * 1024

	maxInt32 = int(^uint32(0) >> 1)
)

// Block represents a data block for a mipmap.
type Block struct {
	Magic            string // "COPY" or "LZ4 "
	Data             []byte // Body data
	Size             int32  // Total size of the block body
	UncompressedSize int32  // Total uncompressed size
}

// writeBlockData writes block data (inner body).
func writeBlockData(w io.Writer, block *Block) error {
	if block.Magic == BlockMagicLZ4 {
		// Write UncompressedSize
		if err := binary.Write(w, binary.LittleEndian, block.UncompressedSize); err != nil {
			return fmt.Errorf("writing uncompressed size: %w", err)
		}
		// Write Chunk Stream
		if _, err := w.Write(block.Data); err != nil {
			return fmt.Errorf("writing chunk stream: %w", err)
		}
	} else {
		// Write Raw Data
		if _, err := w.Write(block.Data); err != nil {
			return fmt.Errorf("writing block data: %w", err)
		}
	}
	return nil
}

// compressBlock compresses data into 64KB chunks using LZ4 HC.
func compressBlock(data []byte) (*Block, error) {
	if len(data) > maxInt32 {
		return nil, fmt.Errorf("input data too large: %d bytes", len(data))
	}
	uncompressedSize := int32(len(data)) //nolint:gosec // Guarded by size check above.

	// 1. Threshold Check.
	// If the data is smaller than 1KB, use COPY.
	// Small LZ4 blocks often cause overhead and parser issues.
	if len(data) < 1024 {
		return &Block{
			Magic:            BlockMagicCOPY,
			Size:             uncompressedSize,
			UncompressedSize: 0,
			Data:             data,
		}, nil
	}

	var chunkStream bytes.Buffer

	// Pre-allocate buffer for compression (reused)
	maxCompressedSize := lz4.CompressBlockBound(ChunkSize)
	compressBuf := make([]byte, maxCompressedSize)

	totalCompressedPayload := 0

	// 2. Iterate in 64KB chunks
	for i := 0; i < len(data); i += ChunkSize {
		end := i + ChunkSize
		if end > len(data) {
			end = len(data)
		}

		srcChunk := data[i:end]
		isLast := end == len(data)

		// 3. Compress using High Compression (Level 0 = Default HC)
		cn, err := lz4.CompressBlockHC(srcChunk, compressBuf, 0, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("LZ4 compression failed: %w", err)
		}

		// PARANOID CHECK:
		// If a chunk didn't compress well (e.g. > 85% of original size),
		// abort the whole operation and fallback to COPY.
		// "Bad" compressed chunks are the #1 cause of parser desync.
		if cn == 0 || float64(cn) > float64(len(srcChunk))*0.85 {
			return &Block{
				Magic:            BlockMagicCOPY,
				Size:             uncompressedSize,
				UncompressedSize: 0,
				Data:             data,
			}, nil
		}

		if cn > 0x7FFFFF {
			return nil, fmt.Errorf("compressed chunk too large: %d", cn)
		}

		// Write Chunk Header: Size (3 bytes) + Flags (1 byte)
		chunkStream.WriteByte(byte(cn))
		chunkStream.WriteByte(byte(cn >> 8))
		chunkStream.WriteByte(byte(cn >> 16))

		if isLast {
			chunkStream.WriteByte(0x80)
		} else {
			chunkStream.WriteByte(0x00)
		}

		// Write Compressed Data
		chunkStream.Write(compressBuf[:cn])
		totalCompressedPayload += cn
	}

	compressedData := chunkStream.Bytes()
	totalOverhead := 4 + len(compressedData) // 4 bytes for UncompressedSize header
	if totalOverhead > maxInt32 {
		return nil, fmt.Errorf("compressed data too large: %d bytes", totalOverhead)
	}

	// 4. Global Fallback Check
	// If total size isn't significantly smaller (at least 15% saving), use COPY.
	if float64(totalOverhead) > float64(len(data))*0.85 {
		return &Block{
			Magic:            BlockMagicCOPY,
			Size:             uncompressedSize,
			UncompressedSize: 0,
			Data:             data,
		}, nil
	}

	return &Block{
		Magic:            BlockMagicLZ4,
		Size:             int32(totalOverhead), //nolint:gosec // Guarded by size check above.
		UncompressedSize: uncompressedSize,
		Data:             compressedData,
	}, nil
}

// decompressBlock decompresses an EDDS block.
// For LZ4 blocks DayZ uses Enfusion "chunk-stream":
// [u32 targetSize][ (int24 cSize) (u8 flags) (cSize bytes compressed) ]...
// Decoder is a CHAIN decoder with 64KB rolling dictionary.
func decompressBlock(block *Block, expectedUncompressedSize int) ([]byte, error) {
	if block.Magic == BlockMagicCOPY {
		if len(block.Data) != expectedUncompressedSize {
			return nil, fmt.Errorf("COPY block size mismatch: expected %d, got %d", expectedUncompressedSize, len(block.Data))
		}
		out := make([]byte, len(block.Data))
		copy(out, block.Data)
		return out, nil
	}

	if block.Magic != BlockMagicLZ4 {
		return nil, fmt.Errorf("unknown block magic: %q", block.Magic)
	}

	targetSize := expectedUncompressedSize
	if block.UncompressedSize > 0 {
		targetSize = int(block.UncompressedSize)
	}
	if targetSize <= 0 {
		return nil, fmt.Errorf("invalid target size: %d", targetSize)
	}

	data := block.Data

	// Some EDDS store targetSize inside payload: [u32 targetSize][chunkstream...]
	// TS viewer always expects it.
	if len(data) >= 8 {
		peek := int(binary.LittleEndian.Uint32(data[:4]))
		// If peek equals expected full mip size => it's very likely the embedded targetSize
		// And the next 3 bytes must look like a sane int24 chunk size (< 1MB)
		c0 := int(data[4]) | (int(data[5]) << 8) | (int(data[6]) << 16)
		if (peek == expectedUncompressedSize || peek == targetSize) && c0 > 0 && c0 < (1<<20) {
			targetSize = peek
			data = data[4:]
		}
	}

	// Chain decoder with rolling 64KB dict
	const dictCap = 64 * 1024
	dict := make([]byte, dictCap)
	dictSize := 0

	target := make([]byte, targetSize)
	outIdx := 0

	r := bytes.NewReader(data)

	for {
		if r.Len() < 4 {
			return nil, fmt.Errorf("LZ4 chunk-stream truncated (need 4 bytes header, have %d)", r.Len())
		}

		var hdr [4]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil {
			return nil, fmt.Errorf("reading chunk header: %w", err)
		}

		cSize := int(hdr[0]) | (int(hdr[1]) << 8) | (int(hdr[2]) << 16)
		flags := hdr[3]

		if (flags &^ 0x80) != 0 {
			return nil, fmt.Errorf("unknown LZ4 flags: 0x%02x", flags)
		}
		if cSize <= 0 || cSize > r.Len() {
			return nil, fmt.Errorf("invalid compressed chunk size: %d (remaining %d)", cSize, r.Len())
		}

		compressed := make([]byte, cSize)
		if _, err := io.ReadFull(r, compressed); err != nil {
			return nil, fmt.Errorf("reading chunk data: %w", err)
		}

		remaining := targetSize - outIdx
		if remaining <= 0 {
			return nil, fmt.Errorf("decoded LZ4 overruns target buffer")
		}

		want := ChunkSize
		if want > remaining {
			want = remaining
		}
		dst := target[outIdx : outIdx+want]

		n, err := lz4.UncompressBlockWithDict(compressed, dst, dict[:dictSize])
		if err != nil {
			return nil, fmt.Errorf("LZ4 chunk decode failed: %w", err)
		}

		outIdx += n

		// update rolling dict
		decoded := target[outIdx-n : outIdx]
		if len(decoded) >= dictCap {
			copy(dict, decoded[len(decoded)-dictCap:])
			dictSize = dictCap
		} else {
			avail := dictCap - dictSize
			if len(decoded) <= avail {
				copy(dict[dictSize:], decoded)
				dictSize += len(decoded)
			} else {
				shift := len(decoded) - avail
				copy(dict, dict[shift:dictSize])
				copy(dict[dictCap-len(decoded):], decoded)
				dictSize = dictCap
			}
		}

		if (flags & 0x80) != 0 {
			break
		}
	}

	if outIdx != targetSize {
		return nil, fmt.Errorf("LZ4 decoded size mismatch: expected %d, got %d", targetSize, outIdx)
	}
	if r.Len() != 0 {
		return nil, fmt.Errorf("LZ4 block length mismatch: %d bytes left after decode", r.Len())
	}

	return target, nil
}

// blockHeader is Magic+Size entry from EDDS header table.
type blockHeader struct {
	Magic string
	Size  int32
}

// readBlockTable reads mipMapCount headers: [Magic(4)][Size(i32)]...
func readBlockTable(r io.Reader, mipMapCount uint32) ([]blockHeader, error) {
	hdrs := make([]blockHeader, 0, mipMapCount)

	for i := uint32(0); i < mipMapCount; i++ {
		magicBytes := make([]byte, 4)
		if _, err := io.ReadFull(r, magicBytes); err != nil {
			return nil, fmt.Errorf("reading block table magic %d: %w", i, err)
		}
		magic := string(magicBytes)

		var size int32
		if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
			return nil, fmt.Errorf("reading block table size %d: %w", i, err)
		}

		if magic != BlockMagicCOPY && magic != BlockMagicLZ4 {
			return nil, fmt.Errorf("unknown block magic in table %d: %q", i, magic)
		}
		if size < 0 {
			return nil, fmt.Errorf("invalid block size in table %d: %d", i, size)
		}

		hdrs = append(hdrs, blockHeader{Magic: magic, Size: size})
	}

	return hdrs, nil
}

// readBlockBody reads a block body using header info (Magic+Size).
func readBlockBody(r io.Reader, h blockHeader) (*Block, error) {
	if h.Size < 0 {
		return nil, fmt.Errorf("invalid block size: %d", h.Size)
	}

	data := make([]byte, h.Size)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("reading %s body: %w", h.Magic, err)
	}

	return &Block{
		Magic:            h.Magic,
		Size:             h.Size,
		UncompressedSize: 0,    // Filled by decoder when needed.
		Data:             data, // For LZ4 this is the raw payload.
	}, nil
}
