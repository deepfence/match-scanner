package extractor

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/match-scanner/pkg/log"
)

const (
	MAX_OPEN_FILE = 5
)

type fileErr struct {
	f     fs.File
	fpath string
	fsize int
	fperm os.FileMode
	err   error
}

type DirectoryExtractor struct {
	rootDir string
	files   chan fileErr
	ctx     context.Context
	cancel  context.CancelFunc
}

func removeRootDir(path, rootDir string) string {
	if rootDir == "/" {
		return path
	}
	return strings.Replace(path, rootDir, "", 1)
}

func resolveSymlink(path string) string {
	linkTarget, err := os.Readlink(path)
	if err != nil {
		log.ErrLogger(err)
		return ""
	}

	absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), linkTarget))
	if err != nil {
		log.ErrLogger(err)
		return ""
	}
	return absTarget
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

func NewDirectoryExtractor(filters config.Filters, rootDir string, skipSymlink bool) (*DirectoryExtractor, error) {

	files := make(chan fileErr, MAX_OPEN_FILE)
	visited := make(map[string]struct{})
	ctx, cancel := context.WithCancel(context.Background())

	if isSymlink(rootDir) {
		rootDir = resolveSymlink(rootDir)
	}

	var visit func(path string, d fs.DirEntry, err error) error
	visit = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.ErrLogger(err)
			return nil
		}

		info, err := d.Info()
		if err != nil {
			log.ErrLogger(err)
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			if skipSymlink {
				return nil
			}
			absTarget := resolveSymlink(path)

			if _, has := visited[absTarget]; has {
				return nil
			}

			visited[absTarget] = struct{}{}

			return filepath.WalkDir(path, visit)
		}

		if d.IsDir() && filters.PathFilters.IsExcludedPath(removeRootDir(path, rootDir)) {
			return filepath.SkipDir
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if !d.IsDir() {
			if filters.FileNameFilters.IsExcludedExtension(path) {
				return nil
			}

			if filters.MaxFileSize != 0 && info.Size() > int64(filters.MaxFileSize) {
				return nil
			}

			f, e := os.Open(path)

			select {
			case files <- fileErr{
				fpath: path,
				f:     f,
				err:   e,
				fsize: int(info.Size()),
				fperm: info.Mode().Perm(),
			}:
			case <-ctx.Done():
				return io.EOF
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
		files:   files,
		ctx:     ctx,
		cancel:  cancel,
		rootDir: rootDir,
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
		Filename:        fErr.fpath,
		Content:         fErr.f.(io.ReadSeeker),
		ContentSize:     int(fErr.fsize),
		FilePermissions: fErr.fperm,
		Cleanup: func() {
			fErr.f.Close()
		},
	}, nil
}

func (ce *DirectoryExtractor) Close() error {
	ce.cancel()
	return nil
}
