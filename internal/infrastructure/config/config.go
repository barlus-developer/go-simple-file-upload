package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	App     AppConfig
	Server  ServerConfig
	Storage StorageConfig
}

type AppConfig struct {
	Environment string
}

type ServerConfig struct {
	Host string
	Port int
}

// StorageConfig holds settings for the local-disk file storage used by
// the file upload/download API. There is no database involved.
type StorageConfig struct {
	Dir             string
	MaxUploadSizeMB int64
}

// MaxUploadSizeBytes returns the configured max upload size in bytes,
// for use as Gin's MaxMultipartMemory.
func (c StorageConfig) MaxUploadSizeBytes() int64 {
	return c.MaxUploadSizeMB << 20
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.environment", "development")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("storage.dir", "./uploads")
	v.SetDefault("storage.max_upload_size_mb", 32)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, err
		}
	}

	return Config{
		App: AppConfig{
			Environment: v.GetString("app.environment"),
		},
		Server: ServerConfig{
			Host: v.GetString("server.host"),
			Port: v.GetInt("server.port"),
		},
		Storage: StorageConfig{
			Dir:             v.GetString("storage.dir"),
			MaxUploadSizeMB: v.GetInt64("storage.max_upload_size_mb"),
		},
	}, nil
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c ServerConfig) ZapFields() []zap.Field {
	return []zap.Field{
		zap.String("host", c.Host),
		zap.Int("port", c.Port),
	}
}

func (c ServerConfig) ErrorField(err error) zap.Field {
	return zap.Error(err)
}
