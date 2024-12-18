package migrations

import (
	"errors"
	"fmt"

	"github.com/silas/jimmy/internal/constants"
)

func (ms *Migrations) Validate() error {
	if ms == nil || ms.migrations == nil || ms.Config == nil {
		return errors.New("must be initialized using New")
	}

	if ms.Path == "" {
		return fmt.Errorf("%q path required", constants.ConfigFile)
	}

	if ms.Config.ProjectId == "" {
		return errors.New("project ID required")
	}

	if ms.Config.InstanceId == "" {
		return errors.New("instance ID required")
	}

	if ms.Config.DatabaseId == "" {
		return errors.New("database ID required")
	}

	var defaultTemplateID string
	for templateID, template := range ms.Config.Templates {
		if !template.GetDefault() {
			continue
		}

		if defaultTemplateID != "" {
			return fmt.Errorf(
				"%q and %q can't both be marked as the default template",
				defaultTemplateID,
				templateID,
			)
		}

		defaultTemplateID = templateID
	}

	return nil
}
