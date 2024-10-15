package constants

const (
	AppName         = "jimmy"
	ConfigFile      = ".jimmy.yaml"
	MigrationsPath  = "./migrations"
	MigrationsTable = "migrations"

	EnvEmulatorHost        = "SPANNER_EMULATOR_HOST"
	EnvEmulatorHostDefault = "127.0.0.1:9010"
	EnvGoogleCloudProject  = "GOOGLE_CLOUD_PROJECT"
	EnvProjectId           = "SPANNER_PROJECT_ID"
	EnvInstanceId          = "SPANNER_INSTANCE_ID"
	EnvDatabaseId          = "SPANNER_DATABASE_ID"
)
