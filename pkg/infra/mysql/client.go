package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"kei-services/pkg/config"
	"os"
	"time"

	gomysqldriver "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func Client(dbCfg *SqlDbConfig, sslCfg *config.Ssl, log *zap.Logger) (*gorm.DB, error) {
	log.Info("Connecting to SQL database...")
	log.Debug("Connection parameters",
		zap.String("User", dbCfg.User),
		zap.String("Host", dbCfg.Host),
		zap.Int("Port", dbCfg.Port),
		zap.String("DBName", dbCfg.DBName),
		zap.String("Params", dbCfg.Params),
	)

	if dbCfg.SSLEnabled {
		log.Debug("SSL enabled, using CA file", zap.String("CAFile", sslCfg.CAFile))
		if err := registerMySQLTLSCA(sslCfg.CAFile); err != nil {
			return nil, err
		}
		dbCfg.Params += "&tls=custom"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.DBName,
		dbCfg.Params,
	)

	gormLog := zapgorm2.New(log)
	gormLog.LogMode(logger.Info)
	gormLog.Info(context.Background(), "Gorm logger set to Info level")
	gormCfg := &gorm.Config{
		Logger: gormLog,
	}

	db, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// connection pool
	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(dbCfg.ConnMaxLifetimeMinutes) * time.Minute)

	// ping
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}

	log.Info("Connected to SQL database", zap.String("dsn", dsn))

	return db, nil
}

func registerMySQLTLSCA(caFile string) error {
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("failed to read CA file: %w", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("failed to append CA certs")
	}

	return gomysqldriver.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:            rootCertPool,
		InsecureSkipVerify: false,
	})
}
