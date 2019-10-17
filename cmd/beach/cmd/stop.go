// Copyright © 2019 Robert Lemke / Flownative GmbH
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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/flownative/localbeach/pkg/beachsandbox"
	"gitlab.com/flownative/localbeach/pkg/exec"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Local Beach instance in the current directory",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := beachsandbox.GetActiveSandbox()
		if err != nil {
			log.Fatal(err)
			return
		}

		commandArgs := []string{"-f", ".localbeach.docker-compose.yaml", "stop"}
		err = exec.RunInteractiveCommand("docker-compose", commandArgs)
		if err != nil {
			log.Fatal(err)
			return
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
