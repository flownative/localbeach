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
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/flownative/localbeach/pkg/beachsandbox"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"os"
	"path/filepath"
)

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
	resourceDownloadCmd.Flags().StringVar(&projectCluster, "cluster", "", "The cluster name of the Beach instance to download from, eg. 'h9acc4'")
	resourceDownloadCmd.Flags().StringVar(&bucketName, "bucket", "", "name of the bucket to download resources from")
	resourceDownloadCmd.Flags().StringVar(&resourcesPath, "resources-path", "", "custom path where to store the downloaded resources, e.g. 'Data/Persistent/Protected'")
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

	if resourcesPath == "" {
		resourcesPath = sandbox.ProjectDataPersistentResourcesPath
	}

	_, err = os.Stat(resourcesPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("The path %v does not exist", resourcesPath))
		return
	}

	err, bucketNameFromCredentials, privateKeyDecoded := retrieveCloudStorageCredentials(instanceIdentifier, projectNamespace)
	if err != nil {
		log.Fatal(err)
		return
	}

	if bucketName == "" {
		bucketName = bucketNameFromCredentials
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(privateKeyDecoded))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize cloud storage client: %v", err))
		return
	}

	log.Info(fmt.Sprintf("Downloading resources from bucket %v to local directory %v ...", bucketName, resourcesPath))

	bucket := client.Bucket(bucketName)
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
			targetPathAndFilename := filepath.Join(resourcesPath, getRelativePersistentResourcePathByHash(attributes.Name), filepath.Base(attributes.Name))

			err = os.MkdirAll(filepath.Dir(targetPathAndFilename), 0755)
			if err != nil {
				log.Fatal(err)
				return
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
