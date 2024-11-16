package migrations

import (
	"iter"
	"maps"
	"slices"

	"google.golang.org/protobuf/proto"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

const (
	builtinDefaultTemplateID = "create-table"
)

var (
	builtinTemplates = map[string]*jimmyv1.Template{
		builtinDefaultTemplateID:     {Sql: constants.CreateTable, Type: jimmyv1.Type_DDL},
		"drop-table":                 {Sql: constants.DropTable, Type: jimmyv1.Type_DDL},
		"add-column":                 {Sql: constants.AddColumn, Type: jimmyv1.Type_DDL},
		"drop-column":                {Sql: constants.DropColumn, Type: jimmyv1.Type_DDL},
		"set-default":                {Sql: constants.SetDefault, Type: jimmyv1.Type_DDL},
		"drop-default":               {Sql: constants.DropDefault, Type: jimmyv1.Type_DDL},
		"add-check-constraint":       {Sql: constants.AddCheckConstraint, Type: jimmyv1.Type_DDL},
		"add-foreign-key-constraint": {Sql: constants.AddForeignKeyConstraint, Type: jimmyv1.Type_DDL},
		"drop-constraint":            {Sql: constants.DropConstraint, Type: jimmyv1.Type_DDL},
		"create-index":               {Sql: constants.CreateIndex, Type: jimmyv1.Type_DDL},
		"drop-index":                 {Sql: constants.DropIndex, Type: jimmyv1.Type_DDL},
	}
)

func BuiltinTemplates() iter.Seq2[string, *jimmyv1.Template] {
	return templateSeq(maps.Clone(builtinTemplates))
}

func (ms *Migrations) Templates() iter.Seq2[string, *jimmyv1.Template] {
	templates := maps.Clone(builtinTemplates)

	for templateID, template := range ms.Config.GetTemplates() {
		templates[templateID] = template
	}

	return templateSeq(templates)
}

func templateSeq(templates map[string]*jimmyv1.Template) iter.Seq2[string, *jimmyv1.Template] {
	var defaultTemplate *jimmyv1.Template

	for _, template := range templates {
		if template.GetDefault() {
			defaultTemplate = template
			break
		}
	}

	if defaultTemplate == nil {
		defaultTemplate = proto.Clone(templates[builtinDefaultTemplateID]).(*jimmyv1.Template)

		v := true
		defaultTemplate.Default = &v

		templates[builtinDefaultTemplateID] = defaultTemplate
	}

	keys := slices.Sorted(maps.Keys(templates))

	return func(yield func(string, *jimmyv1.Template) bool) {
		for _, key := range keys {
			if !yield(key, templates[key]) {
				return
			}
		}
	}
}
