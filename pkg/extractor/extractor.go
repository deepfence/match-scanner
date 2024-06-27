package extractor

import "io"

type ExtractedFile struct {
	Filename    string
	Content     io.ReadSeeker
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
