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
	"github.com/flownative/localbeach/pkg/path"
	"path/filepath"
	"strings"

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
		commandArgs := []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "rm", "--force", "--stop", "-v"}
		output, err := exec.RunCommand("docker-compose", commandArgs)
		if err != nil {
			log.Fatal(output)
			return
		}
	}

	log.Info("Stopping reverse proxy and database server ...")
	commandArgs := []string{"-f", path.Base + "docker-compose.yml", "rm", "--force", "--stop", "-v"}
	output, err := exec.RunCommand("docker-compose", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	return
}

func findInstanceRoots() ([]string, error) {
	var configurationFiles []string

	output, err := exec.RunCommand("docker", []string{"ps", "-q", "--filter", "name=_devkit"})
	if err != nil {
		return nil, errors.New(output)
	}
	for _, line := range strings.Split(output, "\n") {
		containerID := strings.TrimSpace(line)
		if len(containerID) > 0 {
			output, err := exec.RunCommand("docker", []string{"inspect", "-f", "{{index .Config.Labels \"com.docker.compose.project.config_files\"}}", containerID})
			if err != nil {
				return nil, errors.New(output)
			}
			configurationFiles = append(configurationFiles, filepath.Dir(strings.TrimSpace(output)))
		}
	}

	return configurationFiles, nil
}
