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
	"github.com/flownative/localbeach/pkg/beachsandbox"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var instanceIdentifier string

// resourceUploadCmd represents the resource-upload command
var resourceUploadCmd = &cobra.Command{
	Use:   "resource-upload",
	Short: "Upload resources (assets) from a local Flow or Neos installation to Beach",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run:   handleResourceUploadRun,
}

func init() {
	resourceUploadCmd.Flags().StringVar(&instanceIdentifier, "instance", "", "The instance identifier of the Beach instance to upload to.")
	_ = resourceUploadCmd.MarkFlagRequired("instance")
	rootCmd.AddCommand(resourceUploadCmd)
}

func handleResourceUploadRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal("Could not activate sandbox: ", err)
		return
	}

	log.Info(sandbox.ProjectRootPath)

	return
}
