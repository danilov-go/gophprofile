package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validBaseConfig() BaseConfig {
	return BaseConfig{
		DatabaseDSN: "postgres://localhost:5432/gophprofile?sslmode=disable",
		S3: S3Config{
			BucketName: "avatars",
			Region:     "us-east-1",
			Endpoint:   "http://minio:9000",
			AccessKey:  "minioadmin",
			SecretKey:  "changeme",
			UseSSL:     false,
		},
		Kafka: KafkaConfig{
			Brokers: []string{"kafka:29092"},
			Topic:   "avatar-events",
			GroupID: "avatar-workers-group",
		},
	}
}

func TestConfigServer_Validate(t *testing.T) {
	setupCfg := func() ConfigServer {
		return ConfigServer{
			BaseConfig:   validBaseConfig(),
			Net:          "0.0.0.0:8080",
			WebPath:      "./web/static",
			SecretKey:    "test-secret-keys",
			PollInterval: time.Second,
			EventLimit:   50,
		}
	}
	tests := []struct {
		name    string
		modify  func(*ConfigServer)
		wantErr bool
	}{
		{
			name:    "положительный тест",
			modify:  func(c *ConfigServer) {},
			wantErr: false,
		},
		{
			name:    "пустой DATABASE_DSN",
			modify:  func(c *ConfigServer) { c.DatabaseDSN = "" },
			wantErr: true,
		},
		{
			name:    "пустой S3_BUCKET_NAME",
			modify:  func(c *ConfigServer) { c.S3.BucketName = "" },
			wantErr: true,
		},
		{
			name:    "пустой KAFKA_BROKERS",
			modify:  func(c *ConfigServer) { c.Kafka.Brokers = nil },
			wantErr: true,
		},
		{
			name:    "пустой KAFKA_TOPIC",
			modify:  func(c *ConfigServer) { c.Kafka.Topic = "" },
			wantErr: true,
		},
		{
			name:    "пустой KAFKA_GROUP_ID",
			modify:  func(c *ConfigServer) { c.Kafka.GroupID = "" },
			wantErr: true,
		},
		{
			name:    "пустой ADDRESS",
			modify:  func(c *ConfigServer) { c.Net = "  " },
			wantErr: true,
		},
		{
			name:    "короткий SECRET_KEY",
			modify:  func(c *ConfigServer) { c.SecretKey = "test" },
			wantErr: true,
		},
		{
			name:    "нулевой POLL_INTERVAL",
			modify:  func(c *ConfigServer) { c.PollInterval = 0 },
			wantErr: true,
		},
		{
			name:    "отрицательный EVENT_LIMIT",
			modify:  func(c *ConfigServer) { c.EventLimit = -1 },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := setupCfg()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestConfigWorker_Validate(t *testing.T) {
	setupCfg := func() ConfigWorker {
		return ConfigWorker{
			BaseConfig: validBaseConfig(),
			MaxRetries: 3,
			RetryDelay: 500 * time.Millisecond,
		}
	}
	tests := []struct {
		name    string
		modify  func(*ConfigWorker)
		wantErr bool
	}{
		{
			name:    "положительный тест",
			modify:  func(c *ConfigWorker) {},
			wantErr: false,
		},
		{
			name:    "пустой DatabaseDSN",
			modify:  func(c *ConfigWorker) { c.DatabaseDSN = "" },
			wantErr: true,
		},
		{
			name:    "отрицательный MAX_RETRIES",
			modify:  func(c *ConfigWorker) { c.MaxRetries = -1 },
			wantErr: true,
		},
		{
			name:    "нулевой RETRY_DELAY",
			modify:  func(c *ConfigWorker) { c.RetryDelay = 0 },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := setupCfg()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
