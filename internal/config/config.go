// Package config управляет конфигурацией приложения.
package config

import (
	"fmt"
	"io"

	"github.com/caarlos0/env/v11"
)

// BaseConfig объединяет общие настройки для сервера и воркера.
type BaseConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN,notEmpty"`
	S3          S3Config
	Kafka       KafkaConfig
}

// S3Config определяет настройки подключения к MinIO / S3.
type S3Config struct {
	BucketName string `env:"S3_BUCKET_NAME" envDefault:"avatars"`
	Region     string `env:"S3_REGION"      envDefault:"us-east-1"`
	Endpoint   string `env:"S3_ENDPOINT,notEmpty"`
	AccessKey  string `env:"S3_ACCESS_KEY,notEmpty"`
	SecretKey  string `env:"S3_SECRET_KEY,notEmpty"`
	UseSSL     bool   `env:"S3_USE_SSL"     envDefault:"false"`
}

// KafkaConfig определяет настройки подключения к Kafka.
type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS"  envDefault:"localhost:9092"`
	Topic   string   `env:"KAFKA_TOPIC"    envDefault:"avatar-events"`
	GroupID string   `env:"KAFKA_GROUP_ID" envDefault:"avatar-workers-group"`
}

// PrintBuild выводит в консоль информацию о текущей сборке приложения.
func PrintBuild(w io.Writer, version, date, commit string) {
	fmt.Fprintf(w, "Build version: %s\n", version)
	fmt.Fprintf(w, "Build date: %s\n", date)
	fmt.Fprintf(w, "Build commit: %s\n", commit)
}

// ConfigServer определяет настройки сервера.
type ConfigServer struct {
	Net     string `env:"ADDRESS"  envDefault:"0.0.0.0:8080"`
	WebPath string `env:"WEB_PATH" envDefault:"./web/static/index.html"`
	BaseConfig
}

// Parse парсит конфигурацию сервера.
func (c *ConfigServer) Parse() error {
	return env.Parse(c)
}

// ConfigServer описывает настройки воркера.
type ConfigWorker struct {
	BaseConfig
}

// Parse парсит конфигурацию воркера.
func (c *ConfigWorker) Parse() error {
	return env.Parse(c)
}
