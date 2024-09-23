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
	"os"
	"path/filepath"
	"strings"

	"github.com/flownative/localbeach/pkg/exec"
	"github.com/flownative/localbeach/pkg/path"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Local Beach on this computer",
	Long:  "This command is usually run automatically during installation (for example by the Homebrew setup scripts).",
	Args:  cobra.ExactArgs(0),
	Run:   handleSetupRun,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func handleSetupRun(cmd *cobra.Command, args []string) {
	_ = setupLocalBeach()
}

func setupLocalBeach() error {
	log.Debug("setting up Local Beach with base path " + path.Base)

	err := os.MkdirAll(path.Base, os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	_, err = os.Stat(path.OldBase)
	if err == nil {
		log.Info("migrating old data from " + path.OldBase + " to " + path.Base)

		log.Info("stopping reverse proxy and database server")
		commandArgs := []string{"compose", "-f", path.OldBase + "docker-compose.yml", "down", "-v"}
		output, err := exec.RunCommand("nerdctl", commandArgs)
		if err != nil {
			log.Error(output)
		}

		log.Info("moving certificates")
		err = os.Rename(path.OldBase+"Nginx/Certificates", path.Certificates)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(err)
			} else {
				log.Fatal(err)
				return err
			}
		}

		log.Info("moving database data")
		err = os.Rename(path.OldBase+"MariaDB", path.Database)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(err)
			} else {
				log.Fatal(err)
				return err
			}
		}

		err = os.RemoveAll(path.OldBase + "Nginx")
		if err != nil {
			log.Error(err)
		}

		err = os.Remove(path.OldBase + "docker-compose.yml")
		if err != nil {
			log.Error(err)
		}

		err = os.RemoveAll(path.OldBase)
		if err != nil {
			log.Error(err)
		}
	}

	log.Debug("creating directory for certificates at " + path.Certificates)
	err = os.MkdirAll(path.Certificates, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Error(err)
	}

	log.Debug("creating directory for databases at " + path.Database)
	err = os.MkdirAll(path.Database, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Error(err)
	}

	composeFileContent := readFileFromAssets("local-beach/compose.yaml")
	composeFileContent = strings.ReplaceAll(composeFileContent, "{{databasePath}}", path.Database)
	composeFileContent = strings.ReplaceAll(composeFileContent, "{{certificatesPath}}", path.Certificates)

	destination, err := os.Create(filepath.Join(path.Base, "compose.yaml"))
	if err != nil {
		log.Error("failed creating compose.yaml: ", err)
	} else {
		_, err = destination.WriteString(composeFileContent)
		if err != nil {
			log.Error(err)
		}

	}
	_ = destination.Close()

	return nil
}
