// Copyright 2019-2025 Robert Lemke, Karsten Dambekalns, Christian Müller
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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop instances, reverse proxy and database server and remove containers",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleDownRun,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func handleDownRun(cmd *cobra.Command, args []string) {
	if err := bringBeachDown(); err != nil {
		log.Fatal(err)
	}
}
