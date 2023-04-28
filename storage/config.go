package storage

import (
	"errors"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"
	"time"
)

const (
	defaultMaxIdleConns    = 5
	defaultMaxOpenConns    = 40
	defaultConnMaxLifetime = time.Hour
)

type Config struct {
	Type    string `env:"DB_TYPE" required:"true"`
	Network string `env:"DB_NETWORK"` // Network type, either tcp or unix, Default is tcp
	Host    string `env:"DB_HOST"`    // TCP host:port or Unix socket depending on Network
	Port    string `env:"DB_PORT"`

	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Database string `env:"DB_DATABASE" required:"true"`

	SSLMode      string            `yaml:"sslMode"`
	CertFile     string            `yaml:"sslCertFile"`
	KeyFile      string            `yaml:"sslKeyFile"`
	RootCertFile string            `yaml:"sslRootCertFile"`
	ConnPool     ConnPoolConfig    `yaml:"connPool"`
	Params       map[string]string `yaml:"params"`
	Log          *LogConfig        `yaml:"log"`
}

type LogConfig struct {
	Stdout                    bool               `yaml:"stdout"`
	Level                     string             `yaml:"level"`
	Colorful                  bool               `yaml:"colorful"`
	SlowThreshold             time.Duration      `yaml:"slowThreshold" default:"200ms"`
	IgnoreRecordNotFoundError bool               `yaml:"ignoreRecordNotFoundError"`
	Logger                    *lumberjack.Logger `yaml:"logger"`
}

type ConnPoolConfig struct {
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
}

func (cfg *Config) LoggerConfig() (logger.Config, error) {
	if cfg.Log == nil {
		return logger.Config{}, nil
	}

	var logLevel logger.LogLevel
	switch cfg.Log.Level {
	case "Silent":
		logLevel = logger.Silent
	case "Error":
		logLevel = logger.Error
	case "", "Warn":
		logLevel = logger.Warn
	case "Info":
		logLevel = logger.Info
	default:
		return logger.Config{}, errors.New("log level must be one of [Silent, Error, Warn, Info], Default is 'Warn'")
	}

	return logger.Config{
		SlowThreshold:             cfg.Log.SlowThreshold,
		LogLevel:                  logLevel,
		IgnoreRecordNotFoundError: cfg.Log.IgnoreRecordNotFoundError,
		Colorful:                  cfg.Log.Colorful,
	}, nil
}

func (cfg *Config) getConnPoolConfig() (ConnPoolConfig, error) {
	connPool := ConnPoolConfig{
		MaxIdleConns:    cfg.ConnPool.MaxIdleConns,
		MaxOpenConns:    cfg.ConnPool.MaxOpenConns,
		ConnMaxLifetime: cfg.ConnPool.ConnMaxLifetime,
	}
	if connPool.MaxIdleConns <= 0 {
		connPool.MaxIdleConns = defaultMaxIdleConns
	}
	if connPool.MaxOpenConns <= 0 {
		connPool.MaxOpenConns = defaultMaxOpenConns
	}
	lifeTimeSeconds := connPool.ConnMaxLifetime.Seconds()
	if lifeTimeSeconds > 0 && lifeTimeSeconds < defaultConnMaxLifetime.Seconds() {
		connPool.ConnMaxLifetime = defaultConnMaxLifetime
	}

	if connPool.MaxOpenConns < connPool.MaxIdleConns {
		return ConnPoolConfig{}, fmt.Errorf("connPool maxIdleConns is bigger than maxOpenConns, config detail: %v, please check the config", connPool)
	}
	return connPool, nil
}
