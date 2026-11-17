// Copyright 2019-2025 Robert Lemke, Karsten Dambekalns, Christian Müller
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

var setupAssumeYes bool

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVarP(&setupAssumeYes, "yes", "y", false, "Migrate databases from MariaDB to MySQL without asking")
}

func handleSetupRun(cmd *cobra.Command, args []string) {
	if err := setupLocalBeach(); err != nil {
		log.Fatal(err)
	}
	if err := migrateMariaDBIfNeeded(setupAssumeYes); err != nil {
		log.Fatal(err)
	}
}

func migrateOldBase() error {
	_, err := os.Stat(path.OldBase)
	if err == nil {
		log.Info("migrating old data from " + path.OldBase + " to " + path.Base)

		log.Info("stopping reverse proxy and database server")
		commandArgs := []string{"compose", "-f", filepath.Join(path.OldBase, "docker-compose.yml"), "rm", "--force", "--stop", "-v"}
		output, err := exec.RunCommand("docker", commandArgs)
		if err != nil {
			log.Error(output)
		}

		log.Info("moving certificates")
		err = os.Rename(filepath.Join(path.OldBase, "Nginx", "Certificates"), path.Certificates)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(err)
			} else {
				log.Fatal(err)
				return err
			}
		}

		log.Info("moving database data")
		err = os.Rename(filepath.Join(path.OldBase, "MariaDB"), path.MariaDBDatabase)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(err)
			} else {
				log.Fatal(err)
				return err
			}
		}

		err = os.RemoveAll(filepath.Join(path.OldBase, "Nginx"))
		if err != nil {
			log.Error(err)
		}

		err = os.Remove(filepath.Join(path.OldBase, "docker-compose.yml"))
		if err != nil {
			log.Error(err)
		}

		err = os.RemoveAll(path.OldBase)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}

func setupLocalBeach() error {
	log.Debug("setting up Local Beach with base path " + path.Base)

	err := os.MkdirAll(path.Base, os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	err = migrateOldBase()
	if err != nil {
		return err
	}

	log.Debug("creating directory for certificates at " + path.Certificates)
	err = os.MkdirAll(path.Certificates, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Error(err)
	}

	log.Debug("creating directory for databases at " + path.MySQLDatabase)
	err = os.MkdirAll(path.MySQLDatabase, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Error(err)
	}

	writeLocalBeachComposeFile()

	return nil
}
