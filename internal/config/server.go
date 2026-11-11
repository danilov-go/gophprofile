package config

import (
	"errors"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ConfigServer описывает настройки клиента.
type ConfigServer struct {
	// Net содержит сетевой адрес и порт сервера.
	Net NetAddress `mapstructure:"address"`
	// ConfigPath содержит путь к конфигурационному файлу JSON.
	ConfigPath string `mapstructure:"config_path"`
	// DatabaseDSN содержит строку подключения к базе данных.
	DatabaseDSN string `mapstructure:"database_dsn"`
	// WebPath содержит путь к файлу html.
	WebPath string `mapstructure:"web_path"`
}

// Get парсит конфигурацию сервера.
func (c *ConfigServer) Get() error {
	if pflag.CommandLine.Lookup("address") == nil {
		pflag.CommandLine.StringP("address", "a", "", "Address and port to connect to server")
	}
	if pflag.CommandLine.Lookup("config_path") == nil {
		pflag.CommandLine.StringP("config_path", "c", "", "Path to server JSON config file")
	}
	if pflag.CommandLine.Lookup("database_dsn") == nil {
		pflag.CommandLine.StringP("database_dsn", "d", "", "DatabaseDSN")
	}
	if pflag.CommandLine.Lookup("web_path") == nil {
		pflag.CommandLine.StringP("web_path", "w", "", "Path to file index.html")
	}
	pflag.Parse()
	v := viper.New()
	v.SetDefault("address", "localhost:8080")
	v.SetDefault("web_path", "./web/static/index.html")
	v.SetEnvPrefix("server")
	v.AutomaticEnv()
	if err := v.BindPFlag("address", pflag.CommandLine.Lookup("address")); err != nil {
		return err
	}
	if err := v.BindPFlag("config_path", pflag.CommandLine.Lookup("config_path")); err != nil {
		return err
	}
	if err := v.BindPFlag("database_dsn", pflag.CommandLine.Lookup("database_dsn")); err != nil {
		return err
	}
	if err := v.BindPFlag("web_path", pflag.CommandLine.Lookup("web_path")); err != nil {
		return err
	}
	configFilePath := v.GetString("config_path")
	if configFilePath != "" {
		if _, err := os.Stat(configFilePath); err == nil {
			v.SetConfigFile(configFilePath)
			v.SetConfigType("json")
			if err := v.ReadInConfig(); err != nil {
				return err
			}
		}
	}
	err := v.Unmarshal(c, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
		),
	))
	if err != nil {
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
	return nil
}
