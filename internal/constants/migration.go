package constants

const SelectMigrationsTable = `
SELECT 1
FROM information_schema.tables
WHERE table_schema = @tableSchema AND table_name = @tableName
`

const SelectMigration = `
SELECT id, complete_time IS NOT NULL
FROM %s
ORDER BY id DESC
LIMIT 1
`

const CreateMigrationTable = `
CREATE TABLE IF NOT EXISTS %s (
  id INT64 NOT NULL,
  start_time TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  complete_time TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (id)
`
