// Copyright © 2020 Karsten Dambekalns / Flownative GmbH
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
	commandArgs := []string{"-f", path.Base + "docker-compose.yml", "stop", "webserver", "database"}
	output, err := exec.RunCommand("docker-compose", commandArgs)
	if err != nil {
		log.Fatal(output)
		return
	}
	return
}
