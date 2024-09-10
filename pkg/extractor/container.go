package extractor

import (
	"context"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/vessel"
)

type ContainerExtractor struct {
	runtime  vessel.Runtime
	tfs      fs.FS
	rootFile string
	filters  config.Filters
	ctx      context.Context
	cancel   context.CancelFunc
	files    chan fileErr
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

	tfs, ctx, cancel, files, err := WalkLayer(f, filters)
	if err != nil {
		return nil, err
	}

	return &ContainerExtractor{
		runtime:  runtime,
		tfs:      tfs,
		filters:  filters,
		rootFile: rootFile,
		ctx:      ctx,
		cancel:   cancel,
		files:    files,
	}, nil

}

func (ce *ContainerExtractor) NextFile() (ExtractedFile, error) {
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

func (ce *ContainerExtractor) Close() error {
	ce.cancel()
	return os.Remove(ce.rootFile)
}
