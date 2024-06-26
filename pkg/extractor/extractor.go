package extractor

import (
	"bufio"
)

type ExtractedFile struct {
	Filename    string
	Content     *bufio.Reader
	ContentSize int
	Cleanup     func()
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
