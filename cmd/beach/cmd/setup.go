// Copyright 2019-2020 Robert Lemke, Karsten Dambekalns, Christian Müller
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
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var dockerFolder string
var databaseFolder string

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
	err := os.MkdirAll(databaseFolder, os.ModePerm)
	if err != nil {
		log.Error(err)
	}
	err = os.MkdirAll(dockerFolder, os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	err = os.MkdirAll("/usr/local/lib/localbeach/certificates", os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	composeFileContent := readFileFromAssets("local-beach/docker-compose.yml")
	composeFileContent = strings.ReplaceAll(composeFileContent, "{{databaseFolder}}", databaseFolder)

	destination, err := os.Create(filepath.Join(dockerFolder, "docker-compose.yml"))
	defer destination.Close()
	if err != nil {
		log.Error("failed creating docker-compose.yml: ", err)
	} else {
		_, err = destination.WriteString(composeFileContent)
		if err != nil {
			log.Error(err)
		}

	}
	return
}
