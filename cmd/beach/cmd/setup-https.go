// Copyright © 2020 Robert Lemke / Flownative GmbH
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
