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
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// setupHttpsCmd represents the setup-https command
var setupHttpsCmd = &cobra.Command{
	Use:   "setup-https",
	Short: "Setup HTTPS for Local Beach projects",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleSetupHttpsRun,
}

func init() {
	rootCmd.AddCommand(setupHttpsCmd)
}

func handleSetupHttpsRun(cmd *cobra.Command, args []string) {
	commandArgs := []string{"-install"}
	err := exec.RunInteractiveCommand("mkcert", commandArgs)
	if err != nil {
		log.Error(err)
		return
	}

	commandArgs = []string{"-cert-file", "/usr/local/lib/localbeach/certificates/default.crt", "-key-file", "/usr/local/lib/localbeach/certificates/default.key", "*.localbeach.net"}
	err = exec.RunInteractiveCommand("mkcert", commandArgs)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Restarting reverse proxy ...")
	commandArgs = []string{"-f", "/usr/local/lib/localbeach/docker-compose.yml", "restart"}
	output, err := exec.RunCommand("docker-compose", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}

	return
}
