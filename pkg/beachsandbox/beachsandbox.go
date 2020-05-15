package beachsandbox

import (
	"os"
)

type BeachSandbox struct {
	ProjectName     string ``
	ProjectRootPath string ``
}

func (sandbox *BeachSandbox) Init(rootPath string) error {
	projectRootPath, err := detectProjectRootPathFromWorkingDir()
	sandbox.ProjectRootPath = projectRootPath

	if err != nil {
		return err
	}

	if err = loadLocalBeachEnvironment(projectRootPath); err != nil {
		return err
	}

	sandbox.ProjectName = os.Getenv("BEACH_PROJECT_NAME")

	return nil
}

// GetActiveSandbox returns the sandbox based on the current working dir
func GetActiveSandbox() (*BeachSandbox, error) {
	sandbox := &BeachSandbox{}
	if err := sandbox.Init(""); err != nil {
		return sandbox, err
	}

	return sandbox, nil
}
