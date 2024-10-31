package migrations

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"buf.build/go/protoyaml"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

var (
	slugifyChars         = regexp.MustCompile(`[^a-z0-9_]`)
	slugifyMultiDashes   = regexp.MustCompile(`_+`)
	dmlPartitionedPrefix = regexp.MustCompile(`^(?i)(DELETE|UPDATE)`)
	dmlPrefix            = regexp.MustCompile(`^(?i)INSERT`)
)

func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugifyChars.ReplaceAllString(s, "_")
	s = slugifyMultiDashes.ReplaceAllString(s, "_")
	s = strings.Trim(s, "-_")
	return s
}

func Marshal(path string, m proto.Message) error {
	b, err := protoyaml.MarshalOptions{
		Indent: 2,
	}.Marshal(m)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0644)
}

func Unmarshal(path string, m proto.Message) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	validator, err := protovalidate.New()
	if err != nil {
		return err
	}

	return protoyaml.UnmarshalOptions{
		Path:      path,
		Validator: validator,
	}.Unmarshal(b, m)
}

func detectType(sql string) jimmyv1.Type {
	sql = strings.TrimSpace(sql)

	if dmlPrefix.MatchString(sql) {
		return jimmyv1.Type_DML
	} else if dmlPartitionedPrefix.MatchString(sql) {
		return jimmyv1.Type_PARTITIONED_DML
	}

	return jimmyv1.Type_DDL
}

func templateSQL(template jimmyv1.Template) (string, error) {
	var sql string

	switch template {
	case jimmyv1.Template_CREATE_TABLE, jimmyv1.Template_TEMPLATE_UNSPECIFIED:
		sql = constants.CreateTable
	case jimmyv1.Template_DROP_TABLE:
		sql = constants.DropTable
	case jimmyv1.Template_ADD_COLUMN:
		sql = constants.AddColumn
	case jimmyv1.Template_DROP_COLUMN:
		sql = constants.DropColumn
	case jimmyv1.Template_SET_DEFAULT:
		sql = constants.SetDefault
	case jimmyv1.Template_DROP_DEFAULT:
		sql = constants.DropDefault
	case jimmyv1.Template_CREATE_INDEX:
		sql = constants.CreateIndex
	case jimmyv1.Template_DROP_INDEX:
		sql = constants.DropIndex
	case jimmyv1.Template_ADD_CHECK_CONSTRAINT:
		sql = constants.AddCheckConstraint
	case jimmyv1.Template_ADD_FOREIGN_KEY_CONSTRAINT:
		sql = constants.AddForeignKeyConstraint
	case jimmyv1.Template_DROP_CONSTRAINT:
		sql = constants.DropConstraint
	default:
		return "", fmt.Errorf("%q template unknown", template.String())
	}

	return sql, nil
}

func newStatement(
	sql string,
	env jimmyv1.Environment,
	template jimmyv1.Template,
	statementType jimmyv1.Type,
) (*jimmyv1.Statement, error) {
	var err error

	if sql == "" && statementType == jimmyv1.Type_AUTOMATIC {
		statementType = jimmyv1.Type_DDL
	}

	if sql == "" {
		sql, err = templateSQL(template)
		if err != nil {
			return nil, err
		}
	}

	if sql == "" {
		sql, err = templateSQL(jimmyv1.Template_CREATE_TABLE)
		if err != nil {
			return nil, err
		}
	}

	if statementType == jimmyv1.Type_AUTOMATIC {
		statementType = detectType(sql)
	}

	sql = strings.TrimSpace(sql) + "\n"

	return &jimmyv1.Statement{
		Sql:  sql,
		Env:  env,
		Type: statementType,
	}, nil
}
