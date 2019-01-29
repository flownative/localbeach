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
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Local Beach instance in the current directory.",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("pull", "p", true, "Pull Docker images before start")
}

func run() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flowRootPath, err := findFlowRootPathStartingFrom(workingDirectory)
	if err != nil {
		panic(err)
	}

	localBeachFolderPath := getLocalBeachDockerComposePathAndFilename(flowRootPath)
	if _, err := os.Stat(localBeachFolderPath); os.IsNotExist(err) {
		fmt.Println("We found a Flow or Neos installation but no Local Beach configuration, please run 'beach local:init' to get the initial configuration.")
	}

	fmt.Printf("The Flow root path is: %s", localBeachFolderPath)
}