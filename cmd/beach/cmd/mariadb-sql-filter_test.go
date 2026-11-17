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
	"bytes"
	"strings"
	"testing"
)

func convert(t *testing.T, input string) string {
	t.Helper()
	var output bytes.Buffer
	if err := convertMariaDBDump(strings.NewReader(input), &output); err != nil {
		t.Fatalf("convertMariaDBDump failed: %v", err)
	}
	return output.String()
}

func TestConvertMariaDBDumpRemovesSandboxModeLine(t *testing.T) {
	input := "/*M!999999\\- enable the sandbox mode */ \n-- MariaDB dump 10.19\n"
	expected := "-- MariaDB dump 10.19\n"
	if actual := convert(t, input); actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestConvertMariaDBDumpConvertsJSONColumns(t *testing.T) {
	input := "CREATE TABLE `flownative_beach_projection_instancedeployedconfiguration_v1` (\n" +
		"  `id` varchar(40) NOT NULL,\n" +
		"  `recordValues` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '(DC2Type:json)' CHECK (json_valid(`recordValues`)),\n" +
		"  `analysis` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '(DC2Type:json)' CHECK (json_valid(`analysis`)),\n" +
		"  `plain` longtext CHECK (json_valid(`plain`)),\n" +
		"  PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;\n"
	expected := "CREATE TABLE `flownative_beach_projection_instancedeployedconfiguration_v1` (\n" +
		"  `id` varchar(40) NOT NULL,\n" +
		"  `recordValues` json NOT NULL COMMENT '(DC2Type:json)',\n" +
		"  `analysis` json DEFAULT NULL COMMENT '(DC2Type:json)',\n" +
		"  `plain` json,\n" +
		"  PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;\n"
	if actual := convert(t, input); actual != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, actual)
	}
}

func TestConvertMariaDBDumpKeepsLongtextWithCheckOnOtherColumn(t *testing.T) {
	input := "CREATE TABLE `t` (\n" +
		"  `a` longtext NOT NULL CHECK (json_valid(`b`)),\n" +
		") ENGINE=InnoDB;\n"
	if actual := convert(t, input); actual != input {
		t.Errorf("expected line to be unchanged, got:\n%s", actual)
	}
}

func TestConvertMariaDBDumpTurnsLiteralDefaultsOfBlobLikeColumnsIntoExpressions(t *testing.T) {
	input := "CREATE TABLE `t` (\n" +
		"  `options` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '[]' COMMENT '(DC2Type:json)' CHECK (json_valid(`options`)),\n" +
		"  `settings` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT json_object() CHECK (json_valid(`settings`)),\n" +
		"  `description` text NOT NULL DEFAULT 'it''s \\'quoted\\', DEFAULT x',\n" +
		"  `data` blob DEFAULT NULL,\n" +
		"  `computed` text DEFAULT (concat('a','b')),\n" +
		"  `name` varchar(255) NOT NULL DEFAULT 'text',\n" +
		"  `created` datetime NOT NULL DEFAULT current_timestamp(),\n" +
		") ENGINE=InnoDB;\n"
	expected := "CREATE TABLE `t` (\n" +
		"  `options` json NOT NULL DEFAULT ('[]') COMMENT '(DC2Type:json)',\n" +
		"  `settings` json NOT NULL DEFAULT (json_object()),\n" +
		"  `description` text NOT NULL DEFAULT ('it''s \\'quoted\\', DEFAULT x'),\n" +
		"  `data` blob DEFAULT NULL,\n" +
		"  `computed` text DEFAULT (concat('a','b')),\n" +
		"  `name` varchar(255) NOT NULL DEFAULT 'text',\n" +
		"  `created` datetime NOT NULL DEFAULT current_timestamp(),\n" +
		") ENGINE=InnoDB;\n"
	if actual := convert(t, input); actual != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, actual)
	}
}

func TestConvertMariaDBDumpRenamesGeneratedCheckConstraints(t *testing.T) {
	input := "CREATE TABLE `flownative_beach_projection_instancedeployedconfiguration_v1` (\n" +
		"  `amount` int(11) NOT NULL,\n" +
		"  CONSTRAINT `CONSTRAINT_1` CHECK (`amount` >= 0)\n" +
		") ENGINE=InnoDB;\n" +
		"CREATE TABLE `short` (\n" +
		"  CONSTRAINT `CONSTRAINT_1` CHECK (`amount` >= 0),\n" +
		"  CONSTRAINT `my_own_name` CHECK (`amount` < 10)\n" +
		") ENGINE=InnoDB;\n"
	expected := "CREATE TABLE `flownative_beach_projection_instancedeployedconfiguration_v1` (\n" +
		"  `amount` int(11) NOT NULL,\n" +
		"  CONSTRAINT `flownative_beach_projection_instancedeployedconfiguration__chk_1` CHECK (`amount` >= 0)\n" +
		") ENGINE=InnoDB;\n" +
		"CREATE TABLE `short` (\n" +
		"  CONSTRAINT `short_chk_1` CHECK (`amount` >= 0),\n" +
		"  CONSTRAINT `my_own_name` CHECK (`amount` < 10)\n" +
		") ENGINE=InnoDB;\n"
	actual := convert(t, input)
	if actual != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, actual)
	}
	if name := checkConstraintName("flownative_beach_projection_instancedeployedconfiguration_v1", "1"); len(name) != mysqlMaxIdentifierLength {
		t.Errorf("expected shortened name with %d characters, got %d (%s)", mysqlMaxIdentifierLength, len(name), name)
	}
}

func TestConvertMariaDBDumpLeavesDataUnchanged(t *testing.T) {
	input := "INSERT INTO `t` VALUES ('x longtext CHECK (json_valid(`x`))', ' DEFAULT \\'[]\\',');\n" +
		"  `looks_like_a_column` text DEFAULT 'x',\n" +
		"CONSTRAINT `CONSTRAINT_1` CHECK (1)\n" +
		"-- no line ending at the end"
	if actual := convert(t, input); actual != input {
		t.Errorf("expected data outside of CREATE TABLE to be unchanged, got:\n%s", actual)
	}
}

func TestConvertMariaDBDumpHandlesVeryLongLines(t *testing.T) {
	input := "INSERT INTO `t` VALUES ('" + strings.Repeat("x", 5*1024*1024) + "');\n"
	if actual := convert(t, input); actual != input {
		t.Errorf("expected long line to be passed through unchanged (got %d bytes, expected %d)", len(actual), len(input))
	}
}
