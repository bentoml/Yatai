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

type YataiOAuthGithubConfigYaml struct {
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type YataiOAuthConfigYaml struct {
	Github YataiOAuthGithubConfigYaml `yaml:"github"`
}

type YataiConfigYaml struct {
	IsSass     bool                      `yaml:"is_sass"`
	Server     YataiServerConfigYaml     `yaml:"server"`
	Postgresql YataiPostgresqlConfigYaml `yaml:"postgresql"`
	OAuth      YataiOAuthConfigYaml      `yaml:"oauth"`
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
	githubClientId, ok := os.LookupEnv(consts.EnvGithubClientId)
	if ok {
		YataiConfig.OAuth.Github.ClientId = githubClientId
	}
	githubClientSecret, ok := os.LookupEnv(consts.EnvGithubClientSecret)
	if ok {
		YataiConfig.OAuth.Github.ClientSecret = githubClientSecret
	}
	return nil
}
