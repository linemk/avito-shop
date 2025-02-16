package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/linemk/avito-shop/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMustLoadByPath_Success(t *testing.T) {
	// Устанавливаем обязательные переменные окружения
	os.Setenv("DB_PASSWORD", "mypassword")
	os.Setenv("JWT_SECRET", "mysecret")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("JWT_SECRET")

	// Пример содержимого конфигурационного файла
	content := `
env: "local"
http_server:
  address: "localhost:8080"
  timeout: "4s"
  idle_timeout: "60s"
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  name: "shop"
jwt:
  token_ttl: 60
migrations:
  path: "./migrations"
`
	// Создаем временный файл с конфигурацией
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// Загружаем конфигурацию из временного файла
	cfg := config.MustLoadByPath(tmpFile.Name())

	// Проверяем, что конфигурация загружена корректно
	assert.Equal(t, "local", cfg.Env)
	assert.Equal(t, "localhost:8080", cfg.HTTPServer.Address)
	assert.Equal(t, 4*time.Second, cfg.HTTPServer.Timeout)
	assert.Equal(t, 60*time.Second, cfg.HTTPServer.IdleTimeout)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "shop", cfg.Database.Name)
	assert.Equal(t, 60, cfg.JWT.TokenTTL)
	assert.Equal(t, "./migrations", cfg.Migrations.Path)
}

func TestMustLoadByPath_FileNotFound(t *testing.T) {
	// Ожидаем панику, если файла не существует
	assert.Panics(t, func() {
		config.MustLoadByPath("non_existent_config.yaml")
	})
}
