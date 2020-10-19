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
	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the status of the Local Beach instance containers",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		sandbox, err := beachsandbox.GetActiveSandbox()
		if err != nil {
			log.Fatal(err)
			return
		}

		commandArgs := []string{"-f", sandbox.ProjectRootPath + "/.localbeach.docker-compose.yaml", "ps"}
		err = exec.RunInteractiveCommand("docker-compose", commandArgs)
		if err != nil {
			log.Fatal(err)
			return
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
