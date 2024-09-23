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

package beachsandbox

import (
	"errors"
	"os"
	"path/filepath"
)

type BeachSandbox struct {
	ProjectName                        string ``
	ProjectRootPath                    string ``
	ProjectDataPersistentResourcesPath string ``
	DockerComposeFilePath              string ``
	FlowRootPath                       string ``
}

func (sandbox *BeachSandbox) Init(rootPath string) error {
	sandbox.ProjectRootPath = rootPath

	if err := loadLocalBeachEnvironment(rootPath); err != nil {
		return err
	}

	sandbox.DockerComposeFilePath = filepath.Join(sandbox.ProjectRootPath, ".localbeach.docker-compose.yaml")
	sandbox.ProjectName = os.Getenv("BEACH_PROJECT_NAME")
	sandbox.FlowRootPath = os.Getenv("BEACH_FLOW_ROOTPATH")
	sandbox.ProjectDataPersistentResourcesPath = filepath.Join(rootPath, sandbox.FlowRootPath, "/Data/Persistent/Resources")

	return nil
}

// GetActiveSandbox returns the active sandbox based on the current working dir
func GetActiveSandbox() (*BeachSandbox, error) {
	rootPath, err := detectProjectRootPathFromWorkingDir()
	if err != nil {
		return nil, err
	}

	return GetSandbox(rootPath)
}

// GetRawSandbox returns the (unconfigured) sandbox based on the current working dir
func GetRawSandbox() (*BeachSandbox, error) {
	rootPath, err := detectProjectRootPathFromWorkingDir()
	if errors.Is(err, ErrNoFlowFound) {
		return nil, err
	}

	return GetSandbox(rootPath)
}

// GetSandbox returns the sandbox based on the given dir
func GetSandbox(rootPath string) (*BeachSandbox, error) {
	sandbox := &BeachSandbox{}
	if err := sandbox.Init(rootPath); err != nil {
		return sandbox, err
	}

	return sandbox, nil
}
