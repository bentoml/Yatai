package config

import (
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/common/consts"
)

type YataiServerConfigYaml struct {
	EnableHTTPS       bool   `yaml:"enable_https"`
	Port              uint   `yaml:"port"`
	SessionSecretKey  string `yaml:"session_secret_key"`
	MigrationDir      string `yaml:"migration_dir"`
	ReadHeaderTimeout int    `yaml:"read_header_timeout"`
}

type YataiPostgresqlConfigYaml struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type YataiS3ConfigYaml struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	Region     string `yaml:"region"`
	Secure     bool   `yaml:"secure"`
	BucketName string `yaml:"bucket_name"`
}

type YataiConfigYaml struct {
	IsSass              bool                      `yaml:"is_sass"`
	InCluster           bool                      `yaml:"in_cluster"`
	Server              YataiServerConfigYaml     `yaml:"server"`
	Postgresql          YataiPostgresqlConfigYaml `yaml:"postgresql"`
	NewsURL             string                    `yaml:"news_url"`
	InitializationToken string                    `yaml:"initialization_token"`
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
	pgSSLMode, ok := os.LookupEnv(consts.EnvPgSSLMode)
	if ok {
		YataiConfig.Postgresql.SSLMode = pgSSLMode
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

	readHeaderTimeout, ok := os.LookupEnv(consts.EnvReadHeaderTimeout)
	if ok {
		readHeaderTimeout_, err := strconv.Atoi(readHeaderTimeout)
		if err != nil {
			return errors.Wrapf(err, "convert %s from env to int", consts.EnvReadHeaderTimeout)
		}
		YataiConfig.Server.ReadHeaderTimeout = readHeaderTimeout_
	}

	initializationToken, ok := os.LookupEnv(consts.EnvInitializationToken)
	if ok {
		YataiConfig.InitializationToken = initializationToken
	}
	return nil
}
