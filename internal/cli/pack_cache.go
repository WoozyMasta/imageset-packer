package cli

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/cespare/xxhash/v2"
)

// cacheEntry is a cache entry.
type cacheEntry struct {
	Path string
	Hash string
	Size int64
}

// computeInputsHash computes the hash of the input files.
func computeInputsHash(opts *CmdPack, files []imageFile) (uint64, error) {
	root, err := filepath.Abs(opts.Args.Input)
	if err != nil {
		return 0, fmt.Errorf("resolve input path: %w", err)
	}

	entries := make([]cacheEntry, 0, len(files))
	for _, f := range files {
		absPath, err := filepath.Abs(f.path)
		if err != nil {
			return 0, fmt.Errorf("resolve file path %q: %w", f.path, err)
		}

		rel, err := filepath.Rel(root, absPath)
		if err != nil {
			return 0, fmt.Errorf("resolve relative path for %q: %w", absPath, err)
		}

		fileHash, size, err := hashFileXX(absPath)
		if err != nil {
			return 0, err
		}

		entries = append(entries, cacheEntry{
			Path: filepath.ToSlash(rel),
			Hash: fileHash,
			Size: size,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	h := xxhash.New()
	for _, e := range entries {
		if _, err := h.WriteString(e.Path); err != nil {
			return 0, err
		}
		if _, err := h.Write([]byte{0}); err != nil {
			return 0, err
		}
		if _, err := h.WriteString(e.Hash); err != nil {
			return 0, err
		}
		if _, err := h.Write([]byte{0}); err != nil {
			return 0, err
		}
		if _, err := h.WriteString(strconv.FormatInt(e.Size, 10)); err != nil {
			return 0, err
		}
		if _, err := h.Write([]byte{'\n'}); err != nil {
			return 0, err
		}
	}

	return h.Sum64(), nil
}

// shouldSkipPack checks if the pack should be skipped.
func shouldSkipPack(cachePath, imagesetPath, eddsPath string, nextHash uint64) bool {
	prevHash, ok, err := readCacheHash(cachePath)
	if err != nil || !ok {
		return false
	}
	if prevHash != nextHash {
		return false
	}
	if _, err := os.Stat(imagesetPath); err != nil {
		return false
	}
	if _, err := os.Stat(eddsPath); err != nil {
		return false
	}

	return true
}

// readCacheHash reads the cache hash from the file.
func readCacheHash(path string) (uint64, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("read cache: %w", err)
	}

	if len(data) != 8 {
		return 0, false, nil
	}

	return binary.LittleEndian.Uint64(data), true, nil
}

// writeCacheHash writes the cache hash to the file.
func writeCacheHash(path string, hash uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, hash)
	if err := os.WriteFile(path, buf, 0600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}

	return nil
}

// hashFileXX hashes the file using XXHash.
func hashFileXX(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("open %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("stat %q: %w", path, err)
	}

	h := xxhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", 0, fmt.Errorf("hash %q: %w", path, err)
	}

	return fmt.Sprintf("%016x", h.Sum64()), info.Size(), nil
}
