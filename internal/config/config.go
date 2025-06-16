package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	Env        string `mapstructure:"ENV" env-default:"prod"`
	GRPSServer `mapstructure:",squash"`
	TLS        `mapstructure:",squash"`
	DataBase   `mapstructure:",squash"`
	Redis      `mapstructure:",squash"`
	JWT        `mapstructure:",squash"`
	JWTKeys    `mapstructure:",squash"`
}

type GRPSServer struct {
	Port    int `mapstructure:"SERVER_PORT" env-default:"50051"`
	Timeout int `mapstructure:"SERVER_TIMEOUT" env-default:"10"`
}

type TLS struct {
	TLSKey  string `mapstructure:"TLS_SERVER" env-default:"./server.example.key"`
	TLSCert string `mapstructure:"TLS_CERT" env-default:"./server.example.crt"`
}

type DataBase struct {
	Host     string `mapstructure:"DB_HOST" env-default:"postgres"`
	Port     int    `mapstructure:"DB_PORT" env-default:"5432"`
	User     string `mapstructure:"DB_USER" env-default:"postgres"`
	Password string `mapstructure:"DB_PASSWORD" env-default:"secret"`
	Name     string `mapstructure:"DB_NAME" env-default:"payment_db"`
	SSLMode  string `mapstructure:"SSL_MODE" env-default:"disable"`
}

type Redis struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type JWT struct {
	AccessTokenTTL  time.Duration `mapstructure:"JWT_ACCESS_TOKEN_TTL" env-default:"15m"`
	RefreshTokenTTL time.Duration `mapstructure:"JWT_REFRESH_TOKEN_TTL" env-default:"168h"`
	Issuer          string        `mapstructure:"JWT_ISSUER" env-default:"payment-system"`
}

type JWTKeys struct {
	JWTPrivateKeyPath string `mapstructure:"JWT_KEY_PRIVATE" env-default:"./private.example.pem"`
	JWTPublicKeyPath  string `mapstructure:"JWT_KEY_PUBLIC" env-default:"./public.example.pem"`
}

func (c DataBase) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func LoadConfig(env string) (config *Configuration, err error) {
	v := viper.New()
	v.SetConfigName(fmt.Sprintf(".env.%s", env))
	v.AddConfigPath(".")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
