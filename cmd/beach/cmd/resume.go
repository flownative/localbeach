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
	"path/filepath"
)

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the temporarily paused reverse proxy and database server",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleResumeRun,
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func handleResumeRun(cmd *cobra.Command, args []string) {
	log.Info("Starting reverse proxy and database server ...")
	commandArgs := []string{"compose", "-f", filepath.Join(path.Base, "docker-compose.yml"), "start", "webserver", "database"}
	output, err := exec.RunCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}
	return
}
