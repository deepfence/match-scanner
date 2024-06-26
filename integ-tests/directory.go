package main

import (
	"context"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/match-scanner/pkg/extractor"
	"github.com/deepfence/match-scanner/pkg/scanner"
)

func main() {
	root := "/tmp"

	cfg, err := config.ParseConfig("integ-tests/config.yaml")
	if err != nil {
		println(err.Error())
		return
	}

	extract, err := extractor.NewDirectoryExtractor(config.Config2Filter(cfg), root)
	if err != nil {
		println(err.Error())
		return
	}
	defer extract.Close()

	scanner.ApplyScan(context.Background(), extract, func(f extractor.ExtractedFile) {
		println(f.Filename)
	})
}
