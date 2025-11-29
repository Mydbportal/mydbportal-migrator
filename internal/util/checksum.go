package util

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// ComputeChecksum calculates the SHA256 checksum of a file.
func ComputeChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
