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
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/flownative/localbeach/pkg/beachsandbox"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var sourceBucketName, targetResourcesPath string
var synchronize bool

// resourceDownloadCmd represents the resource-download command
var resourceDownloadCmd = &cobra.Command{
	Use:   "resource-download",
	Short: "Download resources (assets) from a local Flow or Neos installation to Beach",
	Long: `resource-download

This command downloads Flow resources from a Beach instance to a local Flow or Neos project. 

Resource data (that is, the actual files containing binary data, like images or documents)
will be downloaded to the Data/Persistent/Resources directory. It is your responsibility 
to make sure that the database content is matching this data. 

The Google Cloud Storage bucket name will be determined automatically through the environment
variables set in the given instance. You can override the bucket name by specifying the --bucket
parameter.

Be aware that Neos and Flow keep track of existing resources by a database table. If 
resources are not registered in there, Flow does not know about them.

Notes:
 - existing data in the local Neos instance will be left unchanged
 - older Beach instances may use a namespace called "beach"
`,
	Args: cobra.ExactArgs(0),
	Run:  handleResourceDownloadRun,
}

func init() {
	resourceDownloadCmd.Flags().StringVar(&instanceIdentifier, "instance", "", "instance identifier of the Beach instance to download from, eg. 'instance-123abc45-def6-7890-abcd-1234567890ab'")
	resourceDownloadCmd.Flags().StringVar(&projectNamespace, "namespace", "", "The project namespace of the Beach instance to download from, eg. 'beach-project-123abc45-def6-7890-abcd-1234567890ab'")
	resourceDownloadCmd.Flags().StringVar(&clusterIdentifier, "cluster", "", "The cluster identifier of the Beach instance to download from, eg. 'h9acc4'")
	resourceDownloadCmd.Flags().StringVar(&sourceBucketName, "bucket", "", "name of the bucket to download resources from")
	resourceDownloadCmd.Flags().StringVar(&targetResourcesPath, "resources-path", "", "custom path where to store the downloaded resources, e.g. 'Data/Persistent/Protected'")
	resourceDownloadCmd.Flags().BoolVar(&synchronize, "sync", false, "Skip unchanged existing files")

	_ = resourceDownloadCmd.MarkFlagRequired("instance")
	_ = resourceDownloadCmd.MarkFlagRequired("namespace")
	rootCmd.AddCommand(resourceDownloadCmd)
}

func handleResourceDownloadRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal("Could not activate sandbox: ", err)
		return
	}

	if targetResourcesPath == "" {
		targetResourcesPath = sandbox.ProjectDataPersistentResourcesPath
	}

	_, err = os.Stat(targetResourcesPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("The path %v does not exist", targetResourcesPath))
		return
	}

	err, bucketNameFromCredentials, privateKeyDecoded := retrieveCloudStorageCredentials(instanceIdentifier, projectNamespace, clusterIdentifier)
	if err != nil {
		log.Fatal(err)
		return
	}

	if sourceBucketName == "" {
		sourceBucketName = bucketNameFromCredentials
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(privateKeyDecoded))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize cloud storage client: %v", err))
		return
	}

	log.Info(fmt.Sprintf("Downloading resources from bucket %v to local directory %v ...", sourceBucketName, targetResourcesPath))

	bucket := client.Bucket(sourceBucketName)
	it := bucket.Objects(ctx, nil)
	for {
		attributes, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Error(err)
		} else {
			source := bucket.Object(attributes.Name)
			targetPathAndFilename := filepath.Join(targetResourcesPath, getRelativePersistentResourcePathByHash(attributes.Name), filepath.Base(attributes.Name))

			err = os.MkdirAll(filepath.Dir(targetPathAndFilename), 0755)
			if err != nil {
				log.Fatal(err)
				return
			}

			if synchronize == true {
				if checkFileExists(targetPathAndFilename, attributes) {
					log.Debug("Skipped " + attributes.Name + " as it already exists")
					continue
				}
			}

			file, err := os.OpenFile(targetPathAndFilename, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Fatal(err)
				return
			}
			reader, err := source.NewReader(ctx)
			if err != nil {
				log.Fatal(err)
				return
			}
			if _, err := io.Copy(file, reader); err != nil {
				log.Fatal(err)
				return
			}
			if err := reader.Close(); err != nil {
				log.Fatal(err)
				return
			}
			log.Debug("Downloaded " + attributes.Name)
		}
	}

	log.Info("Done")
	return
}

func checkFileExists(targetPathAndFilename string, attributes *storage.ObjectAttrs) bool {
	if _, err := os.Stat(targetPathAndFilename); err == nil {
		file, err := os.Open(targetPathAndFilename)
		if err != nil {
			return false
		}
		defer file.Close()

		crc32c := crc32.New(crc32.MakeTable(crc32.Castagnoli))
		if _, err := io.Copy(crc32c, file); err != nil {
			return false
		}

		if crc32c.Sum32() == attributes.CRC32C {
			return true
		}
	}
	return false
}
