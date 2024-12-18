// Copyright 2019-2024 Robert Lemke, Karsten Dambekalns, Christian Müller
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
	"context"
	"errors"
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

var targetBucketName, sourceResourcesPath, resumeWithFile string
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

The Google Cloud Storage bucket name will be determined automatically through the environment
variables set in the given instance. You can override the bucket name by specifying the --bucket
parameter.

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
	resourceUploadCmd.Flags().StringVar(&projectNamespace, "namespace", "", "The project namespace of the Beach instance to upload to, eg. 'beach-project-123abc45-def6-7890-abcd-1234567890ab'")
	resourceUploadCmd.Flags().StringVar(&clusterIdentifier, "cluster", "", "The cluster identifier of the Beach instance to upload to, eg. 'h9acc4'")
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
	if sourceResourcesPath == "" {
		sourceResourcesPath = sandbox.ProjectDataPersistentResourcesPath
	}
	_, err = os.Stat(sourceResourcesPath)
	if err != nil {
		log.Fatal("The path %v does not exist", sourceResourcesPath)
		return
	}

	err, bucketNameFromCredentials, privateKeyDecoded := retrieveCloudStorageCredentials(instanceIdentifier, projectNamespace, clusterIdentifier)
	if err != nil {
		log.Fatal(err)
		return
	}

	if targetBucketName == "" {
		targetBucketName = bucketNameFromCredentials
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(privateKeyDecoded))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize cloud storage client: %v", err))
		return
	}

	log.Info(fmt.Sprintf("Uploading resources from local directory %v to bucket %v...", sourceResourcesPath, targetBucketName))

	var fileList []string
	err = filepath.Walk(sourceResourcesPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed creating list of files to upload: %v", err))
		return
	}

	bucket := client.Bucket(targetBucketName)
	for _, pathAndFilename := range fileList {
		filename := filepath.Base(pathAndFilename)

		if resumeWithFile != "" && filename < resumeWithFile {
			log.Debug("Skipped  " + filename)
			continue
		}

		_, err = bucket.Object(filename).Attrs(ctx)
		if errors.Is(err, storage.ErrObjectNotExist) || force == true {
			source, err := os.Open(pathAndFilename)
			if err != nil {
				log.Fatal(err)
				return
			}
			destination := bucket.Object(filename).NewWriter(ctx)
			if _, err = io.Copy(destination, source); err != nil {
				_ = source.Close()
				log.Fatal(err)
				return
			}
			if err := destination.Close(); err != nil {
				_ = source.Close()
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
