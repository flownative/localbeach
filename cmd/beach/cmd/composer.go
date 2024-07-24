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

// composerCmd represents the composer command
var composerCmd = &cobra.Command{
	Use:                "composer",
	Short:              "Run a Composer command within a Local Beach project",
	Long:               "",
	Example:			"beach composer update --verbose",
	DisableFlagParsing: true,
	Run:                handleComposerRun,
}

func init() {
	rootCmd.AddCommand(composerCmd)
}

func handleComposerRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal(err)
		return
	}

	commandArgs := []string{"exec", "-ti", sandbox.ProjectName + "_php"}
	if len(args) > 0 {
		commandArgs = append(commandArgs, "bash", "-c", strings.Trim(fmt.Sprint(args), "[]"))
	} else {
		commandArgs = append(commandArgs, "bash", "-c", "/application/composer")
	}

	err = exec.RunInteractiveCommand("docker", commandArgs)
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}
