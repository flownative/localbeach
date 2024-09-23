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
	"errors"
	"strings"

	"github.com/flownative/localbeach/pkg/exec"
	"github.com/flownative/localbeach/pkg/path"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

var host string

// setupHttpsCmd represents the setup-https command
var setupHttpsCmd = &cobra.Command{
	Use:   "setup-https",
	Short: "Setup HTTPS for Local Beach projects",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleSetupHttpsRun,
}

func init() {
	setupHttpsCmd.Flags().StringVar(&host, "host", "*.localbeach.net", "Host to use for the certificate. Multiple can be given comma-separated.")
	rootCmd.AddCommand(setupHttpsCmd)
}

func handleSetupHttpsRun(cmd *cobra.Command, args []string) {
	log.Info("Setting up HTTPS for Local Beach.")
	log.Info("You will be asked for your password in order to install the CA certificate")

	commandArgs := []string{"-install"}
	err := exec.RunInteractiveCommand("mkcert", commandArgs)
	if err != nil {
		log.Error(err)
		return
	}

	commandArgs = []string{"-cert-file", path.Certificates + "default.crt", "-key-file", path.Certificates + "default.key"}
	for _, hostname := range strings.Split(host, ",") {
		commandArgs = append(commandArgs, strings.Trim(hostname, " "))
	}
	err = exec.RunInteractiveCommand("mkcert", commandArgs)
	if err != nil {
		log.Error(err)
		return
	}

	statusOutput, err := exec.RunCommand("nerdctl", []string{"ps", "--format", "{{.Names}}"})
	if err != nil {
		log.Error(errors.New("failed checking status of container local_beach_nginx container"))
	}

	if !strings.Contains(statusOutput, "local_beach_nginx") {
		log.Info("Restarting reverse proxy ...")
		commandArgs = []string{"compose", "-f", path.Base + "compose.yaml", "-p", "LocalBeach", "restart"}
		output, err := exec.RunCommand("nerdctl", commandArgs)
		if err != nil {
			log.Fatal(output)
			return
		}
	}

	return
}
