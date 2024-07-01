package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type StringFilters struct {
	StartsWith []string
	Extensions []string
}

type Filters struct {
	PathFilters       StringFilters
	FileNameFilters   StringFilters
	MaxFileSize       int
	SkipNonExecutable bool
}

type Config struct {
	ExcludedExtensions []string `yaml:"exclude_extensions"`
	ExcludedPaths      []string `yaml:"exclude_paths"`
	MaxFileSize        int      `yaml:"max_file_size"`
	SkipNonExecutable  bool     `yaml:"skip_non_executable"`
}

func Config2Filter(cfg Config) Filters {
	return Filters{
		PathFilters: StringFilters{
			StartsWith: cfg.ExcludedPaths,
		},
		FileNameFilters: StringFilters{
			Extensions: cfg.ExcludedExtensions,
		},
		MaxFileSize:       cfg.MaxFileSize,
		SkipNonExecutable: cfg.SkipNonExecutable,
	}
}

func ParseConfig(configPath string) (Config, error) {
	config := Config{}

	if len(configPath) <= 0 {
		return config, errors.New("no config file")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)

	return config, err
}

func (c *StringFilters) IsExcludedPath(filePath string) bool {
	for _, prefix := range c.StartsWith {
		if strings.HasPrefix(filePath, prefix) {
			return true
		}
	}
	return false
}

func (c *StringFilters) IsExcludedExtension(filePath string) bool {
	for _, ext := range c.Extensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}
	return false
}
