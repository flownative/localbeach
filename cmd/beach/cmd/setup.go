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
	"github.com/flownative/localbeach/pkg/exec"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var dockerFolder string
var databaseFolder string
var nginxFolder string

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Local Beach on this host",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleSetupRun,
}

func init() {
	setupCmd.Flags().StringVar(&dockerFolder, "docker-folder", "", "Defines the folder used for docker metadata.")
	setupCmd.Flags().StringVar(&nginxFolder, "nginx-folder", "", "Defines the folder used for Nginx configuration.")
	setupCmd.Flags().StringVar(&databaseFolder, "database-folder", "", "Defines the folder used for the database server.")
	rootCmd.AddCommand(setupCmd)
}

func handleSetupRun(cmd *cobra.Command, args []string) {
	if len(databaseFolder) == 0 {
		log.Fatal("database-folder must be given.")
		return
	}
	if len(dockerFolder) == 0 {
		log.Fatal("docker-folder must be given.")
		return
	}
	if len(nginxFolder) == 0 {
		log.Fatal("nginx-folder must be given.")
		return
	}

	err := os.MkdirAll(databaseFolder, os.ModePerm)
	if err != nil {
		log.Error(err)
	}
	err = os.MkdirAll(dockerFolder, os.ModePerm)
	if err != nil {
		log.Error(err)
	}
	err = os.MkdirAll(nginxFolder, os.ModePerm)
	if err != nil {
		log.Error(err)
	}
	err = os.MkdirAll(nginxFolder + "/Certificates", os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	composeFileContent := readFileFromAssets("local-beach/docker-compose.yml")
	composeFileContent = strings.ReplaceAll(composeFileContent, "{{databaseFolder}}", databaseFolder)
	composeFileContent = strings.ReplaceAll(composeFileContent, "{{nginxFolder}}", nginxFolder)

	destination, err := os.Create(filepath.Join(dockerFolder, "docker-compose.yml"))
	if err != nil {
		log.Error(err)
	}
	defer destination.Close()

	_, err = destination.WriteString(composeFileContent)
	if err != nil {
		log.Error(err)
	}

	commandArgs := []string{"-install"}
	err = exec.RunInteractiveCommand("mkcert", commandArgs)
	if err != nil {
		log.Error(err)
		return
	}

	commandArgs = []string{"-cert-file", nginxFolder + "/Certificates/default.crt", "-key-file", nginxFolder + "/Certificates/default.key", "*.localbeach.net"}
	err = exec.RunInteractiveCommand("mkcert", commandArgs)
	if err != nil {
		log.Error(err)
		return
	}


	return
}
