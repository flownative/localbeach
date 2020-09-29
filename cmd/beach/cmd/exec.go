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
	"fmt"
	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:                "exec",
	Short:              "Execute a command in or enter a Local Beach container",
	Long:               "",
	DisableFlagParsing: true,
	Run:                handleExecRun,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func handleExecRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal(err)
		return
	}

	commandArgs := []string{"exec", "-ti", sandbox.ProjectName + "_php"}
	if len(args) > 0 {
		commandArgs = append(commandArgs, "bash", "-c", strings.Trim(fmt.Sprint(args), "[]"))
	} else {
		commandArgs = append(commandArgs, "bash")
	}

	err = exec.RunInteractiveCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}
