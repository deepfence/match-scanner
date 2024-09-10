package extractor

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/deepfence/match-scanner/pkg/config"
)

type TarExtractor struct {
	tarReader      *tar.Reader
	layerTarReader fs.FS
	lastLayerErr   error
	tarPath        string
	filters        config.Filters
	ctx            context.Context
	cancel         context.CancelFunc
	files          chan fileErr
}

func NewTarExtractor(filters config.Filters, imageNamespace, tarPath string) (*TarExtractor, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}

	var reader io.Reader = f

	// Check if the tarball is compressed
	// If so, use gzip reader
	gzipReader, err := gzip.NewReader(f)
	if err == nil {
		reader = gzipReader
	}

	tr := tar.NewReader(reader)

	return &TarExtractor{
		tarReader: tr,
		filters:   filters,
		tarPath:   tarPath,
	}, nil

}

func (te *TarExtractor) nextLayerFile() (ExtractedFile, error) {
	fErr, opened := <-te.files

	if !opened {
		return ExtractedFile{}, io.EOF
	}

	if fErr.err != nil {
		return ExtractedFile{}, fErr.err
	}

	return ExtractedFile{
		Filename:    fErr.fpath,
		Content:     fErr.f.(io.ReadSeeker),
		ContentSize: int(fErr.fsize),
		Cleanup: func() {
			fErr.f.Close()
		},
	}, nil
}

func (te *TarExtractor) NextFile() (ExtractedFile, error) {
	for {
		if te.layerTarReader != nil {
			file, err := te.nextLayerFile()
			if err != nil {
				te.layerTarReader = nil
				goto next_layer
			}
			return file, nil
		}
	next_layer:
		_, err := te.tarReader.Next()
		if err != nil {
			return ExtractedFile{}, err
		}
		te.layerTarReader, te.ctx, te.cancel, te.files, err = WalkLayer(te.tarReader, te.filters)
		if err != nil {
			fmt.Printf("WalkLayer error: %v\n", err)
		}
	}
}

func (te *TarExtractor) Close() error {
	return os.Remove(te.tarPath)
}
