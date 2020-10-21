// Copyright © 2020 Christian Müller / Flownative GmbH
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/flownative/localbeach/pkg/beachsandbox"
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
	sandbox, err := beachsandbox.GetRawSandbox()
	if err != nil {
		log.Fatal(err)
		return
	}

	projectNameFilter := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	projectName := strings.Trim(projectName, " ")
	if len(projectName) == 0 {
		projectName = path.Base(sandbox.ProjectRootPath)
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
	defer destination.Close()
	destination.WriteString(environmentContent)
	log.Info("Created '.localbeach.dist.env'.")

	log.Info("You are all set")
	return
}
