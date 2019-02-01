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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "beach",
	Short: "Beach and Local Beach support for the command line",
	Long: `
888                             888      
888                             888      
888                             888      
88888b.  .d88b.  8888b.  .d8888b88888b.  
888 "88bd8P  Y8b    "88bd88P"   888 "88b 
888  88888888888.d888888888     888  888 
888 d88PY8b.    888  888Y88b.   888  888 
88888P"  "Y8888 "Y888888 "Y8888P888  888

Beach is the tool for managing projects in Beach and Local Beach.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		DisableLevelTruncation: true,

	})
	log.SetLevel(log.DebugLevel)
}

func initConfig() {
}
