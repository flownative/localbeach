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
)

var pull bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Local Beach instance in the current directory",
	Long:  "",
	Args: cobra.ExactArgs(0),
	Run: handleStartRun,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&pull, "pull", "p", true, "Pull images before start")
}

func handleStartRun(cmd *cobra.Command, args []string) {
	commandArgs := []string{""}

	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal(err)
		return
	}

	if pull {
		log.Debug("Pulling images ...")
		commandArgs = []string{"-f", ".localbeach.docker-compose.yaml", "pull"}
		_, err := exec.RunCommand("docker-compose", commandArgs)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		log.Info("Skipping image pull")
	}

	commandArgs = []string{"-f", ".localbeach.docker-compose.yaml", "up", "--remove-orphans", "-d"}
	err = exec.RunInteractiveCommand("docker-compose", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}

	commandArgs = []string{"exec", "local_beach_database" ,"/bin/bash" ,"-c", "echo 'CREATE DATABASE IF NOT EXISTS `" + sandbox.ProjectName + "`' | mysql -u root --password=password"}
	_, err = exec.RunCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Info("You are all set")
	log.Info("When files have been synced, you can access this instance at http://" + sandbox.ProjectName + ".localbeach.net")
	return
}
