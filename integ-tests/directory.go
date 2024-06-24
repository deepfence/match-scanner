package main

import (
	"context"

	"github.com/deepfence/match-scanner/pkg/extractor"
	"github.com/deepfence/match-scanner/pkg/scanner"
)

func main() {
	root := "/"
	extract, err := extractor.NewDirectoryExtractor(root)
	if err != nil {
		println(err.Error())
		return
	}
	defer extract.Close()

	scanner.ApplyScan(context.Background(), extract, func(f extractor.ExtractedFile) {
		println(f.Filename)
	})
}
