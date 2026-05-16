package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Cache    CacheConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

type CacheConfig struct {
	FlushInterval int
	MaxEntries    int
}

type JWTConfig struct {
	Secret        string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	// 支持以环境变量覆盖嵌套配置，例如 DATABASE_HOST、JWT_SECRET（Docker / K8s）
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
			Mode: viper.GetString("server.mode"),
		},
		Database: DatabaseConfig{
			Host:         viper.GetString("database.host"),
			Port:         viper.GetInt("database.port"),
			User:         viper.GetString("database.user"),
			Password:     viper.GetString("database.password"),
			DBName:       viper.GetString("database.dbname"),
			MaxOpenConns: viper.GetInt("database.max_open_conns"),
			MaxIdleConns: viper.GetInt("database.max_idle_conns"),
		},
		Cache: CacheConfig{
			FlushInterval: viper.GetInt("cache.flush_interval"),
			MaxEntries:    viper.GetInt("cache.max_entries"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("jwt.secret"),
			AccessExpire:  viper.GetDuration("jwt.access_expire"),
			RefreshExpire: viper.GetDuration("jwt.refresh_expire"),
		},
	}
	return nil
}
