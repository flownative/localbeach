package beachsandbox

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

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
		return projectRootPath, errors.New("found a Flow or Neos installation but no Local Beach configuration – run \"beach init\" to create some")
	} else if projectRootPath == "/" {
		return "", errors.New("could not find Flow or Neos installation in your current path")
	}

	return detectProjectRootPath(path.Dir(projectRootPath))
}

func loadLocalBeachEnvironment(projectRootPath string) (err error) {
	envFilenames := []string{".env", ".localbeach.env", ".localbeach.dist.env"}
	envPathAndFilename := ""

	for _, envFilename := range envFilenames {
		envPathAndFilename = projectRootPath + "/" + envFilename

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
	}

	return
}
