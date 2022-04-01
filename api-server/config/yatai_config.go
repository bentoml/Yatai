package config

import (
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/common/consts"
)

type YataiServerConfigYaml struct {
	EnableHTTPS      bool   `yaml:"enable_https"`
	Port             uint   `yaml:"port"`
	SessionSecretKey string `yaml:"session_secret_key"`
	MigrationDir     string `yaml:"migration_dir"`
}

type YataiPostgresqlConfigYaml struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type YataiS3ConfigYaml struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	Region     string `yaml:"region"`
	Secure     bool   `yaml:"secure"`
	BucketName string `yaml:"bucket_name"`
}

type YataiDockerRegistryConfigYaml struct {
	BentoRepositoryName string `yaml:"bento_repository_name"`
	ModelRepositoryName string `yaml:"model_repository_name"`
	Server              string `yaml:"server"`
	Username            string `yaml:"username"`
	Password            string `yaml:"password"`
	Secure              bool   `yaml:"secure"`
}

type YataiConfigYaml struct {
	IsSass              bool                           `yaml:"is_sass"`
	InCluster           bool                           `yaml:"in_cluster"`
	Server              YataiServerConfigYaml          `yaml:"server"`
	Postgresql          YataiPostgresqlConfigYaml      `yaml:"postgresql"`
	S3                  *YataiS3ConfigYaml             `yaml:"s3,omitempty"`
	DockerRegistry      *YataiDockerRegistryConfigYaml `yaml:"docker_registry,omitempty"`
	NewsURL             string                         `yaml:"news_url"`
	InitializationToken string                         `yaml:"initialization_token"`
}

var YataiConfig = &YataiConfigYaml{}

func PopulateYataiConfig() error {
	pgHost, ok := os.LookupEnv(consts.EnvPgHost)
	if ok {
		YataiConfig.Postgresql.Host = pgHost
	}
	pgPort, ok := os.LookupEnv(consts.EnvPgPort)
	if ok {
		pgPort_, err := strconv.Atoi(pgPort)
		if err != nil {
			return errors.Wrap(err, "convert port from env to int")
		}
		YataiConfig.Postgresql.Port = uint(pgPort_)
	}
	pgUser, ok := os.LookupEnv(consts.EnvPgUser)
	if ok {
		YataiConfig.Postgresql.User = pgUser
	}
	pgPassword, ok := os.LookupEnv(consts.EnvPgPassword)
	if ok {
		YataiConfig.Postgresql.Password = pgPassword
	}
	pgDatabase, ok := os.LookupEnv(consts.EnvPgDatabase)
	if ok {
		YataiConfig.Postgresql.Database = pgDatabase
	}
	migrationDir, ok := os.LookupEnv(consts.EnvMigrationDir)
	if ok {
		YataiConfig.Server.MigrationDir = migrationDir
	}
	sessionSecretKey, ok := os.LookupEnv(consts.EnvSessionSecretKey)
	if ok {
		YataiConfig.Server.SessionSecretKey = sessionSecretKey
	}
	if YataiConfig.Server.Port == 0 {
		YataiConfig.Server.Port = 7777
	}
	initialization_token, ok := os.LookupEnv(consts.EnvInitializationToken)
	if ok {
		YataiConfig.InitializationToken = initialization_token
	}
	return nil
}
