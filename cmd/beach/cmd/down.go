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
	"errors"
	"os"
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
		commandArgs := []string{"compose", "-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "rm", "--force", "--stop", "-v"}
		output, err := exec.RunCommand("nerdctl", commandArgs)
		if err != nil {
			log.Fatal(output)
			return
		}
	}

	log.Info("Stopping reverse proxy and database server ...")
	commandArgs := []string{"compose", "-f", path.Base + "compose.yaml", "rm", "--force", "--stop", "-v"}
	output, err := exec.RunCommand("nerdctl", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	return
}

func findInstanceRoots() ([]string, error) {
	var configurationFiles []string

	output, err := exec.RunCommand("nerdctl", []string{"ps", "-q", "--filter", "network=local_beach"})
	if err != nil {
		return nil, errors.New(output)
	}
	for _, line := range strings.Split(output, "\n") {
		containerID := strings.TrimSpace(line)
		if len(containerID) > 0 {
			output, err := exec.RunCommand("nerdctl", []string{"inspect", "-f", "{{index .Config.Labels \"com.docker.compose.project.config_files\"}}", containerID})
			if err != nil {
				return nil, errors.New(output)
			}
			projectDirectory := filepath.Dir(strings.TrimSpace(output))
			if containsLocalBeachInstance(projectDirectory) {
				configurationFiles = append(configurationFiles, projectDirectory)
			}
		}
	}

	return removeDuplicates(configurationFiles), nil
}

func containsLocalBeachInstance(path string) bool {
	path = path + "/.localbeach.docker-compose.yaml"
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func removeDuplicates(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
