package extractor

import (
	"compress/gzip"
	"io"
	"os"
)

// Check if the tarball is compressed
// If so, use gzip reader
func UnzipIfCompressed(f *os.File) (io.Reader, error) {
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		_, err = f.Seek(0, 0)
		return f, err
	}
	return gzipReader, nil
}
