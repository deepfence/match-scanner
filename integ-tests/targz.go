package main

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/deepfence/match-scanner/pkg/config"
	"github.com/deepfence/match-scanner/pkg/extractor"
	"github.com/deepfence/match-scanner/pkg/scanner"
)

func main() {
	imgTar := "/tmp/xmrig/image.tar"
	imgTarGz := "/tmp/xmrig/image.tar.gz"
	imageName := "metal3d/xmrig:latest"
	cmd := exec.Command("skopeo", "copy", "--insecure-policy", "--src-tls-verify=false",
		"docker://"+imageName, "docker-archive:"+imgTar)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	errorOnRun := cmd.Run()
	if errorOnRun != nil {
		println("Error: ", errorOnRun.Error())
		println("stderr: ", stderr.String())
		return
	}

	cmd = exec.Command("gzip", imgTar, imgTarGz)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	errorOnRun = cmd.Run()
	if errorOnRun != nil {
		println("Error: ", errorOnRun.Error())
		println("stderr: ", stderr.String())
		return
	}

	cfg, err := config.ParseConfig("integ-tests/config.yaml")
	if err != nil {
		println(err.Error())
		return
	}

	extract, err := extractor.NewTarExtractor(config.Config2Filter(cfg), "", imgTarGz)
	if err != nil {
		println(err.Error())
		return
	}
	defer extract.Close()

	scanner.ApplyScan(context.Background(), extract, func(f extractor.ExtractedFile) {
		println(f.Filename)
	})
}
