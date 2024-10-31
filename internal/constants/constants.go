package constants

const (
	AppName         = "jimmy"
	FileExt         = ".yaml"
	ConfigFile      = ".jimmy" + FileExt
	MigrationsPath  = "./migrations"
	MigrationsTable = "migrations"

	EnvEmulatorHost        = "SPANNER_EMULATOR_HOST"
	EnvEmulatorHostDefault = "127.0.0.1:9010"
	EnvGoogleCloudProject  = "GOOGLE_CLOUD_PROJECT"
	EnvProjectID           = "SPANNER_PROJECT_ID"
	EnvInstanceID          = "SPANNER_INSTANCE_ID"
	EnvDatabaseID          = "SPANNER_DATABASE_ID"
)
