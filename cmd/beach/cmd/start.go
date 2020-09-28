// Copyright © 2019 - 2020 Robert Lemke / Flownative GmbH
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
	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var pull bool

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
	startCmd.Flags().BoolVarP(&pull, "pull", "p", true, "Pull images before start")
}

func handleStartRun(cmd *cobra.Command, args []string) {
	commandArgs := []string{""}

	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = startLocalBeach()
	if err != nil {
		log.Fatal(err)
		return
	}

	if pull {
		log.Debug("Pulling images ...")
		commandArgs = []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "pull"}
		_, err := exec.RunCommand("docker-compose", commandArgs)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		log.Info("Skipping image pull")
	}

	commandArgs = []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "up", "--remove-orphans", "-d"}
	err = exec.RunInteractiveCommand("docker-compose", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}

	commandArgs = []string{"exec", "local_beach_database", "/bin/bash", "-c", "echo 'CREATE DATABASE IF NOT EXISTS `" + sandbox.ProjectName + "`' | mysql -u root --password=password"}
	_, err = exec.RunCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Info("You are all set")
	log.Info("When files have been synced, you can access this instance at http://" + sandbox.ProjectName + ".localbeach.net")
	return
}

func startLocalBeach() error {
	output, err := exec.RunCommand("docker-compose", []string{"-f", "/usr/local/lib/beach-cli/localbeach/docker-compose.yml", "ps", "-q"})
	if err != nil {
		return err
	}
	if len(output) == 0 {
		log.Info("Starting nginx & database ...")
		commandArgs := []string{"-f", "/usr/local/lib/beach-cli/localbeach/docker-compose.yml", "up", "--remove-orphans", "-d"}
		err = exec.RunInteractiveCommand("docker-compose", commandArgs)
		if err != nil {
			return err
		}

		log.Info("Waiting for database ...")
		tries := 1
		for {
			output, err := exec.RunCommand("docker", []string{"inspect", "-f", "{{.State.Health.Status}}", "local_beach_database"})
			if err != nil {
				return err
			}
			if strings.TrimSpace(output) == "healthy" {
				break
			}
			if tries == 10 {
				return errors.New("timeout waiting for database to start")
			}
			tries++
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}
