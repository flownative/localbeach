// Copyright © 2019 - 2020 Robert Lemke / Flownative GmbH
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
	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the status of the Local Beach instance containers",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := beachsandbox.GetActiveSandbox()
		if err != nil {
			log.Fatal(err)
			return
		}

		commandArgs := []string{"-f", ".localbeach.docker-compose.yaml", "ps"}
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
