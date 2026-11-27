// Package config управляет конфигурацией приложения.
package config

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// BaseConfig объединяет общие настройки для сервера и воркера.
type BaseConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN,notEmpty"`
	S3          S3Config
	Kafka       KafkaConfig
}

// Validate проверяет корректность общих настроек.
func (b *BaseConfig) Validate() error {
	if strings.TrimSpace(b.DatabaseDSN) == "" {
		return errors.New("DATABASE_DSN не может быть пустым")
	}
	if strings.TrimSpace(b.S3.BucketName) == "" {
		return errors.New("S3_BUCKET_NAME не может быть пустым")
	}
	if len(b.Kafka.Brokers) == 0 || strings.TrimSpace(b.Kafka.Brokers[0]) == "" {
		return errors.New("KAFKA_BROKERS должен содержать хотя бы один брокер")
	}
	if strings.TrimSpace(b.Kafka.Topic) == "" {
		return errors.New("KAFKA_TOPIC не может быть пустым")
	}
	if strings.TrimSpace(b.Kafka.GroupID) == "" {
		return errors.New("KAFKA_GROUP_ID не может быть пустым")
	}
	return nil
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
	_, _ = fmt.Fprintf(w, "Build version: %s\n", version)
	_, _ = fmt.Fprintf(w, "Build date: %s\n", date)
	_, _ = fmt.Fprintf(w, "Build commit: %s\n", commit)
}

// ConfigServer определяет настройки сервера.
type ConfigServer struct {
	BaseConfig

	Net          string        `env:"ADDRESS"  envDefault:"0.0.0.0:8080"`
	WebPath      string        `env:"WEB_PATH" envDefault:"./web/static"`
	SecretKey    string        `env:"SECRET_KEY"`
	PollInterval time.Duration `env:"POLL_INTERVAL" envDefault:"1s"`
	EventLimit   int           `env:"EVENT_LIMIT"   envDefault:"50"`
}

// Validate проверяет корректность настроек сервера.
func (c *ConfigServer) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.Net) == "" {
		return errors.New("ADDRESS не может быть пустым")
	}
	if len(strings.TrimSpace(c.SecretKey)) < 16 {
		return errors.New("SECRET_KEY должен содержать не менее 16 символов")
	}
	if c.PollInterval <= 0 {
		return errors.New("POLL_INTERVAL должен быть больше 0")
	}
	if c.EventLimit <= 0 {
		return errors.New("EVENT_LIMIT должен быть больше 0")
	}
	return nil
}

// Parse парсит конфигурацию сервера.
func (c *ConfigServer) Parse() error {
	return env.Parse(c)
}

// ConfigServer описывает настройки воркера.
type ConfigWorker struct {
	BaseConfig

	MaxRetries int           `env:"MAX_RETRIES"    envDefault:"3"`
	RetryDelay time.Duration `env:"RETRY_DELAY" envDefault:"500ms"`
}

// Parse парсит конфигурацию воркера.
func (c *ConfigWorker) Parse() error {
	return env.Parse(c)
}

// Validate проверяет корректность настроек воркера.
func (c *ConfigWorker) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	if c.MaxRetries < 0 {
		return errors.New("MAX_RETRIES не может быть меньше нуля")
	}
	if c.RetryDelay <= 0 {
		return errors.New("RETRY_DELAY не может быть равным или меньше нуля")
	}
	return nil
}
