package migrations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type BootstrapInput struct {
	Name string
}

func (ms *Migrations) Bootstrap(ctx context.Context, input BootstrapInput) (*Migration, error) {
	slug := Slugify(input.Name)
	if slug == "" {
		slug = "init"
	}

	err := ms.ensureAll(ctx)
	if err != nil {
		return nil, err
	}

	var upgrade []*jimmyv1.Statement

	dbAdmin, err := ms.DatabaseAdmin(ctx)
	if err != nil {
		return nil, err
	}

	ddl, err := dbAdmin.GetDatabaseDdl(ctx, &databasepb.GetDatabaseDdlRequest{
		Database: ms.DatabaseName(),
	})
	if err != nil {
		return nil, err
	}

	if len(ddl.Statements) == 0 {
		return nil, errors.New("no statements")
	}

	hasFileDescriptorSet := len(ddl.ProtoDescriptors) > 0

	migrationTableDDL := fmt.Sprintf("CREATE TABLE %s (", ms.Config.Table)

	for _, sql := range ddl.Statements {
		if strings.Contains(sql, migrationTableDDL) {
			continue
		}

		statement, err := ms.newStatement(
			sql,
			jimmyv1.Environment_ALL,
			"",
			jimmyv1.Type_DDL,
		)
		if err != nil {
			return nil, err
		}

		if hasFileDescriptorSet && isProtoDDL(sql) {
			statement.FileDescriptorSet = Ref(constants.UpgradeFileDescriptorSet)
		}

		upgrade = append(upgrade, statement)
	}

	data := &jimmyv1.Migration{
		Upgrade:            upgrade,
		FileDescriptorSets: map[string]*descriptorpb.FileDescriptorSet{},
		SquashId:           Ref[int32](0),
	}

	if hasFileDescriptorSet {
		fileDescriptorSet := &descriptorpb.FileDescriptorSet{}

		err = proto.Unmarshal(ddl.ProtoDescriptors, fileDescriptorSet)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal file descriptor set: %w", err)
		}

		data.FileDescriptorSets[constants.UpgradeFileDescriptorSet] = fileDescriptorSet
	}

	return ms.create(slug, data)
}
