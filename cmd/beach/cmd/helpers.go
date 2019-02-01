package cmd

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
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
		} else {
			return "", errors.New("found a Flow or Neos installation but no Local Beach configuration – run \"beach init\" to create some")
		}
	} else if projectRootPath == "/" {
		return "", errors.New("could not find Flow or Neos installation in your current path")
	}

	return detectProjectRootPath(path.Dir(projectRootPath))
}
