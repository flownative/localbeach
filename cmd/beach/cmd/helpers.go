package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func getRelativePersistentResourcePathByHash(hash string) string {
	return string(hash[0]) + "/" + string(hash[1]) + "/" + string(hash[2]) + "/" + string(hash[3])
}

func retrieveCloudStorageCredentials(instanceIdentifier string, projectNamespace string) (err error, bucketName string, privateKey []byte) {
	log.Info("Retrieving cloud storage access data from instance")

	internalHost := "beach@" + instanceIdentifier + "." + projectNamespace
	output, err := exec.RunCommand("ssh", []string{
		"-J", "beach@ssh.flownative.cloud", internalHost,
		"/bin/bash", "-c", "env | grep BEACH_GOOGLE_CLOUD_STORAGE_",
	})
	if err != nil {
		return errors.New(fmt.Sprintf("failed connecting to instance with internal host %v - %v", internalHost, err)), "", []byte("")
	}

	var encodedPrivateKey string

	for _, line := range strings.Split(output, "\n") {
		s := strings.SplitN(line, "=", 2)
		if len(s) == 2 {
			switch s[0] {
			case "BEACH_GOOGLE_CLOUD_STORAGE_PUBLIC_BUCKET":
				bucketName = s[1]
			case "BEACH_GOOGLE_CLOUD_STORAGE_SERVICE_ACCOUNT_PRIVATE_KEY":
				encodedPrivateKey = s[1]
			}
		}
	}

	if len(bucketName) == 0 {
		return errors.New("could not determine cloud storage bucket name from instance variables"), "", []byte("")
	}
	log.Debug("Using cloud storage bucket " + bucketName)

	if len(encodedPrivateKey) == 0 {
		return errors.New("could not determine cloud storage private key from instance variables"), "", []byte("")
	}

	privateKey, err = base64.StdEncoding.DecodeString(encodedPrivateKey)
	if err != nil {
		return errors.New("failed decoding cloud storage private key"), "", []byte("")
	}

	log.Info("Retrieved cloud storage private key")
	return nil, bucketName, privateKey
}
