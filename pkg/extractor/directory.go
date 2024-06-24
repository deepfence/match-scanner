package extractor

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	MAX_OPEN_FILE = 5
)

type fileErr struct {
	f   *os.File
	err error
}

type DirectoryExtractor struct {
	rootDir string
	files   chan fileErr
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewDirectoryExtractor(rootDir string) (*DirectoryExtractor, error) {

	files := make(chan fileErr, MAX_OPEN_FILE)
	visited := make(map[string]struct{})
	ctx, cancel := context.WithCancel(context.Background())

	var visit func(path string, d fs.DirEntry, err error) error
	visit = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println(err.Error())
			return nil
		}
		info, err := d.Info()
		if err != nil {
			println(err.Error())
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				println(err.Error())
				return nil
			}

			absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), linkTarget))
			if err != nil {
				println(err.Error())
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
			f, e := os.Open(path)
			select {
			case files <- fileErr{
				f:   f,
				err: e,
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
			println("err:", err.Error())
		}
		close(files)
	}()

	return &DirectoryExtractor{
		files:  files,
		ctx:    ctx,
		cancel: cancel,
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
	}, nil
}

func (ce *DirectoryExtractor) Close() error {
	ce.cancel()
	return nil
}
