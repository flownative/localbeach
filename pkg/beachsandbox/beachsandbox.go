package beachsandbox

import (
	"os"
)

type BeachSandbox struct {
	ProjectName     string ``
	ProjectRootPath string ``
}

func (sandbox *BeachSandbox) Init(rootPath string) error {
	sandbox.ProjectRootPath = rootPath

	if err := loadLocalBeachEnvironment(rootPath); err != nil {
		return err
	}

	sandbox.ProjectName = os.Getenv("BEACH_PROJECT_NAME")

	return nil
}

// GetActiveSandbox returns the sandbox based on the current working dir
func GetActiveSandbox() (*BeachSandbox, error) {
	rootPath, err := detectProjectRootPathFromWorkingDir()
	if err != nil {
		return nil, err
	}

	sandbox := &BeachSandbox{}
	if err := sandbox.Init(rootPath); err != nil {
		return sandbox, err
	}

	return sandbox, nil
}

// GetSandbox returns the sandbox based on the given dir
func GetSandbox(rootPath string) (*BeachSandbox, error) {
	sandbox := &BeachSandbox{}
	if err := sandbox.Init(rootPath); err != nil {
		return sandbox, err
	}

	return sandbox, nil
}
