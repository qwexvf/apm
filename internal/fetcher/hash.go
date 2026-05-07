package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

type hashReader struct {
	r io.Reader
	h hash.Hash
}

func newHashReader(r io.Reader) *hashReader {
	return &hashReader{r: r, h: sha256.New()}
}

func (hr *hashReader) Read(p []byte) (int, error) {
	n, err := hr.r.Read(p)
	if n > 0 {
		hr.h.Write(p[:n])
	}
	return n, err
}

func (hr *hashReader) HexSum() string {
	return hex.EncodeToString(hr.h.Sum(nil))
}
