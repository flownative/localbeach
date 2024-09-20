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
	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var restartPull bool
var restartRemove bool

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart the Local Beach instance in the current directory",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleRestartRun,
}

func init() {
	rootCmd.AddCommand(restartCmd)
	restartCmd.Flags().BoolVarP(&restartRemove, "remove", "r", true, "Remove containers before restart")
	restartCmd.Flags().BoolVarP(&restartPull, "pull", "p", false, "Pull images before restart")
}

func handleRestartRun(cmd *cobra.Command, args []string) {

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

	commandArgs := []string{"compose", "-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml"}
	if restartRemove {
		log.Debug("Stopping and removing containers ...")
		commandArgs = append(commandArgs, "down", "--remove-orphans", "--volumes")
	} else {
		log.Debug("Stopping containers ...")
		commandArgs = append(commandArgs, "stop")
	}

	err = exec.RunInteractiveCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}

	if restartPull {
		log.Debug("Pulling images ...")
		commandArgs = []string{"compose", "-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "pull"}
		output, err := exec.RunCommand("docker", commandArgs)
		if err != nil {
			log.Fatal(output)
			return
		}
	}

	log.Debug("Starting containers ...")

	commandArgs = []string{"compose", "-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "up", "--remove-orphans", "-d"}
	output, err := exec.RunCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	log.Info("Local Beach instance was restarted.")
	log.Info("When files have been synced, you can access this instance at http://" + sandbox.ProjectName + ".localbeach.net")
	return
}
