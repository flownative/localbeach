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

// 💡 See https://golang.org/cmd/go/#hdr-Build_constraints for explanation of build constraints

// +build darwin

package path

import (
	"log"
	"os"
)

var Base = ""
var Certificates = ""
var  Database = ""

func init() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal("Failed detecting home directory")
		return
	}

	Base = homeDir + "/Library/Application Support/Flownative/Local Beach/"
	Certificates = Base + "Nginx/Certificates/"
	Database = Base + "MariaDB"
}
