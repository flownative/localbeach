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

// 💡 See https://golang.org/cmd/go/#hdr-Build_constraints for explanation of build constraints

//go:build darwin

package path

import (
	"log"
	"os"
	"path/filepath"
)

var OldBase = ""
var Base = ""
var Certificates = ""
var MariaDBDatabase = ""
var MySQLDatabase = ""

func init() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal("Failed detecting home directory")
		return
	}

	OldBase = filepath.Join(homeDir, "Library", "Application Support", "Flownative", "Local Beach")
	Base = filepath.Join(homeDir, ".LocalBeach")
	Certificates = filepath.Join(Base, "Certificates")
	MariaDBDatabase = filepath.Join(Base, "MariaDB")
	MySQLDatabase = filepath.Join(Base, "MySQL")
}
