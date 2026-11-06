package storage

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// HashingReader wraps an io.Reader and calculates SHA-256 checksum on the fly.
type HashingReader struct {
	reader io.Reader
	hasher hash.Hash
}

// NewHashingReader creates a streaming reader that computes SHA-256 digest.
func NewHashingReader(r io.Reader) *HashingReader {
	return &HashingReader{
		reader: r,
		hasher: sha256.New(),
	}
}

func (h *HashingReader) Read(p []byte) (int, error) {
	n, err := h.reader.Read(p)
	if n > 0 {
		h.hasher.Write(p[:n])
	}
	return n, err
}

// SumHex returns the hex-encoded SHA-256 checksum of all bytes read so far.
func (h *HashingReader) SumHex() string {
	return hex.EncodeToString(h.hasher.Sum(nil))
}

// DetectCompression determines compression format from filename or magic bytes.
func DetectCompression(filename string, header []byte) string {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".gzip") {
		return "gzip"
	}
	if strings.HasSuffix(lower, ".zst") || strings.HasSuffix(lower, ".zstd") {
		return "zstd"
	}

	if len(header) >= 2 && header[0] == 0x1f && header[1] == 0x8b {
		return "gzip"
	}
	if len(header) >= 4 && header[0] == 0x28 && header[1] == 0xb5 && header[2] == 0x2f && header[3] == 0xfd {
		return "zstd"
	}

	return "none"
}

// OpenDecompressedStream takes an underlying stream and returns a decompressed stream reader.
func OpenDecompressedStream(r io.Reader, filename string) (io.Reader, string, error) {
	bufReader := bufio.NewReader(r)
	peek, _ := bufReader.Peek(8)

	comp := DetectCompression(filename, peek)
	switch comp {
	case "gzip":
		gz, err := gzip.NewReader(bufReader)
		if err != nil {
			return nil, comp, fmt.Errorf("open gzip stream: %w", err)
		}
		return gz, comp, nil

	case "zstd":
		decoder, err := zstd.NewReader(bufReader)
		if err != nil {
			return nil, comp, fmt.Errorf("open zstd stream: %w", err)
		}
		return decoder.IOReadCloser(), comp, nil

	default:
		return bufReader, "none", nil
	}
}
