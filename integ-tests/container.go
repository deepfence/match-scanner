package main

import (
	"context"

	"github.com/deepfence/match-scanner/pkg/extractor"
	"github.com/deepfence/match-scanner/pkg/scanner"
)

func main() {
	containerID := "9612e7b41bc5"
	extract, err := extractor.NewContainerExtractor("", "", containerID)
	if err != nil {
		println(err.Error())
		return
	}
	defer extract.Close()

	scanner.ApplyScan(context.Background(), extract, func(f extractor.ExtractedFile) {
		// println(f.Filename)
	})
}
