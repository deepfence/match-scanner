package extractor

import (
	"archive/tar"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/vessel"
)

type ImageExtractor struct {
	runtime        vessel.Runtime
	tarReader      *tar.Reader
	layerTarReader *tar.Reader
	lastLayerErr   error
	rootFile       string
	filters        config.Filters
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

	tr := tar.NewReader(f)

	return &ImageExtractor{
		runtime:   runtime,
		tarReader: tr,
		filters:   filters,
		rootFile:  rootFile,
	}, nil

}

func (ce *ImageExtractor) nextLayerFile() (ExtractedFile, error) {
	h, err := ce.layerTarReader.Next()
	if err != nil {
		return ExtractedFile{}, err
	}
	if ce.filters.PathFilters.IsExcludedPath(h.Name) {
		return ExtractedFile{}, io.EOF
	}
	if ce.filters.FileNameFilters.IsExcludedExtension(h.Name) {
		return ExtractedFile{}, io.EOF
	}
	if h.Size > int64(ce.filters.MaxFileSize) {
		return ExtractedFile{}, io.EOF
	}
	return ExtractedFile{
		Filename: filepath.Join("/", h.Name),
		Content:  bufio.NewReader(ce.tarReader),
	}, err
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
		ce.layerTarReader = tar.NewReader(ce.tarReader)
	}
}

func (ce *ImageExtractor) Close() error {
	return os.Remove(ce.rootFile)
}
