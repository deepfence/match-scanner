package extractor

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/match-scanner/pkg/log"
)

const (
	MAX_OPEN_FILE = 5
)

type fileErr struct {
	f   *os.File
	err error
}

type DirectoryExtractor struct {
	rootDir     string
	files       chan fileErr
	ctx         context.Context
	cancel      context.CancelFunc
	matchConfig *config.Config
}

func NewDirectoryExtractor(configPath, rootDir string) (*DirectoryExtractor, error) {

	files := make(chan fileErr, MAX_OPEN_FILE)
	visited := make(map[string]struct{})
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	var visit func(path string, d fs.DirEntry, err error) error
	visit = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.ErrLogger(err)
			return nil
		}
		info, err := d.Info()
		if err != nil {
			log.ErrLogger(err)
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				log.ErrLogger(err)
				return nil
			}

			absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), linkTarget))
			if err != nil {
				log.ErrLogger(err)
				return nil
			}

			if _, has := visited[absTarget]; has {
				return nil
			}

			visited[absTarget] = struct{}{}
			return filepath.WalkDir(path, visit)
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if !d.IsDir() {
			cfg.IsExcludedExtension(path)
			f, e := os.Open(path)
			select {
			case files <- fileErr{
				f:   f,
				err: e,
			}:
			case <-ctx.Done():
				return io.EOF
			}
		} else {
			if cfg.IsExcludedPath(path) {
				return filepath.SkipDir
			}
		}
		return nil
	}

	go func() {
		err := filepath.WalkDir(rootDir, visit)
		if err != nil {
			log.ErrLogger(err)
		}
		close(files)
	}()

	return &DirectoryExtractor{
		files:       files,
		ctx:         ctx,
		cancel:      cancel,
		matchConfig: cfg,
	}, nil

}

func (ce *DirectoryExtractor) NextFile() (ExtractedFile, error) {
	fErr, opened := <-ce.files

	if !opened {
		return ExtractedFile{}, io.EOF
	}

	if fErr.err != nil {
		return ExtractedFile{}, fErr.err
	}

	return ExtractedFile{
		Filename: fErr.f.Name(),
		Content:  bufio.NewReader(fErr.f),
		Cleanup: func() {
			fErr.f.Close()
		},
	}, nil
}

func (ce *DirectoryExtractor) Close() error {
	ce.cancel()
	return nil
}
