package extractor

import (
	"os"
	"path/filepath"
	"regexp"
)

var (
	TempPluginNamespace = "deepfence"
	TempPluginName      = "scanner"
)

func GetTmpDir(fileName string) (string, error) {
	tempDirectory := os.TempDir()
	scanID := "df_" + getSanitizedString(fileName)

	tempPath := filepath.Join(tempDirectory, TempPluginNamespace, TempPluginName)

	err := CreateRecursiveDir(tempPath)
	if err != nil {
		return "", err
	}

	return filepath.Join(tempPath, scanID), err
}

func getSanitizedString(imageName string) string {
	//nolint:gocritic
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return "error"
	}
	sanitizedName := reg.ReplaceAllString(imageName, "")
	return sanitizedName
}

func CreateRecursiveDir(completePath string) error {
	if _, err := os.Stat(completePath); os.IsNotExist(err) {
		err = os.MkdirAll(completePath, os.ModePerm)
		return err
	} else if err != nil {
		_ = os.RemoveAll(completePath)
		return err
	}

	return nil
}
