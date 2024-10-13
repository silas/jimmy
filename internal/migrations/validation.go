package migrations

import (
	"errors"
)

func (m *Migrations) Validate() error {
	if m == nil || m.migrations == nil {
		return errors.New("must be initialized using New")
	}

	if m.Path == "" {
		return errors.New("config path required")
	}

	if m.Config == nil {
		return errors.New("config required")
	}

	if m.Config.Project == "" {
		return errors.New("config project required")
	}

	if m.Config.InstanceId == "" {
		return errors.New("config instance required")
	}

	if m.Config.DatabaseId == "" {
		return errors.New("config database required")
	}

	if m.Config.Table == "" {
		return errors.New("config table required")
	}

	return nil
}
