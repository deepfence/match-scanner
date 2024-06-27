package extractor

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/vessel"
)

type ContainerExtractor struct {
	runtime   vessel.Runtime
	tarReader *tar.Reader
	rootFile  string
	filters   config.Filters
}

func NewContainerExtractor(filters config.Filters, containerNamespace, containerID string) (*ContainerExtractor, error) {
	runtime, err := vessel.NewRuntime()
	if err != nil {
		return nil, err
	}

	rootFile, err := GetTmpDir(strings.Join([]string{containerNamespace, containerID}, "-"))
	if err != nil {
		return nil, err
	}

	err = runtime.ExtractFileSystemContainer(
		containerID,
		containerNamespace,
		rootFile)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(rootFile)
	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(f)

	return &ContainerExtractor{
		runtime:   runtime,
		tarReader: tr,
		filters:   filters,
		rootFile:  rootFile,
	}, nil

}

func (ce *ContainerExtractor) NextFile() (ExtractedFile, error) {
	h, err := ce.tarReader.Next()
	if err != nil {
		return ExtractedFile{}, err
	}
	if ce.filters.PathFilters.IsExcludedPath(h.Name) {
		return ExtractedFile{}, io.EOF
	}
	if ce.filters.FileNameFilters.IsExcludedExtension(h.Name) {
		return ExtractedFile{}, io.EOF
	}
	if ce.filters.MaxFileSize != 0 && h.Size > int64(ce.filters.MaxFileSize) {
		return ExtractedFile{}, io.EOF
	}
	buf := make([]byte, 0, h.Size)
	io.Copy(bytes.NewBuffer(buf), ce.tarReader)
	return ExtractedFile{
		Filename:    filepath.Join("/", h.Name),
		Content:     bytes.NewReader(buf),
		ContentSize: int(h.Size),
	}, err
}

func (ce *ContainerExtractor) Close() error {
	return os.Remove(ce.rootFile)
}
