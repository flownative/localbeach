// Copyright © 2020 Karsten Dambekalns / Flownative GmbH
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
	"errors"
	"path/filepath"
	"strings"

	"github.com/flownative/localbeach/pkg/path"

	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop instances, reverse proxy and database server and remove containers",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleDownRun,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func handleDownRun(cmd *cobra.Command, args []string) {
	instanceRoots, err := findInstanceRoots()
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, instanceRoot := range instanceRoots {
		log.Info("Stopping instance in " + instanceRoot + "...")
		sandbox, err := beachsandbox.GetSandbox(instanceRoot)
		if err != nil {
			log.Fatal(err)
			return
		}
		commandArgs := []string{"compose", "-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "down", "-v"}
		output, err := exec.RunCommand("nerdctl", commandArgs)
		if err != nil {
			log.Fatal(output)
			return
		}
	}

	log.Info("Stopping reverse proxy and database server ...")
	commandArgs := []string{"compose", "-f", path.Base + "compose.yaml", "down", "-v", "-p", "LocalBeach"}
	output, err := exec.RunCommand("nerdctl", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	return
}

func findInstanceRoots() ([]string, error) {
	var configurationFiles []string
	var containerData []string
	var containerID string
	var containerName string

	output, err := exec.RunCommand("nerdctl", []string{"ps", "--format", "{{.ID}} {{.Names}}"})
	if err != nil {
		return nil, errors.New(output)
	}
	for _, line := range strings.Split(output, "\n") {
		containerData = strings.Split(line, " ")
		containerID = strings.TrimSpace(containerData[0])
		containerName = ""
		if len(containerData) > 1 {
			containerName = strings.TrimSpace(containerData[1])
		}

		if len(containerID) > 0 && strings.Contains(containerName, "_devkit") {
			output, err := exec.RunCommand("nerdctl", []string{"inspect", "-f", "{{index .Config.Labels \"com.docker.compose.project.config_files\"}}", containerID})
			if err != nil {
				return nil, errors.New(output)
			}
			configurationFiles = append(configurationFiles, filepath.Dir(strings.TrimSpace(output)))
		}
	}

	return configurationFiles, nil
}
