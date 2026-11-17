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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flownative/localbeach/pkg/exec"
	"github.com/flownative/localbeach/pkg/path"
	log "github.com/sirupsen/logrus"
)

const (
	mariaDBContainerName      = "local_beach_mariadb"
	mySQLContainerName        = "local_beach_database"
	mariaDBComposeFileName    = "mariadb-compose.yml"
	mariaDBComposeProject     = "local_beach_mariadb_migration"
	databaseRootPassword      = "password"
	legacyMigrationMarkerName = ".mariadb-to-mysql-migration-complete"
)

var systemDatabases = map[string]bool{
	"information_schema": true,
	"performance_schema": true,
	"mysql":              true,
	"sys":                true,
}

// mariaDBMigrationNeeded tells if there is MariaDB data which has not been
// migrated to MySQL yet. After a successful migration, the MariaDB directory
// is renamed, so its mere existence means that a migration is due.
func mariaDBMigrationNeeded() (bool, error) {
	entries, err := os.ReadDir(path.MariaDBDatabase)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read MariaDB database directory %s: %w", path.MariaDBDatabase, err)
	}
	return !isEmptyDirectory(entries), nil
}

func isEmptyDirectory(entries []os.DirEntry) bool {
	for _, entry := range entries {
		if entry.Name() != ".DS_Store" {
			return false
		}
	}
	return true
}

// migrateMariaDBIfNeeded migrates all databases from MariaDB to MySQL, if
// there is MariaDB data which was not migrated yet. Unless assumeYes is set,
// the user is asked for confirmation first.
func migrateMariaDBIfNeeded(assumeYes bool) error {
	needed, err := mariaDBMigrationNeeded()
	if err != nil {
		return err
	}
	if !needed {
		removeLegacyMigrationMarker()
		return nil
	}

	mySQLDirectoryInUse, err := directoryHasContent(path.MySQLDatabase)
	if err != nil {
		return err
	}

	log.Warn("Local Beach now uses MySQL 8.0 instead of MariaDB 10.11.")
	log.Warn("Your databases in " + path.MariaDBDatabase + " need to be migrated to MySQL once.")
	log.Warn("")
	log.Warn("The migration will:")
	log.Warn("  - stop all running Local Beach instances, the reverse proxy and the database server")
	if mySQLDirectoryInUse {
		log.Warn("  - move the existing directory " + path.MySQLDatabase + " aside (nothing is deleted)")
	}
	log.Warn("  - copy every database from MariaDB to MySQL and compare the row counts of all tables")
	log.Warn("  - keep your MariaDB data and rename its directory to " + migratedMariaDBDirectory())
	log.Warn("")
	log.Warn("Depending on the size of your databases, this can take several minutes.")

	if !assumeYes {
		if !isTTY() {
			return errors.New("the databases need to be migrated from MariaDB to MySQL before Local Beach can start; run \"beach setup\" in a terminal, or \"beach setup --yes\" to migrate without asking")
		}
		if !askForConfirmation("Migrate your databases to MySQL now?") {
			return errors.New("Local Beach cannot start before the databases are migrated to MySQL; run \"beach setup\" when you are ready")
		}
	}

	if err := migrateMariaDBToMySQL(mySQLDirectoryInUse); err != nil {
		log.Error("The migration failed. Your MariaDB data in " + path.MariaDBDatabase + " has not been changed.")
		log.Error("Fix the problem and run \"beach setup\" to try again.")
		stopLocalBeachInfrastructure()
		return err
	}
	return nil
}

func migrateMariaDBToMySQL(mySQLDirectoryInUse bool) error {
	if err := bringBeachDown(); err != nil {
		return fmt.Errorf("failed to stop Local Beach before the migration: %w", err)
	}

	if mySQLDirectoryInUse {
		backupPath := path.MySQLDatabase + ".backup-" + time.Now().Format("20060102-150405")
		log.Info("Moving existing MySQL data aside to " + backupPath)
		if err := os.Rename(path.MySQLDatabase, backupPath); err != nil {
			return fmt.Errorf("failed to move %s aside: %w", path.MySQLDatabase, err)
		}
	}
	if err := os.MkdirAll(path.MySQLDatabase, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create %s: %w", path.MySQLDatabase, err)
	}

	log.Info("Step 1/4: Starting MySQL and MariaDB ...")
	writeLocalBeachComposeFile()
	if err := startLocalBeachInfrastructure(); err != nil {
		return err
	}
	if err := startMariaDB(); err != nil {
		return err
	}
	defer stopMariaDB()

	log.Info("Step 2/4: Discovering databases ...")
	databases, err := listMariaDBDatabases()
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Found %d database(s) to migrate", len(databases)))

	log.Info("Step 3/4: Migrating databases ...")
	for i, database := range databases {
		log.Info(fmt.Sprintf("  [%d/%d] %s", i+1, len(databases), database))
		if err := migrateDatabase(database); err != nil {
			return fmt.Errorf("migration of database %s failed: %w", database, err)
		}
	}

	log.Info("Step 4/4: Comparing row counts ...")
	if err := verifyMigration(databases); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	stopMariaDB()

	if err := os.Rename(path.MariaDBDatabase, migratedMariaDBDirectory()); err != nil {
		return fmt.Errorf("the databases were migrated, but renaming %s to %s failed: %w", path.MariaDBDatabase, migratedMariaDBDirectory(), err)
	}
	removeLegacyMigrationMarker()
	_ = os.Remove(filepath.Join(path.Base, mariaDBComposeFileName))

	log.Info("✓ Migration to MySQL completed successfully!")
	log.Info("")
	log.Info("Your MariaDB data has been kept at " + migratedMariaDBDirectory())
	log.Info("Once you have verified that everything works, you can remove it:")
	log.Info("  rm -rf \"" + migratedMariaDBDirectory() + "\"")
	if mySQLDirectoryInUse {
		log.Info("The same goes for the previous MySQL data, which was moved to " + path.MySQLDatabase + ".backup-*")
	}
	log.Info("")
	log.Info("Instances which were running have been stopped. Start them again with \"beach start\".")

	return nil
}

func migratedMariaDBDirectory() string {
	return path.MariaDBDatabase + ".migrated"
}

func removeLegacyMigrationMarker() {
	// Written by development versions of the migration, superseded by renaming the MariaDB directory
	_ = os.Remove(filepath.Join(path.Base, legacyMigrationMarkerName))
}

func directoryHasContent(directoryPath string) (bool, error) {
	entries, err := os.ReadDir(directoryPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read directory %s: %w", directoryPath, err)
	}
	return !isEmptyDirectory(entries), nil
}

func askForConfirmation(question string) bool {
	fmt.Print(question + " [y/N] ")
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func startMariaDB() error {
	replacements := map[string]string{
		"{{mariadbDatabasePath}}": path.MariaDBDatabase,
	}
	if err := writeComposeFile("local-beach/mariadb-compose.yml", mariaDBComposeFileName, replacements); err != nil {
		return err
	}

	commandArgs := []string{"compose", "-p", mariaDBComposeProject, "-f", filepath.Join(path.Base, mariaDBComposeFileName), "up", "-d"}
	if output, err := exec.RunCommand("docker", commandArgs); err != nil {
		return fmt.Errorf("failed to start MariaDB: %s", output)
	}
	return waitForHealthyContainer(mariaDBContainerName)
}

func stopMariaDB() {
	composeFilePath := filepath.Join(path.Base, mariaDBComposeFileName)
	if _, err := os.Stat(composeFilePath); err != nil {
		return
	}
	commandArgs := []string{"compose", "-p", mariaDBComposeProject, "-f", composeFilePath, "rm", "--force", "--stop", "-v"}
	if output, err := exec.RunCommand("docker", commandArgs); err != nil {
		log.Error("Failed to stop MariaDB: ", output)
	}
}

// runSQL runs a query in the given database container and returns its output
// in batch format, without column names
func runSQL(containerName string, query string) (string, error) {
	client := "mysql"
	if containerName == mariaDBContainerName {
		client = "mariadb"
	}
	command := osexec.Command("docker", "exec", "-i", "-e", "MYSQL_PWD="+databaseRootPassword, containerName, client, "-uroot", "--batch", "--skip-column-names")
	command.Stdin = strings.NewReader(query)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("query in %s failed: %s", containerName, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func listMariaDBDatabases() ([]string, error) {
	output, err := runSQL(mariaDBContainerName, "SHOW DATABASES;")
	if err != nil {
		return nil, err
	}
	var databases []string
	for _, line := range strings.Split(output, "\n") {
		database := strings.TrimSpace(line)
		if database != "" && !systemDatabases[database] {
			databases = append(databases, database)
		}
	}
	return databases, nil
}

// migrateDatabase streams a dump of the database from MariaDB through
// convertMariaDBDump into MySQL
func migrateDatabase(database string) error {
	dump := osexec.Command("docker", "exec", "-e", "MYSQL_PWD="+databaseRootPassword, mariaDBContainerName,
		"mariadb-dump", "-uroot",
		"--single-transaction", "--skip-lock-tables", "--quick", "--hex-blob", "--no-tablespaces",
		"--routines", "--triggers", "--events",
		"--databases", database)
	restore := osexec.Command("docker", "exec", "-i", "-e", "MYSQL_PWD="+databaseRootPassword, mySQLContainerName,
		"mysql", "-uroot", "--default-character-set=utf8mb4")

	var dumpStderr, restoreOutput bytes.Buffer
	dump.Stderr = &dumpStderr
	restore.Stdout = &restoreOutput
	restore.Stderr = &restoreOutput

	dumpOutput, err := dump.StdoutPipe()
	if err != nil {
		return err
	}
	restoreInput, err := restore.StdinPipe()
	if err != nil {
		return err
	}

	if err := restore.Start(); err != nil {
		return err
	}
	if err := dump.Start(); err != nil {
		_ = restoreInput.Close()
		_ = restore.Wait()
		return err
	}

	filterErr := convertMariaDBDump(dumpOutput, restoreInput)
	_ = restoreInput.Close()
	if filterErr != nil {
		_ = dump.Process.Kill()
	}
	dumpErr := dump.Wait()
	restoreErr := restore.Wait()

	switch {
	case restoreErr != nil:
		return fmt.Errorf("import into MySQL failed: %s", strings.TrimSpace(restoreOutput.String()))
	case dumpErr != nil:
		return fmt.Errorf("dump from MariaDB failed: %s", strings.TrimSpace(dumpStderr.String()))
	case filterErr != nil:
		return fmt.Errorf("converting the dump failed: %w", filterErr)
	}
	return nil
}

// verifyMigration compares the tables and their row counts of all migrated
// databases between MariaDB and MySQL
func verifyMigration(databases []string) error {
	if len(databases) == 0 {
		return nil
	}

	quotedDatabases := make([]string, len(databases))
	for i, database := range databases {
		quotedDatabases[i] = "'" + strings.ReplaceAll(database, "'", "''") + "'"
	}
	tables, err := runSQL(mariaDBContainerName, "SELECT table_schema, table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema IN ("+strings.Join(quotedDatabases, ",")+");")
	if err != nil {
		return err
	}

	var countQuery strings.Builder
	tableCount := 0
	for _, line := range strings.Split(strings.TrimSpace(tables), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 2 {
			continue
		}
		schema, table := quoteIdentifier(fields[0]), quoteIdentifier(fields[1])
		label := strings.ReplaceAll(fields[0]+"."+fields[1], "'", "''")
		countQuery.WriteString("SELECT '" + label + "', COUNT(*) FROM " + schema + "." + table + ";\n")
		tableCount++
	}
	if tableCount == 0 {
		return nil
	}

	mariaDBCounts, err := runSQL(mariaDBContainerName, countQuery.String())
	if err != nil {
		return err
	}
	mySQLCounts, err := runSQL(mySQLContainerName, countQuery.String())
	if err != nil {
		return err
	}

	differences := compareCounts(mariaDBCounts, mySQLCounts)
	if len(differences) > 0 {
		return fmt.Errorf("row counts differ for %d table(s): %s", len(differences), strings.Join(differences, ", "))
	}
	log.Info(fmt.Sprintf("All %d table(s) in %d database(s) have the same number of rows in MySQL as in MariaDB", tableCount, len(databases)))
	return nil
}

func quoteIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

// compareCounts compares two outputs of "<table>\t<count>" lines and returns
// a description of every table whose counts differ
func compareCounts(expected string, actual string) []string {
	parse := func(output string) map[string]string {
		counts := map[string]string{}
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			if fields := strings.Split(line, "\t"); len(fields) == 2 {
				counts[fields[0]] = fields[1]
			}
		}
		return counts
	}
	expectedCounts, actualCounts := parse(expected), parse(actual)

	var differences []string
	for table, expectedCount := range expectedCounts {
		actualCount, exists := actualCounts[table]
		if !exists {
			actualCount = "missing"
		}
		if actualCount != expectedCount {
			differences = append(differences, fmt.Sprintf("%s (MariaDB: %s, MySQL: %s)", table, expectedCount, actualCount))
		}
	}
	sort.Strings(differences)
	return differences
}
