package beachsandbox

import (
	log "github.com/sirupsen/logrus"
	"os"
)

type BeachSandbox struct {
	ProjectName string ``
	ProjectRootPath string ``
}

func (sandbox *BeachSandbox) Init(rootPath string) error {
	projectRootPath, err := detectProjectRootPathFromWorkingDir()
	if err != nil {
		return err
	}

	if err = loadLocalBeachEnvironment(projectRootPath); err != nil {
		return err
	}

	sandbox.ProjectName = os.Getenv("BEACH_PROJECT_NAME")
	sandbox.ProjectRootPath = projectRootPath

	log.Debugf("Detected project root path at %s", sandbox.ProjectRootPath)

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
