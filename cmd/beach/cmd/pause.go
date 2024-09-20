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
	"github.com/flownative/localbeach/pkg/exec"
	"github.com/flownative/localbeach/pkg/path"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// pauseCmd represents the pause command
var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Temporarily stop the reverse proxy and database server",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handlePauseRun,
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

func handlePauseRun(cmd *cobra.Command, args []string) {
	log.Info("Pausing reverse proxy and database server ...")
	commandArgs := []string{"compose", "-f", path.Base + "docker-compose.yml", "stop", "webserver", "database"}
	output, err := exec.RunCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}
	return
}
