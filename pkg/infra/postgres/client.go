package postgres

import (
	"context"
	"fmt"
	"kei-services/pkg/config"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func Client(dbCfg *Config, sslCfg *config.Ssl, log *zap.Logger) (*gorm.DB, error) {
	if dbCfg == nil {
		return nil, fmt.Errorf("db config is nil")
	}

	log.Info("Connecting to Postgres...",
		zap.String("User", dbCfg.User),
		zap.String("Host", dbCfg.Host),
		zap.Int("Port", dbCfg.Port),
		zap.String("DBName", dbCfg.DBName),
		zap.Bool("ssl", dbCfg.SSLEnabled),
	)

	dsn, redacted := buildPostgresDSN(dbCfg, sslCfg)
	log.Debug("Postgres DSN", zap.String("dsn", redacted))

	gormLog := zapgorm2.New(log)
	gormLog.LogMode(logger.Info)
	gormLog.Info(context.Background(), "Gorm logger set to Info level")
	gormCfg := &gorm.Config{
		Logger: gormLog,
	}

	db, err := gorm.Open(pgdriver.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres sql.DB: %w", err)
	}

	// connection pool
	if dbCfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	}
	if dbCfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	}
	if dbCfg.ConnMaxLifetimeMinutes > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(dbCfg.ConnMaxLifetimeMinutes) * time.Minute)
	}

	// ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	log.Info("Connected to Postgres", zap.String("dsn", dsn))
	return db, nil
}

func buildPostgresDSN(db *Config, ssl *config.Ssl) (dsn, redacted string) {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", db.Host, db.Port),
		Path:   db.DBName, // leading slash added by URL.String()
	}
	if db.User != "" {
		u.User = url.UserPassword(db.User, db.Password)
	}

	// base params
	q := url.Values{}

	// custom params from config
	if strings.TrimSpace(db.Params) != "" {
		for _, kv := range strings.Split(db.Params, "&") {
			if kv == "" {
				continue
			}
			p := strings.SplitN(kv, "=", 2)
			if len(p) == 2 {
				q.Add(strings.TrimSpace(p[0]), strings.TrimSpace(p[1]))
			}
		}
	}

	// SSL
	if db.SSLEnabled {
		sslmode := "verify-full"
		if db.SSLMode != "" {
			sslmode = db.SSLMode
		}
		q.Set("sslmode", sslmode)

		if ssl != nil {
			if fileExists(ssl.CAFile) {
				q.Set("sslrootcert", ssl.CAFile)
			}
			if fileExists(ssl.CertFile) {
				q.Set("sslcert", ssl.CertFile)
			}
			if fileExists(ssl.KeyFile) {
				q.Set("sslkey", ssl.KeyFile)
			}
		}
	} else {
		q.Set("sslmode", "disable")
	}

	u.RawQuery = q.Encode()
	full := u.String()

	// redact password
	red := *u
	if red.User != nil {
		red.User = url.UserPassword(db.User, "********")
	}
	return full, red.String()
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}
