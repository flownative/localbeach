// Copyright 2019-2022 Robert Lemke, Karsten Dambekalns, Christian Müller
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

var stopRemove bool

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Local Beach instance in the current directory",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleStopRun,
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolVarP(&stopRemove, "remove", "r", false, "Remove containers after they stopped")
}

func handleStopRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal("Could not activate sandbox: ", err)
		return
	}

	commandArgs := []string{"compose", "-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml"}

	if stopRemove {
		commandArgs = append(commandArgs, "down", "--remove-orphans", "--volumes")
	} else {
		commandArgs = append(commandArgs, "stop")
	}

	err = exec.RunInteractiveCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}

	return
}
