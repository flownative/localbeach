package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	asset "github.com/flownative/localbeach/assets"
	log "github.com/sirupsen/logrus"
)

func detectProjectRootPathFromWorkingDir() (rootPath string, err error) {
	workingDirPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}

	return detectProjectRootPath(workingDirPath)
}

func detectProjectRootPath(currentPath string) (projectRootPath string, err error) {
	projectRootPath = path.Clean(currentPath)

	if _, err := os.Stat(projectRootPath + "/flow"); err == nil {
		if _, err := os.Stat(projectRootPath + "/.localbeach.docker-compose.yaml"); err == nil {
			return projectRootPath, err
		}
		return "", errors.New("found a Flow or Neos installation but no Local Beach configuration – run \"beach init\" to create some")
	}

	if projectRootPath == "/" {
		return "", errors.New("could not find Flow or Neos installation in your current path")
	}

	return detectProjectRootPath(path.Dir(projectRootPath))
}

func loadLocalBeachEnvironment(projectRootPath string) (err error) {
	envPathAndFilename := projectRootPath + "/.localbeach.dist.env"
	if _, err := os.Stat(envPathAndFilename); err == nil {

		source, err := ioutil.ReadFile(envPathAndFilename)
		if err != nil {
			return errors.New("failed loading environment file " + envPathAndFilename + ": " + err.Error())
		}

		for _, line := range strings.Split(string(source), "\n") {
			trimmedLine := strings.TrimSpace(line)
			if len(trimmedLine) > 0 && !strings.HasPrefix(trimmedLine, "#") {
				nameAndValue := strings.Split(trimmedLine, "=")
				if err := os.Setenv(nameAndValue[0], nameAndValue[1]); err != nil {
					return errors.New("failed setting environment variable " + nameAndValue[0])
				}
			}
		}
	}
	return
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

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
