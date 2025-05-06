// Copyright 2019-2024 Robert Lemke, Karsten Dambekalns, Christian Müller
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package beachsandbox

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var ErrNoLocalBeachConfigurationFound = errors.New("could not find a Local Beach configuration – run \"beach init\" to create some")
var ErrNoFlowFound = errors.New("could not find Flow or Neos installation - try running \"composer install\" to fix that")

func detectProjectRootPathFromWorkingDir() (rootPath string, err error) {
	workingDirPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return detectProjectRootPath(workingDirPath)
}

func detectProjectRootPath(currentPath string) (projectRootPath string, err error) {
	projectRootPath = filepath.Clean(currentPath)
	parentPath := filepath.Dir(projectRootPath)

	if _, err := os.Stat(filepath.Join(projectRootPath, ".localbeach.docker-compose.yaml")); err == nil {
		return projectRootPath, err
	} else if parentPath == projectRootPath {
		// We have reached the root folder
		return "", ErrNoLocalBeachConfigurationFound
	}

	return detectProjectRootPath(parentPath)
}

func loadLocalBeachEnvironment(projectRootPath string) (err error) {
	envFilenames := []string{".localbeach.dist.env", ".localbeach.env", ".env"}
	envPathAndFilename := ""

	for _, envFilename := range envFilenames {
		envPathAndFilename = filepath.Join(projectRootPath, envFilename)

		if _, err := os.Stat(envPathAndFilename); err == nil {

			source, err := os.ReadFile(envPathAndFilename)
			if err != nil {
				return errors.New("failed loading environment file " + envPathAndFilename + ": " + err.Error())
			}

			for _, line := range strings.Split(string(source), "\n") {
				trimmedLine := strings.TrimSpace(line)
				if len(trimmedLine) > 0 && !strings.HasPrefix(trimmedLine, "#") {
					nameAndValue := strings.SplitN(trimmedLine, "=", 2)
					if len(nameAndValue) != 2 {
						return errors.New("failed parsing environment variable " + nameAndValue[0])
					}
					if err := os.Setenv(nameAndValue[0], nameAndValue[1]); err != nil {
						return errors.New("failed setting environment variable " + nameAndValue[0])
					}
				}
			}
		}
	}

	return
}
