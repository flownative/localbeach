// Copyright 2019-2022 Robert Lemke, Karsten Dambekalns, Christian Müller
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
	"os"
)

type BeachSandbox struct {
	ProjectName                        string ``
	ProjectRootPath                    string ``
	ProjectDataPersistentResourcesPath string ``
}

func (sandbox *BeachSandbox) Init(rootPath string) error {
	sandbox.ProjectRootPath = rootPath
	sandbox.ProjectDataPersistentResourcesPath = rootPath + "/Data/Persistent/Resources"

	if err := loadLocalBeachEnvironment(rootPath); err != nil {
		return err
	}

	sandbox.ProjectName = os.Getenv("BEACH_PROJECT_NAME")

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
	if err == ErrNoFlowFound {
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
