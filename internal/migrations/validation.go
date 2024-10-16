package migrations

import (
	"errors"
	"fmt"

	"github.com/silas/jimmy/internal/constants"
)

func (m *Migrations) Validate() error {
	if m == nil || m.migrations == nil || m.Config == nil {
		return errors.New("must be initialized using New")
	}

	if m.Path == "" {
		return fmt.Errorf("%q path required", constants.ConfigFile)
	}

	if m.Config.ProjectId == "" {
		return errors.New("project ID required")
	}

	if m.Config.InstanceId == "" {
		return errors.New("instance ID required")
	}

	if m.Config.DatabaseId == "" {
		return errors.New("database ID required")
	}

	return nil
}
