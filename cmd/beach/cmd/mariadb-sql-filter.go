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
	"fmt"
	"io"
	"regexp"
	"strings"
)

// The MySQL limit for identifiers, such as the names of CHECK constraints
const mysqlMaxIdentifierLength = 64

var (
	// MariaDB emulates JSON as LONGTEXT with a CHECK constraint:
	//   `data` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '(DC2Type:json)' CHECK (json_valid(`data`)),
	jsonColumnPattern = regexp.MustCompile("^(\\s*)`([^`]+)` longtext(?: CHARACTER SET \\w+)?(?: COLLATE \\w+)?(.*)$")
	jsonCheckPattern  = regexp.MustCompile(" CHECK \\(json_valid\\(`([^`]+)`\\)\\)")

	// A column whose type cannot have a literal default value in MySQL
	blobLikeColumnPattern = regexp.MustCompile("^\\s*`[^`]+` (?:json|tinytext|text|mediumtext|longtext|tinyblob|blob|mediumblob|longblob)\\b")
	// DEFAULT followed by a string literal or a function call, but not by NULL or by an expression in parentheses
	literalDefaultPattern = regexp.MustCompile(`( DEFAULT )('(?:[^'\\]|\\.|'')*'|[A-Za-z_][A-Za-z0-9_]*\(\))`)

	// A CHECK constraint MariaDB named itself, such as CONSTRAINT `CONSTRAINT_1` CHECK (...)
	generatedConstraintPattern = regexp.MustCompile("^(\\s*CONSTRAINT )`CONSTRAINT_(\\d+)`( CHECK .*)$")

	createTablePattern = regexp.MustCompile("^CREATE TABLE `([^`]+)`")
)

// convertMariaDBDump reads an SQL dump created by mariadb-dump and writes a
// version that MySQL 8 can import.
//
// Only column and constraint definitions inside CREATE TABLE statements are
// changed, all other lines (including the data in INSERT statements) are
// passed through unchanged:
//
//   - the "sandbox mode" line of newer mariadb-dump versions is removed,
//     because the mysql client cannot parse it
//   - JSON columns, which MariaDB stores as LONGTEXT with a json_valid() CHECK
//     constraint, become native JSON columns. MySQL would otherwise generate
//     names for these constraints which exceed 64 characters for long table
//     names.
//   - literal default values of TEXT, BLOB and JSON columns are turned into
//     expression defaults, the only kind MySQL accepts for these types
//   - CHECK constraints named CONSTRAINT_n by MariaDB get a name which is
//     unique within the database, as MySQL requires
func convertMariaDBDump(input io.Reader, output io.Writer) error {
	reader := bufio.NewReaderSize(input, 1024*1024)
	writer := bufio.NewWriterSize(output, 1024*1024)

	currentTable := ""
	isFirstLine := true
	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			if isFirstLine && strings.HasPrefix(line, "/*M!999999") {
				line = ""
			}
			isFirstLine = false

			if currentTable != "" {
				if strings.HasPrefix(line, ")") {
					currentTable = ""
				} else {
					line = convertTableDefinitionLine(currentTable, line)
				}
			} else if matches := createTablePattern.FindStringSubmatch(line); matches != nil {
				currentTable = matches[1]
			}

			if _, err := writer.WriteString(line); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return writer.Flush()
}

func convertTableDefinitionLine(tableName string, line string) string {
	lineEnding := ""
	if strings.HasSuffix(line, "\n") {
		lineEnding = "\n"
		line = strings.TrimSuffix(line, "\n")
	}

	line = convertJSONColumn(line)

	if blobLikeColumnPattern.MatchString(line) {
		line = literalDefaultPattern.ReplaceAllString(line, "${1}(${2})")
	}

	if matches := generatedConstraintPattern.FindStringSubmatch(line); matches != nil {
		line = matches[1] + "`" + checkConstraintName(tableName, matches[2]) + "`" + matches[3]
	}

	return line + lineEnding
}

func convertJSONColumn(line string) string {
	checkMatches := jsonCheckPattern.FindStringSubmatch(line)
	if checkMatches == nil {
		return line
	}
	columnMatches := jsonColumnPattern.FindStringSubmatch(line)
	if columnMatches == nil || columnMatches[2] != checkMatches[1] {
		return line
	}

	indentation, columnName, attributes := columnMatches[1], columnMatches[2], columnMatches[3]
	attributes = strings.Replace(attributes, checkMatches[0], "", 1)
	return indentation + "`" + columnName + "` json" + attributes
}

// checkConstraintName returns a name like MySQL would generate it
// ("<table>_chk_<n>"), shortened so that it fits into an identifier
func checkConstraintName(tableName string, number string) string {
	suffix := fmt.Sprintf("_chk_%s", number)
	name := []rune(tableName)
	if len(name)+len(suffix) > mysqlMaxIdentifierLength {
		name = name[:mysqlMaxIdentifierLength-len(suffix)]
	}
	return string(name) + suffix
}
