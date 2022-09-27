package consts

const (
	EnvIsSaaS           = "IS_SAAS"
	EnvSassDomainSuffix = "SASS_DOMAIN_SUFFIX"

	EnvPgHost     = "PG_HOST"
	EnvPgPort     = "PG_PORT"
	EnvPgUser     = "PG_USER"
	EnvPgPassword = "PG_PASSWORD"
	EnvPgDatabase = "PG_DATABASE"
	EnvPgSSLMode  = "PG_SSLMODE"

	EnvMigrationDir     = "MIGRATION_DIR"
	EnvSessionSecretKey = "SESSION_SECRET_KEY"
	EnvGithubClientId   = "GITHUB_CLIENT_ID"
	// nolint:gosec
	EnvGithubClientSecret = "GITHUB_CLIENT_SECRET"

	EnvInitializationToken = "YATAI_INITIALIZATION_TOKEN"

	EnvS3Endpoint   = "S3_ENDPOINT"
	EnvS3Region     = "S3_REGION"
	EnvS3BucketName = "S3_BUCKET_NAME"
	EnvS3AccessKey  = "S3_ACCESS_KEY"
	// nolint:gosec
	EnvS3SecretKey = "S3_SECRET_KEY"
	EnvS3Secure    = "S3_SECURE"

	EnvDockerRegistryServer   = "DOCKER_REGISTRY_SERVER"
	EnvDockerRegistryUsername = "DOCKER_REGISTRY_USERNAME"
	// nolint:gosec
	EnvDockerRegistryPassword            = "DOCKER_REGISTRY_PASSWORD"
	EnvDockerRegistrySecure              = "DOCKER_REGISTRY_SECURE"
	EnvDockerRegistryBentoRepositoryName = "DOCKER_REGISTRY_BENTO_REPOSITORY_NAME"
	EnvDockerRegistryModelRepositoryName = "DOCKER_REGISTRY_MODEL_REPOSITORY_NAME"

	EnvDockerImageBuilderPrivileged = "DOCKER_IMAGE_BUILDER_PRIVILEGED"

	EnvReadHeaderTimeout = "READ_HEADER_TIMEOUT"

	EnvTransmissionStrategy = "TRANSMISSION_STRATEGY"
)
