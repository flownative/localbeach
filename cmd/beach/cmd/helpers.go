package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/flownative/localbeach/pkg/path"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"

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
	slashPosition := strings.Index(hash, "/")
	if slashPosition > 0 {
		return string(hash[0:slashPosition]) + "/" + string(hash[slashPosition+1]) + "/" + string(hash[slashPosition+2]) + "/" + string(hash[slashPosition+3])
	} else {
		return string(hash[0]) + "/" + string(hash[1]) + "/" + string(hash[2]) + "/" + string(hash[3])
	}
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
			case "BEACH_GOOGLE_CLOUD_STORAGE_STORAGE_BUCKET", "BEACH_GOOGLE_CLOUD_STORAGE_PUBLIC_BUCKET":
				bucketName = s[1]
			case "BEACH_GOOGLE_CLOUD_STORAGE_SERVICE_ACCOUNT_PRIVATE_KEY":
				encodedPrivateKey = s[1]
			}
		}
	}

	if len(bucketName) == 0 {
		log.Debug("Received the following output while fetching BEACH_GOOGLE_CLOUD_STORAGE_* variables:")
		log.Debug(output)
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

func startLocalBeach() error {
	_, err := os.Stat(path.Base)
	if os.IsNotExist(err) {
		err = setupLocalBeach()
		if err != nil {
			return err
		}
	}

	nginxStatusOutput, err := exec.RunCommand("docker", []string{"ps", "--filter", "name=local_beach_nginx", "--filter", "status=running", "-q"})
	if err != nil {
		return errors.New("failed checking status of container local_beach_nginx container, maybe the Docker daemon is not running")
	}

	databaseStatusOutput, err := exec.RunCommand("docker", []string{"ps", "--filter", "name=local_beach_database", "--filter", "status=running", "-q"})
	if err != nil {
		return errors.New("failed checking status of container local_beach_database container")
	}

	if len(nginxStatusOutput) == 0 || len(databaseStatusOutput) == 0 {
		composeFileContent := readFileFromAssets("local-beach/docker-compose.yml")
		composeFileContent = strings.ReplaceAll(composeFileContent, "{{databasePath}}", path.Database)
		composeFileContent = strings.ReplaceAll(composeFileContent, "{{certificatesPath}}", path.Certificates)

		destination, err := os.Create(filepath.Join(path.Base, "docker-compose.yml"))
		if err != nil {
			log.Error("failed creating docker-compose.yml: ", err)
		} else {
			_, err = destination.WriteString(composeFileContent)
			if err != nil {
				log.Error(err)
			}

		}
		_ = destination.Close()

		log.Info("Starting reverse proxy and database server ...")
		commandArgs := []string{"compose", "-f", path.Base + "docker-compose.yml", "up", "--remove-orphans", "-d"}
		err = exec.RunInteractiveCommand("docker", commandArgs)
		if err != nil {
			return errors.New("container startup failed")
		}

		log.Info("Waiting for database server ...")
		tries := 1
		for {
			output, err := exec.RunCommand("docker", []string{"inspect", "-f", "{{.State.Health.Status}}", "local_beach_database"})
			if err != nil {
				return errors.New("failed to check for database server container health")
			}
			if strings.TrimSpace(output) == "healthy" {
				break
			}
			if tries == 10 {
				return errors.New("timeout waiting for database server to start")
			}
			tries++
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}
