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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/flownative/localbeach/pkg/beachsandbox"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var instanceIdentifier, projectNamespace, resumeWithFile string
var force bool

// resourceUploadCmd represents the resource-upload command
var resourceUploadCmd = &cobra.Command{
	Use:   "resource-upload",
	Short: "Upload resources (assets) from a local Flow or Neos installation to Beach",
	Long: `resource-upload

This command uploads Flow resources from a local Flow or Neos project to a Beach instance. 

Resource data (that is, the actual files containing binary data, like images or documents)
will be uploaded from the Data/Persistent/Resources directory. It is your responsibility 
to make sure that the database content is matching this data. 

Be aware that Neos and Flow keep track of existing resources by a database table. If 
resources are not registered in there, Flow does not know about them.

Notes:
 - existing data in the Beach instance will be left unchanged
 - older instances may use a namespace called "beach"
`,
	Args: cobra.ExactArgs(0),
	Run:  handleResourceUploadRun,
}

func init() {
	resourceUploadCmd.Flags().StringVar(&instanceIdentifier, "instance", "", "instance identifier of the Beach instance to upload to, eg. 'instance-123abc45-def6-7890-abcd-1234567890ab'")
	resourceUploadCmd.Flags().StringVar(&projectNamespace, "namespace", "", "The project namespace of the Beach instance to upload to, eg. 'project-123abc45-def6-7890-abcd-1234567890ab'")
	resourceUploadCmd.Flags().BoolVar(&force, "force", false, "Force uploading resources which already exist in the target bucket")
	resourceUploadCmd.Flags().StringVar(&resumeWithFile, "resume-with-file", "", "If specified, resume uploading resources starting with the given filename, eg. '12dcde4c13142942288c5a973caf0fa720ed2794'")
	_ = resourceUploadCmd.MarkFlagRequired("instance")
	_ = resourceUploadCmd.MarkFlagRequired("namespace")
	rootCmd.AddCommand(resourceUploadCmd)
}

func handleResourceUploadRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal("Could not activate sandbox: ", err)
		return
	}
	_, err = os.Stat(sandbox.ProjectDataPersistentResourcesPath)
	if err != nil {
		log.Fatal("The path %v does not exist", sandbox.ProjectDataPersistentResourcesPath)
		return
	}

	err, bucketName, privateKeyDecoded := retrieveCloudStorageCredentials(instanceIdentifier, projectNamespace)
	if err != nil {
		log.Fatal(err)
		return
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(privateKeyDecoded))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize cloud storage client: %v", err))
		return
	}

	log.Info(fmt.Sprintf("Uploading resources from local directory %v to bucket %v...", sandbox.ProjectDataPersistentResourcesPath, bucketName))

	var fileList []string
	err = filepath.Walk(sandbox.ProjectDataPersistentResourcesPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed creating list of files to upload: %v", err))
		return
	}

	bucket := client.Bucket(bucketName)
	for _, pathAndFilename := range fileList {
		filename := filepath.Base(pathAndFilename)

		if resumeWithFile != "" && filename < resumeWithFile {
			log.Debug("Skipped  " + filename)
			continue
		}

		_, err = bucket.Object(filename).Attrs(ctx)
		if err == storage.ErrObjectNotExist || force == true {
			source, err := os.Open(pathAndFilename)
			if err != nil {
				log.Fatal(err)
				return
			}
			destination := bucket.Object(filename).NewWriter(ctx)
			if _, err = io.Copy(destination, source); err != nil {
				source.Close()
				log.Fatal(err)
				return
			}
			if err := destination.Close(); err != nil {
				source.Close()
				log.Fatal(err)
				return
			}

			if err = source.Close(); err != nil {
				log.Error(err)
			} else {
				log.Debug("Uploaded " + filename)
			}
		} else if err == nil {
			log.Debug("Skipped  " + filename + " (already exists)")
		} else {
			log.Error(err)
		}
	}

	log.Info("Done")
	return
}
