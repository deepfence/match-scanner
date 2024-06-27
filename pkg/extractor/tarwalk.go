package extractor

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/match-scanner/pkg/log"
	"github.com/nlepage/go-tarfs"
)

func WalkLayer(rootFile io.Reader, filters config.Filters) (fs.FS, context.Context, context.CancelFunc, chan fileErr, error) {

	tfs, err := tarfs.New(rootFile)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	files := make(chan fileErr, MAX_OPEN_FILE)

	visit := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.ErrLogger(err)
			return nil
		}

		fullPath := filepath.Join("/", path)

		info, err := d.Info()
		if err != nil {
			log.ErrLogger(err)
		}

		if d.IsDir() && filters.PathFilters.IsExcludedPath(fullPath) {
			return fs.SkipDir
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if !d.IsDir() {
			if filters.FileNameFilters.IsExcludedExtension(fullPath) {
				return nil
			}

			if filters.MaxFileSize != 0 && info.Size() > int64(filters.MaxFileSize) {
				return nil
			}

			f, e := tfs.Open(path)

			if e != nil {
				return nil
			}

			select {
			case files <- fileErr{
				fpath: fullPath,
				f:     f,
				err:   e,
				fsize: int(info.Size()),
			}:
			case <-ctx.Done():
				return io.EOF
			}
		}
		return nil
	}

	go func() {
		err := fs.WalkDir(tfs, ".", visit)
		if err != nil {
			log.ErrLogger(err)
		}
		close(files)
	}()

	return tfs, ctx, cancel, files, nil
}
