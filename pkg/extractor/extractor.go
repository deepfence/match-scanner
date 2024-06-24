package extractor

import (
	"bufio"
)

type ExtractedFile struct {
	Filename string
	Content  *bufio.Reader
}

func (ef *ExtractedFile) Close() {}

type FileExtractor interface {
	NextFile() (ExtractedFile, error)
	Close() error
}
