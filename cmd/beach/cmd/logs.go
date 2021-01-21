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
	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var follow bool
var tail int
var containers bool

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Display logs of the Local Beach instance container",
	Long: `This command allows you to either display the content of log files 
found in Data/Logs/* (default) or show the console output of the
Docker containers (--containers).`,
	Run: func(cmd *cobra.Command, args []string) {
		sandbox, err := beachsandbox.GetActiveSandbox()
		if err != nil {
			log.Fatal("Could not activate sandbox: ", err)
			return
		}

		if containers {
			commandArgs := []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "logs", "--tail=" + strconv.Itoa(tail)}
			if follow {
				commandArgs = append(commandArgs, "-f")
			}
			err = exec.RunInteractiveCommand("docker-compose", commandArgs)
			if err != nil {
				log.Fatal(err)
				return
			}
		} else {
			commandArgs := []string{"exec", "-ti", sandbox.ProjectName + "_php"}
			if follow {
				commandArgs = append(commandArgs, "bash", "-c", "tail -n -" + strconv.Itoa(tail) + " -f /application/Data/Logs/*.log")
			} else {
				commandArgs = append(commandArgs, "bash", "-c", "tail -n -" + strconv.Itoa(tail) + " /application/Data/Logs/*.log")
			}

			err = exec.RunInteractiveCommand("docker", commandArgs)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		return
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVarP(&tail, "tail", "t", 10, "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&containers, "containers", "c", false, "Show log of container console output")
}
