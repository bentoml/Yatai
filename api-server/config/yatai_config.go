package config

import (
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/common/consts"
)

type YataiServerConfigYaml struct {
	EnableHTTPS          bool   `yaml:"enable_https"`
	Port                 uint   `yaml:"port"`
	SessionSecretKey     string `yaml:"session_secret_key"`
	MigrationDir         string `yaml:"migration_dir"`
	ReadHeaderTimeout    int    `yaml:"read_header_timeout"`
	TransmissionStrategy string `yaml:"transmission_strategy"`
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

type YataiDockerRegistryConfigYaml struct {
	BentoRepositoryName string `yaml:"bento_repository_name"`
	ModelRepositoryName string `yaml:"model_repository_name"`
	Server              string `yaml:"server"`
	Username            string `yaml:"username"`
	Password            string `yaml:"password"`
	Secure              bool   `yaml:"secure"`
}

type YataiDockerImageBuilderConfigYaml struct {
	Privileged bool `yaml:"privileged"`
}

type YataiConfigYaml struct {
	IsSaaS              bool                      `yaml:"is_saas"`
	SaasDomainSuffix    string                    `yaml:"saas_domain_suffix"`
	InCluster           bool                      `yaml:"in_cluster"`
	Server              YataiServerConfigYaml     `yaml:"server"`
	Postgresql          YataiPostgresqlConfigYaml `yaml:"postgresql"`
	S3                  *YataiS3ConfigYaml        `yaml:"s3,omitempty"`
	NewsURL             string                    `yaml:"news_url"`
	InitializationToken string                    `yaml:"initialization_token"`
}

var YataiConfig = &YataiConfigYaml{}

func PopulateYataiConfig() error {
	isSaaS, ok := os.LookupEnv(consts.EnvIsSaaS)
	if ok {
		YataiConfig.IsSaaS = isSaaS == "true"
	}

	saasDomainSuffix, ok := os.LookupEnv(consts.EnvSaasDomainSuffix)
	if ok {
		YataiConfig.SaasDomainSuffix = saasDomainSuffix
	}

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

	transmissionStrategy, ok := os.LookupEnv(consts.EnvTransmissionStrategy)
	if ok {
		YataiConfig.Server.TransmissionStrategy = transmissionStrategy
	}

	initializationToken, ok := os.LookupEnv(consts.EnvInitializationToken)
	if ok {
		YataiConfig.InitializationToken = initializationToken
	}
	makesureS3IsNotNil := func() {
		if YataiConfig.S3 == nil {
			YataiConfig.S3 = &YataiS3ConfigYaml{}
		}
	}
	s3Endpoint, ok := os.LookupEnv(consts.EnvS3Endpoint)
	if ok {
		makesureS3IsNotNil()
		YataiConfig.S3.Endpoint = s3Endpoint
	}
	s3AccessKey, ok := os.LookupEnv(consts.EnvS3AccessKey)
	if ok {
		makesureS3IsNotNil()
		YataiConfig.S3.AccessKey = s3AccessKey
	}
	s3SecretKey, ok := os.LookupEnv(consts.EnvS3SecretKey)
	if ok {
		makesureS3IsNotNil()
		YataiConfig.S3.SecretKey = s3SecretKey
	}
	s3Region, ok := os.LookupEnv(consts.EnvS3Region)
	if ok {
		makesureS3IsNotNil()
		YataiConfig.S3.Region = s3Region
	}
	s3Secure, ok := os.LookupEnv(consts.EnvS3Secure)
	if ok {
		makesureS3IsNotNil()
		s3Secure_, err := strconv.ParseBool(s3Secure)
		if err != nil {
			return errors.Wrap(err, "convert s3_secure from env to bool")
		}
		YataiConfig.S3.Secure = s3Secure_
	}
	s3BucketName, ok := os.LookupEnv(consts.EnvS3BucketName)
	if ok {
		makesureS3IsNotNil()
		YataiConfig.S3.BucketName = s3BucketName
	}
	return nil
}
