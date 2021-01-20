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
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/flownative/localbeach/pkg/beachsandbox"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"os"
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

Be aware that Neos and Flow keep track of existing resources by a database table. If 
resources are not registered in there, Flow does not know about them.

Notes:
 - existing data in the local Neos instance will be left unchanged
 - older Beach instances may use a namespace called "beach"
`,
	Args:  cobra.ExactArgs(0),
	Run:   handleResourceDownloadRun,
}

func init() {
	resourceDownloadCmd.Flags().StringVar(&instanceIdentifier, "instance", "", "instance identifier of the Beach instance to download from, eg. 'instance-123abc45-def6-7890-abcd-1234567890ab'")
	resourceDownloadCmd.Flags().StringVar(&projectNamespace, "namespace", "", "The project namespace of the Beach instance to download from, eg. 'project-123abc45-def6-7890-abcd-1234567890ab'")
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

	log.Info(fmt.Sprintf("Downloading resources from bucket %v to local directory %v ...", bucketName, sandbox.ProjectDataPersistentResourcesPath))

	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)
	for {
		attributes, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Error(err)
		} else {
			source := bucket.Object(attributes.Name)
			targetPath := sandbox.ProjectDataPersistentResourcesPath + "/" + getRelativePersistentResourcePathByHash(attributes.Name)

			err = os.MkdirAll(targetPath, 0755)
			if err != nil {
				log.Fatal(err)
				return
			}

			file, err := os.OpenFile(targetPath+"/"+attributes.Name, os.O_RDWR|os.O_CREATE, 0644)
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
