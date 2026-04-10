package config

import (
	"fmt"
	"time"
)

type DBConfig struct {
	User              string        `env:"POSTGRES_USER,required"`
	Password          string        `env:"POSTGRES_PASSWORD,required"`
	Port              string        `env:"POSTGRES_PORT,required"`
	Host              string        `env:"POSTGRES_HOST,required"`
	Database          string        `env:"POSTGRES_DATABASE,required"`
	SSLMode           string        `env:"POSTGRES_SSLMODE,required"`
	MaxConns          int32         `env:"DB_MAX_CONNS" envDefault:"10"`
	MinConns          int32         `env:"DB_MIN_CONNS" envDefault:"1"`
	MaxConnLifetime   time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"30m"`
	MaxConnIdleTime   time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"5m"`
	HealthCheckPeriod time.Duration `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"1m"`
}

func (c *DBConfig) GetConnectString() string {
	info := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
	)

	return info
}
