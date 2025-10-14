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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	databaseStatusOutput, err := exec.RunCommand("docker", []string{"ps", "--filter", "name=local_beach_database", "--filter", "status=running", "-q"})
	if err != nil {
		log.Error(errors.New("failed checking status of container local_beach_database container"))
	}

	if len(databaseStatusOutput) != 0 {
		if err := bringBeachDown(); err != nil {
			return fmt.Errorf("failed to bring Beach down: %w", err)
		}
	}

	err = os.MkdirAll(path.Base, os.ModePerm)
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
		log.Error(err)
		return err
	}

	writeLocalBeachComposeFile()

	return nil
}

func migrateMariaDBToMySQL() error {
	migrationMarkerPath := filepath.Join(path.Base, ".mariadb-to-mysql-migration-complete")

	// Check if migration has already been completed
	if _, err := os.Stat(migrationMarkerPath); err == nil {
		log.Debug("MariaDB to MySQL migration already completed, skipping")
		return nil
	}

	// Check if MariaDB data exists
	_, err := os.Stat(path.MariaDBDatabase)
	if err != nil {
		// No MariaDB data to migrate
		if os.IsNotExist(err) {
			log.Debug("No MariaDB data found, skipping migration")
			return nil
		}
		return fmt.Errorf("failed to check MariaDB database path: %w", err)
	}

	log.Info("Migrating MariaDB data from " + path.MariaDBDatabase + " to MySQL at " + path.MySQLDatabase)
	log.Warn("This process may take several minutes depending on database size.")
	log.Warn("Please do not interrupt this process!")

	// Start both database servers
	if err = startMariaDB(); err != nil {
		return fmt.Errorf("failed to start database servers for migration: %w", err)
	}

	// Ensure cleanup happens on error
	defer func() {
		if stopErr := stopMariaDB(); stopErr != nil {
			log.Error("Failed to stop MariaDB after migration: ", stopErr)
		}
	}()

	// Get list of databases to migrate
	log.Info("Step 1/3: Discovering databases to migrate...")
	commandArgs := []string{"exec", "local_beach_mariadb", "bash", "-c"}
	commandArgs = append(commandArgs, "mysql -h local_beach_mariadb -u root -ppassword --batch --skip-column-names -e \"SHOW DATABASES;\" | grep -E -v \"(information|performance)_schema|mysql|sys\"")
	databases, err := exec.RunCommand("docker", commandArgs)
	if err != nil {
		return fmt.Errorf("failed to list databases from MariaDB: %w", err)
	}

	databaseList := strings.Split(strings.TrimSuffix(databases, "\n"), "\n")
	if len(databaseList) == 0 || (len(databaseList) == 1 && databaseList[0] == "") {
		log.Info("No databases found to migrate")
	} else {
		log.Info(fmt.Sprintf("Found %d database(s) to migrate", len(databaseList)))

		// Migrate each database
		log.Info("Step 2/3: Migrating databases...")
		migratedCount := 0
		for i, database := range databaseList {
			if database == "" {
				continue
			}
			log.Info(fmt.Sprintf("  [%d/%d] Migrating database: %s", i+1, len(databaseList), database))
			commandArgs = []string{"exec", "local_beach_database", "bash", "-c"}
			commandArgs = append(commandArgs, "mysqldump -h local_beach_mariadb -u root -ppassword --add-drop-trigger --compress --comments --dump-date --hex-blob --quote-names --routines --triggers --no-autocommit --no-tablespaces --skip-lock-tables --single-transaction --quick --databases "+database+" | sed -e \"s/DEFAULT '{}' COMMENT '(DC2Type:json)'/DEFAULT (JSON_OBJECT()) COMMENT '(DC2Type:json)'/\" | mysql -h local_beach_database -u root -ppassword")
			output, err := exec.RunCommand("docker", commandArgs)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to migrate database %s: %v", database, err))
				if output != "" {
					log.Error("Output: ", output)
				}
				return fmt.Errorf("migration failed for database %s: %w", database, err)
			}
			migratedCount++
			log.Info(fmt.Sprintf("  [%d/%d] Successfully migrated: %s", i+1, len(databaseList), database))
		}
		log.Info(fmt.Sprintf("Successfully migrated %d database(s)", migratedCount))
	}

	// Verify migration
	log.Info("Step 3/3: Verifying migration...")
	if err = verifyMigration(); err != nil {
		return fmt.Errorf("migration verification failed: %w", err)
	}

	// Stop MariaDB
	if err = stopMariaDB(); err != nil {
		return fmt.Errorf("failed to stop MariaDB after migration: %w", err)
	}

	// Create migration marker file
	markerFile, err := os.Create(migrationMarkerPath)
	if err != nil {
		log.Warn("Failed to create migration marker file: ", err)
		log.Warn("Migration completed successfully, but may run again on next setup")
	} else {
		timestamp := time.Now().Format(time.RFC3339)
		_, _ = markerFile.WriteString(fmt.Sprintf("Migration completed at: %s\n", timestamp))
		markerFile.Close()
	}

	log.Info("✓ Migration to MySQL completed successfully!")
	log.Info("")
	log.Info("Your MariaDB data has been preserved at: " + path.MariaDBDatabase)
	log.Info("Once you've verified everything works correctly, you can safely remove it with:")
	log.Info("  rm -rf " + path.MariaDBDatabase)

	return nil
}

func verifyMigration() error {
	// Check that MySQL server is running and accessible
	commandArgs := []string{"exec", "local_beach_database", "bash", "-c"}
	commandArgs = append(commandArgs, "mysql -h local_beach_database -u root -ppassword --batch --skip-column-names -e \"SELECT 'OK';\"")
	output, err := exec.RunCommand("docker", commandArgs)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	if !strings.Contains(output, "OK") {
		return errors.New("MySQL connection test failed")
	}

	// Get database count from MySQL
	commandArgs = []string{"exec", "local_beach_database", "bash", "-c"}
	commandArgs = append(commandArgs, "mysql -h local_beach_database -u root -ppassword --batch --skip-column-names -e \"SHOW DATABASES;\" | grep -E -v \"(information|performance)_schema|mysql|sys\" | wc -l")
	mysqlDbCount, err := exec.RunCommand("docker", commandArgs)
	if err != nil {
		return fmt.Errorf("failed to count MySQL databases: %w", err)
	}

	mysqlDbCount = strings.TrimSpace(mysqlDbCount)
	log.Info(fmt.Sprintf("Verification: Found %s database(s) in MySQL", mysqlDbCount))

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
