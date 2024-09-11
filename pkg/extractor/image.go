package extractor

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/vessel"
)

type ImageExtractor struct {
	runtime        vessel.Runtime
	tarReader      *tar.Reader
	layerTarReader fs.FS
	lastLayerErr   error
	rootFile       string
	filters        config.Filters
	ctx            context.Context
	cancel         context.CancelFunc
	files          chan fileErr
}

func NewImageExtractor(filters config.Filters, imageNamespace, imageID string) (*ImageExtractor, error) {
	runtime, err := vessel.NewRuntime()
	if err != nil {
		return nil, err
	}

	rootFile, err := GetTmpDir(strings.Join([]string{imageNamespace, imageID}, "-"))
	if err != nil {
		return nil, err
	}

	_, err = runtime.Save(
		imageID,
		rootFile)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(rootFile)
	if err != nil {
		return nil, err
	}

	reader, err := UnzipIfCompressed(f)
	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(reader)

	return &ImageExtractor{
		runtime:   runtime,
		tarReader: tr,
		filters:   filters,
		rootFile:  rootFile,
	}, nil

}

func (ce *ImageExtractor) nextLayerFile() (ExtractedFile, error) {
	fErr, opened := <-ce.files

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

func (ce *ImageExtractor) NextFile() (ExtractedFile, error) {
	for {
		if ce.layerTarReader != nil {
			file, err := ce.nextLayerFile()
			if err != nil {
				ce.layerTarReader = nil
				goto next_layer
			}
			return file, nil
		}
	next_layer:
		_, err := ce.tarReader.Next()
		if err != nil {
			return ExtractedFile{}, err
		}
		ce.layerTarReader, ce.ctx, ce.cancel, ce.files, err = WalkLayer(ce.tarReader, ce.filters)
		if err != nil {
			fmt.Printf("err: %v", err)
		}
	}
}

func (ce *ImageExtractor) Close() error {
	return os.Remove(ce.rootFile)
}
