package constants

const CreateTable = `
CREATE TABLE test (
  id STRING(MAX) NOT NULL,
  name STRING(MAX),
  update_time TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (id)
`

const DropTable = `
DROP TABLE test
`

const AddColumn = `
ALTER TABLE test ADD COLUMN slug STRING(MAX) NOT NULL DEFAULT (GENERATE_UUID())
`

const DropColumn = `
ALTER TABLE test DROP COLUMN slug
`

const SetDefault = `
ALTER TABLE test ALTER COLUMN id SET DEFAULT (GENERATE_UUID())
`

const DropDefault = `
ALTER TABLE test ALTER COLUMN id DROP DEFAULT
`

const AddCheckConstraint = `
ALTER TABLE test ADD CONSTRAINT ck_test_slug CHECK (STARTS_WITH(name, "a"))
`

const AddForeignKeyConstraint = `
ALTER TABLE test ADD CONSTRAINT fk_test_slug
  FOREIGN KEY (slug) REFERENCES test2 (slug) ON DELETE CASCADE
`

const DropConstraint = `
ALTER TABLE test DROP CONSTRAINT ck_test_slug
`

const CreateIndex = `
CREATE UNIQUE NULL_FILTERED INDEX uq_test_slug ON test (slug)
`

const DropIndex = `
DROP INDEX uq_test_slug
`
