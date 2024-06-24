package extractor

import (
	"archive/tar"
	"bufio"
	"os"
	"strings"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/vessel"
)

type ContainerExtractor struct {
	runtime     vessel.Runtime
	tarReader   *tar.Reader
	rootFile    string
	matchConfig *config.Config
}

func NewContainerExtractor(configPath, containerNamespace, containerID string) (*ContainerExtractor, error) {
	runtime, err := vessel.NewRuntime()
	if err != nil {
		return nil, err
	}

	rootFile, err := GetTmpDir(strings.Join([]string{containerNamespace, containerID}, "-"))
	if err != nil {
		return nil, err
	}

	cfg, err := config.ParseConfig(configPath)
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
		runtime:     runtime,
		tarReader:   tr,
		matchConfig: cfg,
	}, nil

}

func (ce *ContainerExtractor) NextFile() (ExtractedFile, error) {
	var (
		h   *tar.Header
		err error
	)
	for {
		h, err = ce.tarReader.Next()
		if err != nil {
			return ExtractedFile{}, err
		}

		if h.Typeflag == tar.TypeDir {
			continue
		}

		if ce.matchConfig.IsExcludedContainerFile(h.Name) {
			continue
		}

		break
	}
	return ExtractedFile{
		Filename: h.Name,
		Content:  bufio.NewReader(ce.tarReader),
	}, err
}

func (ce *ContainerExtractor) Close() error {
	return os.Remove(ce.rootFile)
}
