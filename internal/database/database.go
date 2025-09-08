package database

import (
	"context"
	"fmt"
	"time"

	"realtime-api/internal/config"
	"realtime-api/internal/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

var DB *Database

func Init(cfg *config.DatabaseConfig) (*Database, error) {
	var dialector gorm.Dialector
	var dsn string

	switch cfg.Driver {
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.SSLMode)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dsn = cfg.Database
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Configure GORM logger to use our custom logger
	gormConfig := &gorm.Config{
		Logger: &GormLogger{},
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	database := &Database{DB: db}
	DB = database

	logger.Info("Database connected successfully", logger.WithFields(map[string]interface{}{
		"driver":   cfg.Driver,
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Database,
	}))

	return database, nil
}

func GetDB() *gorm.DB {
	if DB == nil {
		logger.Fatal("Database not initialized")
	}
	return DB.DB
}

func (db *Database) Migrate(models ...interface{}) error {
	for _, model := range models {
		if err := db.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}
	logger.Info("Database migration completed successfully")
	return nil
}

func (db *Database) Health() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GormLogger implements gorm logger interface using our custom logger
type GormLogger struct{}

func (l *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.Info(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.Error(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := map[string]interface{}{
		"elapsed": elapsed.String(),
		"rows":    rows,
		"sql":     sql,
	}

	if err != nil {
		logger.Error("Database query failed", logger.WithFields(fields))
	} else if elapsed > 200*time.Millisecond {
		logger.Warn("Slow database query", logger.WithFields(fields))
	} else {
		logger.Debug("Database query executed", logger.WithFields(fields))
	}
}
