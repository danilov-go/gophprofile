package config

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// ConfigServer описывает настройки сервера.
type ConfigServer struct {
	Net         NetAddress `mapstructure:"address"`
	DatabaseDSN string     `mapstructure:"database_dsn"`
	WebPath     string     `mapstructure:"web_path"`
	S3          S3Config   `mapstructure:"s3"`
}

// S3 описывает настройки подключения к MinIO / S3.
type S3Config struct {
	BucketName string `mapstructure:"bucket_name"`
	Region     string `mapstructure:"region"`
	Endpoint   string `mapstructure:"endpoint"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	UseSSL     bool   `mapstructure:"use_ssl"`
}

// Get парсит конфигурацию сервера.
func (c *ConfigServer) Get() error {
	v := viper.New()
	v.SetDefault("address", "0.0.0.0:8080")
	v.SetDefault("web_path", "./web/static/index.html")
	v.SetDefault("database_dsn", "")
	v.SetDefault("s3.bucket_name", "avatars")
	v.SetDefault("s3.region", "us-east-1")
	v.SetDefault("s3.endpoint", "")
	v.SetDefault("s3.access_key", "")
	v.SetDefault("s3.secret_key", "")
	v.SetDefault("s3.use_ssl", false)
	v.SetEnvPrefix("server")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.Unmarshal(c, viper.DecodeHook(
		mapstructure.TextUnmarshallerHookFunc(),
	)); err != nil {
		return err
	}
	return nil
}

// Validate проверяет конфигурацию сервера.
func (c *ConfigServer) Validate() error {
	if c.Net.Host == "" {
		return errors.New("хост не может быть пустым")
	}
	if c.Net.Port <= 0 || c.Net.Port > 65535 {
		return errors.New("передан неверный порт")
	}
	if c.DatabaseDSN == "" {
		return errors.New("строка подключения к базе данных не может быть пустой")
	}
	if c.WebPath == "" {
		return errors.New("путь к html-файлу не может быть пустым")
	}
	if c.S3.BucketName == "" {
		return errors.New("имя бакета S3 не может быть пустым")
	}
	if c.S3.Region == "" {
		return errors.New("регион S3 не может быть пустым")
	}
	if c.S3.Endpoint == "" {
		return errors.New("endpoint S3 не может быть пустым")
	}
	if c.S3.AccessKey == "" {
		return errors.New("access key S3 не может быть пустым")
	}
	if c.S3.SecretKey == "" {
		return errors.New("secret key S3 не может быть пустым")
	}
	return nil
}
