package consts

const (
	EnvPgHost     = "PG_HOST"
	EnvPgPort     = "PG_PORT"
	EnvPgUser     = "PG_USER"
	EnvPgPassword = "PG_PASSWORD"
	EnvPgDatabase = "PG_DATABASE"

	EnvMigrationDir     = "MIGRATION_DIR"
	EnvSessionSecretKey = "SESSION_SECRET_KEY"
	EnvGithubClientId   = "GITHUB_CLIENT_ID"
	// nolint:gosec
	EnvGithubClientSecret = "GITHUB_CLIENT_SECRET"

	EnvInitializationToken = "YATAI_INITIALIZATION_TOKEN"
	EnvAdminUserName       = "YATAI_ADMIN_USER_NAME"
	EnvAdminUserEmail      = "YATAI_ADMIN_USER_EMAIL"
	EnvAdminUserPassword   = "YATAI_ADMIN_USER_PASSWORD"
)
