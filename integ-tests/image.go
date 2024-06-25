package main

import (
	"context"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/match-scanner/pkg/extractor"
	"github.com/deepfence/match-scanner/pkg/scanner"
)

func main() {
	imageID := "e0c9858e10ed"

	cfg, err := config.ParseConfig("integ-test/config.yaml")
	if err != nil {
		println(err.Error())
		return
	}

	extract, err := extractor.NewImageExtractor(config.Config2Filter(cfg), "", imageID)
	if err != nil {
		println(err.Error())
		return
	}
	defer extract.Close()

	scanner.ApplyScan(context.Background(), extract, func(f extractor.ExtractedFile) {
		println(f.Filename)
	})
}
