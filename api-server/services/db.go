package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bentoml/yatai/common/command"

	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/common/utils"

	"github.com/uber/jaeger-client-go"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/bentoml/yatai/common/tracing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	postgres "go.elastic.co/apm/module/apmgormv2/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbCache   = map[string]*gorm.DB{}
	dbCacheRW sync.RWMutex
)

type DbCtxKeyType string

const DbSessionKey DbCtxKeyType = "session"

type GormLogger struct{}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l.getLogger(context.Background()).LogMode(level)
}

func (l *GormLogger) getLogger(ctx context.Context) gormlogger.Interface {
	prefix := "\r\n"
	span := tracing.GetSpan(ctx)
	if span != nil {
		if sc, ok := span.Context().(jaeger.SpanContext); ok {
			prefix = fmt.Sprintf("%straceID=%s ", prefix, sc.TraceID().String())
		}
	}
	return gormlogger.New(log.New(os.Stdout, prefix, log.LstdFlags), gormlogger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      gormlogger.Warn,
		Colorful:      true,
	})
}

func (l *GormLogger) Info(ctx context.Context, format string, args ...interface{}) {
	l.getLogger(ctx).Info(ctx, format, args...)
}

func (l *GormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	l.getLogger(ctx).Warn(ctx, format, args...)
}

func (l *GormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	l.getLogger(ctx).Error(ctx, format, args...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.getLogger(ctx).Trace(ctx, begin, fc, err)
}

func getPgHost() string {
	return config.YataiConfig.Postgresql.Host
}

// nolint: unparam
func getDBURI() (string, error) {
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.YataiConfig.Postgresql.User,
		config.YataiConfig.Postgresql.Password,
		getPgHost(),
		config.YataiConfig.Postgresql.Port,
		config.YataiConfig.Postgresql.Database)
	return uri, nil
}

func openDB() (*gorm.DB, error) {
	uri, err := getDBURI()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get db uri")
	}

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		PrepareStmt:    false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}
	return db, nil
}

func getDB() (*gorm.DB, error) {
	uri, err := getDBURI()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get db uri")
	}
	dbCacheRW.RLock()
	db, ok := dbCache[uri]
	dbCacheRW.RUnlock()
	if !ok {
		db, err = openDB()
		if err != nil {
			return nil, err
		}
		var rawDb *sql.DB
		rawDb, err = db.DB()
		if err != nil {
			return nil, err
		}
		rawDb.SetMaxOpenConns(25)
		rawDb.SetMaxIdleConns(25)
		rawDb.SetConnMaxLifetime(5 * time.Minute)
		dbCacheRW.Lock()
		dbCache[uri] = db
		dbCacheRW.Unlock()
	}

	if command.GlobalCommandOption.Debug {
		return db.Debug(), nil
	}

	return db, nil
}

// nolint: unparam
func startTransaction(ctx context.Context) (*gorm.DB, context.Context, func(error), error) {
	// FIXME: pq: unexpected Parse response 'D'
	defaultCb := func(err error) {}
	// return mustGetDB(), ctx, defaultCb, nil
	session_ := ctx.Value(DbSessionKey)
	if session_ != nil {
		return session_.(*gorm.DB), ctx, defaultCb, nil
	}
	db := mustGetDB(ctx)
	tx := db.Begin()
	if tx.Error != nil {
		return nil, ctx, defaultCb, tx.Error
	}
	ctx = context.WithValue(ctx, DbSessionKey, tx)
	return tx, ctx, func(err error) {
		// nolint: gocritic
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}, nil
}

func mustGetSession(ctx context.Context) *gorm.DB {
	session_ := ctx.Value(DbSessionKey)
	if session_ != nil {
		return session_.(*gorm.DB)
	}
	return mustGetDB(ctx)
}

func mustGetDB(ctx context.Context) *gorm.DB {
	db, err := getDB()
	if err != nil {
		panic(fmt.Sprintf("cannot get db: %s", err.Error()))
	}
	db = db.WithContext(ctx)
	return db
}

type MigrateLog struct{}

func (*MigrateLog) Printf(format string, v ...interface{}) {
	logrus.Infof(fmt.Sprintf("[%s] %s", time.Now(), format), v...)
}

func (*MigrateLog) Verbose() bool {
	return false
}

func MigrateUp() error {
	uri, err := getDBURI()
	if err != nil {
		return errors.Wrap(err, "cannot get db uri")
	}

	logrus.Debugf("db uri: %s", uri)
	migrationDir := config.YataiConfig.Server.MigrationDir

	exists, err := utils.PathExists(migrationDir)
	if err != nil {
		return errors.Wrapf(err, "check migration dir exists: %s", migrationDir)
	}
	if !exists {
		return errors.Errorf("migration dir is not exists: %s", migrationDir)
	}

	logrus.Debugf("migration dir: %s", migrationDir)
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationDir),
		uri,
	)
	if err != nil {
		return errors.Wrap(err, "cannot create migrate")
	}

	m.Log = &MigrateLog{}

	logrus.Info("migrate up...")
	if err := m.Up(); err != nil && !strings.Contains(err.Error(), "no change") {
		return errors.Wrap(err, "cannot migrate up")
	}
	logrus.Info("[DONE] migrate up")
	return nil
}
