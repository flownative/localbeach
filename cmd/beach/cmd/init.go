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

package cmd

import (
	"os"
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectName string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize a Local Beach instance in the current directory",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleInitRun,
}

func init() {
	initCmd.Flags().StringVar(&projectName, "project-name", "", "Defines the project name, defaults to folder name.")
	rootCmd.AddCommand(initCmd)
}

func handleInitRun(cmd *cobra.Command, args []string) {
	var err error

	projectNameFilter := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	projectName := strings.Trim(projectName, " ")
	if len(projectName) == 0 {
		workingDirPath, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return
		}
		projectName = path.Base(workingDirPath)
	}

	projectName = projectNameFilter.ReplaceAllLiteralString(projectName, "")

	if len(projectName) == 0 {
		log.Fatal("The project-name is empty, but cannot be.")
		return
	}

	log.Info("Project name set as " + projectName)

	_, err = copyFileFromAssets("project/.localbeach.docker-compose.yaml", ".localbeach.docker-compose.yaml")
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Info("Created '.localbeach.docker-compose.yaml'.")

	_, err = copyFileFromAssets("project/Settings.yaml", "Configuration/Development/Beach/Settings.yaml")
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Info("Created 'Configuration/Development/Beach/Settings.yaml'.")

	environmentContent := readFileFromAssets("project/.localbeach.dist.env")
	environmentContent = strings.ReplaceAll(environmentContent, "${BEACH_PROJECT_NAME}", projectName)
	environmentContent = strings.ReplaceAll(environmentContent, "${BEACH_PROJECT_NAME_LOWERCASE}", strings.ToLower(projectName))

	destination, err := os.Create(".localbeach.dist.env")
	if err != nil {
		log.Error(err)
	}
	defer func(destination *os.File) {
		_ = destination.Close()
	}(destination)
	destination.WriteString(environmentContent)
	log.Info("Created '.localbeach.dist.env'.")

	log.Info("You are all set")
	return
}
