package config

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = "config.yaml"
)

type Config struct {
	ExcludedExtensions     []string `yaml:"exclude_extensions"`
	ExcludedPaths          []string `yaml:"exclude_paths"`
	ExcludedContainerPaths []string `yaml:"exclude_container_paths"`
}

func ParseConfig(configPath string) (*Config, error) {
	config := &Config{}
	var (
		data []byte
		err  error
	)

	if len(configPath) > 0 {
		fileInfo, err := os.Stat(configPath)
		if err != nil {
			return config, err
		}

		if fileInfo.IsDir() {
			configPath = path.Join(configPath, ConfigFileName)
		}

		data, err = os.ReadFile(configPath)
		if err != nil {
			return config, err
		}
	} else {
		configPath, err = defaultConfigLookup()
		if err != nil {
			// by default, no config file is not an error
			return config, nil
		}

		data, err = os.ReadFile(configPath)
		if err != nil {
			return config, err
		}
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config, err
	}

	pathSeparator := string(os.PathSeparator)
	var excludedPaths []string
	for _, path := range config.ExcludedPaths {
		excludedPaths = append(excludedPaths, strings.ReplaceAll(path, "{sep}", pathSeparator))
	}
	config.ExcludedPaths = excludedPaths

	var excludedContainerPaths []string
	for _, path := range config.ExcludedContainerPaths {
		excludedContainerPaths = append(excludedContainerPaths, strings.ReplaceAll(path, "{sep}", pathSeparator))
	}
	config.ExcludedContainerPaths = excludedContainerPaths

	return config, nil
}

func defaultConfigLookup() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(ex)
	configPath := path.Join(dir, ConfigFileName)
	_, err = os.ReadFile(configPath)
	if err != nil {
		dir, _ = os.Getwd()
		configPath = path.Join(dir, ConfigFileName)
		_, err = os.ReadFile(configPath)
		if err != nil {
			return "", err
		}
	}
	return configPath, nil
}

func (c *Config) IsExcludedPath(filePath string) bool {
	for _, path := range c.ExcludedPaths {
		if strings.Contains(filePath, path) {
			return true
		}
	}
	return false
}

func (c *Config) IsExcludedExtension(filePath string) bool {
	for _, ext := range c.ExcludedExtensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}
	return false
}

func (c *Config) IsExcludedContainerFile(filePath string) bool {
	for _, path := range c.ExcludedContainerPaths {
		if strings.HasPrefix(filePath, path) {
			return true
		}
	}
	return c.IsExcludedExtension(filePath)
}
