// internal/config/config.go
package config

import (
	"errors"
	"os"
	"time"
)

// Config содержит все параметры конфигурации приложения
type Config struct {
	JWTSecret     string        `json:"jwt_secret"`     // Секретный ключ для JWT
	AccessExpiry  time.Duration `json:"access_expiry"`  // Время жизни access токена
	RefreshExpiry time.Duration `json:"refresh_expiry"` // Время жизни refresh токена
	DBPath        string        `json:"db_path"`        // Путь к файлу базы данных SQLite
	ServerPort    string        `json:"server_port"`    // Порт HTTP сервера
	Env           string        `json:"env"`            // Окружение (development/production)
}

const (
	defaultJWTSecret     = "your-strong-secret-key"
	defaultAccessExpiry  = time.Hour * 1      // 1 час
	defaultRefreshExpiry = time.Hour * 24 * 7 // 1 неделя
	defaultDBPath        = "auth.db"
	defaultServerPort    = "8080"
)

// New создает конфигурацию в зависимости от окружения
func New() (*Config, error) {
	env := getEnv("APP_ENV", "development")

	switch env {
	case "production":
		return newProductionConfig()
	case "development":
		return newDevelopmentConfig()
	default:
		return nil, errors.New("unknown environment")
	}
}

// newDevelopmentConfig создает конфигурацию для разработки
func newDevelopmentConfig() (*Config, error) {
	return &Config{
		JWTSecret:     defaultJWTSecret,
		AccessExpiry:  defaultAccessExpiry,
		RefreshExpiry: defaultRefreshExpiry,
		DBPath:        defaultDBPath,
		ServerPort:    defaultServerPort,
		Env:           "development",
	}, nil
}

// newProductionConfig создает конфигурацию для production
func newProductionConfig() (*Config, error) {
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	return &Config{
		JWTSecret:     jwtSecret,
		AccessExpiry:  parseDuration(getEnv("ACCESS_EXPIRY", defaultAccessExpiry.String())),
		RefreshExpiry: parseDuration(getEnv("REFRESH_EXPIRY", defaultRefreshExpiry.String())),
		DBPath:        getEnv("DB_PATH", defaultDBPath),
		ServerPort:    getEnv("SERVER_PORT", defaultServerPort),
		Env:           "production",
	}, nil
}

// parseDuration преобразует строку в time.Duration с обработкой ошибок
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultAccessExpiry // возвращаем значение по умолчанию при ошибке
	}
	return d
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
