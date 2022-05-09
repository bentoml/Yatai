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

	// tracking related environment
	EnvYataiVersion       = "YATAI_VERSION"
	EnvYataiOrgUID        = "YATAI_ORG_UID"
	EnvYataiDeploymentUID = "YATAI_DEPLOYMENT_UID"
	EnvYataiClusterUID    = "YATAI_CLUSTER_UID"
)
