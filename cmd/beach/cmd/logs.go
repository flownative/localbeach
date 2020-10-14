// Copyright © 2019 - 2020 Robert Lemke / Flownative GmbH
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
			log.Fatal(err)
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
