package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	asset "github.com/flownative/localbeach/assets"
)

func copyFileFromAssets(src, dst string) (int64, error) {
	source, err := asset.Assets.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	ensureDirectoryForFileExists(dst)
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func ensureDirectoryForFileExists(fileName string) {
	directoryName := filepath.Dir(fileName)
	if _, serr := os.Stat(directoryName); serr != nil {
		err := os.MkdirAll(directoryName, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func readFileFromAssets(src string) string {
	source, err := asset.Assets.Open(src)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(source)
	return buf.String()
}
