package migrations

import (
	"fmt"
	"strings"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func (ms *Migrations) newStatement(
	sql string,
	env jimmyv1.Environment,
	templateID string,
	statementType jimmyv1.Type,
) (*jimmyv1.Statement, error) {
	stmt := &jimmyv1.Statement{
		Sql:  sql,
		Env:  env,
		Type: statementType,
	}

	if stmt.Sql == "" {
		var template *jimmyv1.Template

		for id, tmpl := range ms.Templates() {
			if templateID == id || (templateID == "" && tmpl.GetDefault()) {
				template = tmpl
				break
			}
		}

		if template == nil {
			return nil, fmt.Errorf("%q template not found", templateID)
		}

		stmt.Sql = template.Sql
		stmt.Env = template.Env
		stmt.Type = template.Type
	}

	if stmt.Type == jimmyv1.Type_AUTOMATIC {
		stmt.Type = detectType(stmt.Sql)
	}

	stmt.Sql = strings.TrimSpace(stmt.Sql) + "\n"

	return stmt, nil
}
