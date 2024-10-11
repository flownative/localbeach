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
	"errors"
	"github.com/flownative/localbeach/pkg/exec"
	"github.com/flownative/localbeach/pkg/path"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	err = migrateMariaDBToMySQL()
	if err != nil {
		return err
	}

	writeLocalBeachComposeFile()

	return nil
}

func migrateMariaDBToMySQL() error {
	_, err := os.Stat(path.MariaDBDatabase)
	if err == nil {
		log.Info("Migrating MariaDB data from " + path.MariaDBDatabase + " to MySQL at " + path.MySQLDatabase)
		log.Warn("Note: This may take a while, depending on DB size!")

		if err = startMariaDB(); err != nil {
			return err
		}

		log.Debug("dumping data from MariaDB to MySQL")
		commandArgs := []string{"exec", "local_beach_mariadb", "bash", "-c"}
		commandArgs = append(commandArgs, "mysql -h local_beach_mariadb -u root -ppassword --batch --skip-column-names -e \"SHOW DATABASES;\" | grep -E -v \"(information|performance)_schema|mysql|sys\"")
		databases, err := exec.RunCommand("docker", commandArgs)
		if err != nil {
			log.Error(err)
			return err
		}

		for _, database := range strings.Split(strings.TrimSuffix(databases, "\n"), "\n") {
			log.Debug("… " + database)
			commandArgs = []string{"exec", "local_beach_database", "bash", "-c"}
			commandArgs = append(commandArgs, "mysqldump -h local_beach_mariadb -u root -ppassword --add-drop-trigger --compress --comments --dump-date --hex-blob --quote-names --routines --triggers --no-autocommit --no-tablespaces --skip-lock-tables --single-transaction --quick --databases "+database+" | sed -e \"s/DEFAULT '{}' COMMENT '(DC2Type:json)'/DEFAULT (JSON_OBJECT()) COMMENT '(DC2Type:json)'/\" | mysql -h local_beach_database -u root -ppassword")
			_, err := exec.RunCommand("docker", commandArgs)
			if err != nil {
				log.Error(err)
			}
		}

		if err = stopMariaDB(); err != nil {
			return err
		}
	}

	log.Info("Done with migration to MySQL at " + path.MySQLDatabase)
	log.Info("If all works as expected, remove " + path.MariaDBDatabase)

	return nil
}

func startMariaDB() error {
	log.Debug("starting MariaDB server ...")

	writeMariaDBComposeFile()

	commandArgs := []string{"compose", "-f", filepath.Join(path.Base, "mariadb-compose.yml"), "up", "-d"}
	err := exec.RunInteractiveCommand("docker", commandArgs)
	if err != nil {
		return errors.New("Database container startup failed")
	}

	log.Debug("waiting for MariaDB server ...")
	tries := 1
	for {
		output, err := exec.RunCommand("docker", []string{"inspect", "-f", "{{.State.Health.Status}}", "local_beach_mariadb"})
		if err != nil {
			return errors.New("failed to check for MariaDB server container health")
		}
		if strings.TrimSpace(output) == "healthy" {
			break
		}
		if tries == 10 {
			return errors.New("timeout waiting for MariaDB server to start")
		}
		tries++
		time.Sleep(3 * time.Second)
	}

	log.Debug("waiting for MySQL server ...")
	tries = 1
	for {
		output, err := exec.RunCommand("docker", []string{"inspect", "-f", "{{.State.Health.Status}}", "local_beach_database"})
		if err != nil {
			return errors.New("failed to check for MySQL server container health")
		}
		if strings.TrimSpace(output) == "healthy" {
			break
		}
		if tries == 10 {
			return errors.New("timeout waiting for MySQL server to start")
		}
		tries++
		time.Sleep(3 * time.Second)
	}

	return nil
}
func stopMariaDB() error {
	log.Debug("stopping MariaDB server ...")
	commandArgs := []string{"compose", "-f", filepath.Join(path.Base, "mariadb-compose.yml"), "rm", "--force", "--stop", "-v"}
	_, err := exec.RunCommand("docker", commandArgs)

	return err
}
