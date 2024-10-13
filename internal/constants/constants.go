package constants

const (
	AppName         = "jimmy"
	ConfigFile      = ".jimmy.yaml"
	MigrationsPath  = "./migrations"
	MigrationsTable = "migrations"

	EnvEmulatorHost        = "SPANNER_EMULATOR_HOST"
	EnvEmulatorHostDefault = "127.0.0.1:9010"
	EnvProject             = "SPANNER_PROJECT"
	EnvInstance            = "SPANNER_INSTANCE"
	EnvDatabase            = "SPANNER_DATABASE"
)
