package extractor

import (
	"io"
	"os"
)

type ExtractedFile struct {
	Filename        string
	Content         io.ReadSeeker
	ContentSize     int
	FilePermissions os.FileMode
	Cleanup         func()
}

func (ef *ExtractedFile) Close() {
	if ef.Cleanup != nil {
		ef.Cleanup()
	}
}

type FileExtractor interface {
	NextFile() (ExtractedFile, error)
	Close() error
}
