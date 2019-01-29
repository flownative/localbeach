package cmd

import (
	"errors"
	"os"
	"path"
	"strings"
)

func findFlowRootPathStartingFrom(currentPath string) (rootPath string, err error) {
	projectBasePath := getNormalizedPath(currentPath)
	flowCommand := projectBasePath + "flow"

	if _, err := os.Stat(flowCommand); err == nil {
		return projectBasePath, nil
	}
	if projectBasePath == "/" {
		return "", errors.New("could not find Flow or Neos installation in your current path")
	}

	rootPath, err = findFlowRootPathStartingFrom(path.Dir(currentPath))
	return rootPath, err
}

func getLocalBeachDockerComposePathAndFilename(projectPath string) (path string) {
	return getNormalizedPath(projectPath) + ".localbeach.docker-compose.yaml"
}

func getNormalizedPath(path string) string {
	return strings.TrimRight(path, "/") + "/"
}

