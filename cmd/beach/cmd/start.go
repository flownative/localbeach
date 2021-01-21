// Copyright 2019-2021 Robert Lemke, Karsten Dambekalns, Christian Müller
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
	"github.com/flownative/localbeach/pkg/path"
	"os"
	"strings"
	"time"

	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startPull bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Local Beach instance in the current directory",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleStartRun,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&startPull , "pull", "p", false, "Pull images before start")
}

func handleStartRun(cmd *cobra.Command, args []string) {
	commandArgs := []string{""}

	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal("Could not activate sandbox: ", err)
		return
	}

	err = startLocalBeach()
	if err != nil {
		log.Fatal(err)
		return
	}

	if startPull  {
		log.Debug("Pulling images ...")
		commandArgs = []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "pull"}
		output, err := exec.RunCommand("docker-compose", commandArgs)
		if err != nil {
			log.Fatal(output)
			return
		}
	}

	commandArgs = []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "up", "--remove-orphans", "-d"}
	output, err := exec.RunCommand("docker-compose", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	commandArgs = []string{"exec", "local_beach_database", "/bin/bash", "-c", "echo 'CREATE DATABASE IF NOT EXISTS `" + sandbox.ProjectName + "`' | mysql -u root --password=password"}
	output, err = exec.RunCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	log.Info("You are all set")
	log.Info("When files have been synced, you can access this instance at http://" + sandbox.ProjectName + ".localbeach.net")
	return
}

func startLocalBeach() error {
	_, err := os.Stat(path.Base)
	if os.IsNotExist(err) {
		err = setupLocalBeach()
		if err != nil {
			return err
		}
	}

	nginxStatusOutput, err := exec.RunCommand("docker", []string{"ps", "--filter", "name=local_beach_nginx", "--filter", "status=running", "-q"})
	if err != nil {
		return errors.New("failed checking status of container local_beach_nginx container, maybe the Docker daemon is not running")
	}

	databaseStatusOutput, err := exec.RunCommand("docker", []string{"ps", "--filter", "name=local_beach_database", "--filter", "status=running", "-q"})
	if err != nil {
		return errors.New("failed checking status of container local_beach_database container")
	}

	if len(nginxStatusOutput) == 0 || len(databaseStatusOutput) == 0 {
		log.Info("Starting reverse proxy and database server ...")
		commandArgs := []string{"-f", path.Base + "docker-compose.yml", "up", "--remove-orphans", "-d"}
		err = exec.RunInteractiveCommand("docker-compose", commandArgs)
		if err != nil {
			return errors.New("container startup failed")
		}

		log.Info("Waiting for database server ...")
		tries := 1
		for {
			output, err := exec.RunCommand("docker", []string{"inspect", "-f", "{{.State.Health.Status}}", "local_beach_database"})
			if err != nil {
				return errors.New("failed to check for database server container health")
			}
			if strings.TrimSpace(output) == "healthy" {
				break
			}
			if tries == 10 {
				return errors.New("timeout waiting for database server to start")
			}
			tries++
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}
