package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env-default:"development"` // environment
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	Database   DatabaseConfig   `yaml:"database"`
	JWT        JWTConfig        `yaml:"jwt"`
	Migrations MigrationsConfig `yaml:"migrations"`
}

// HTTPServerConfig структура http сервера
type HTTPServerConfig struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// DatabaseConfig структура по работе с БД
type DatabaseConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"-" env:"DB_PASSWORD" env-required:"true"`
	Name     string `yaml:"name" env-required:"true"`
}

// JWTConfig настройка jwt
type JWTConfig struct {
	Secret   string `yaml:"-" env:"JWT_SECRET" env-required:"true"`
	TokenTTL int    `yaml:"token_ttl" env-default:"60"`
}

type MigrationsConfig struct {
	Path string `yaml:"path" env-default:"./migrations"`
}

// MustLoad - если не загружаем - паникуем
func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		log.Fatal("CONFIG_PATH not exists")
	}
	return MustLoadByPath(configPath)
}

func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	return path
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not found: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("can't read config file %s", configPath)
	}

	return &cfg
}
