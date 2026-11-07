package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BasePath   string
	AppPort    string
	PostgreSQL PostgreSQLConfig
	JWT        JWTConfig
}

type BaseDB struct {
	Host string
	Port string
	Pass string
}

type PostgreSQLConfig struct {
	BaseDB
	Name    string
	User    string
	Schema  string
	SSLMode string
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

func getEnvOrDefault(key, def string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}

	return value
}

func LoadConfig() *Config {
	// โหลด .env จาก working directory
	envFile := strings.TrimSpace(os.Getenv("APP_ENV_FILE"))

	if envFile != "" {
		_ = godotenv.Load(envFile)
	} else {
		_ = godotenv.Load()
	}

	return &Config{
		BasePath: ResolveAppBaseFromExe(),

		AppPort: getEnvOrDefault("APP_PORT", "8080"),

		PostgreSQL: PostgreSQLConfig{
			BaseDB: BaseDB{
				Host: getEnvOrDefault("DB_HOST", "localhost"),
				Port: getEnvOrDefault("DB_PORT", "5432"),
				Pass: os.Getenv("DB_PASSWORD"),
			},
			User:    os.Getenv("DB_USER"),
			Name:    os.Getenv("DB_NAME"),
			Schema:  getEnvOrDefault("DB_SCHEMA", "public"),
			SSLMode: getEnvOrDefault("DB_SSLMODE", "disable"),
		},

		JWT: JWTConfig{
			Secret:    os.Getenv("JWT_SECRET"),
			ExpiresIn: 1 * time.Hour,
		},
	}
}

func ResolveAppBaseFromExe() string {
	// สำหรับการพัฒนาผ่าน go run ให้ใช้ working directory ก่อน
	wd, err := os.Getwd()
	if err == nil && wd != "" {
		return wd
	}

	// Fallback ไปยังตำแหน่ง executable
	exe, err := os.Executable()
	if err != nil {
		return "."
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return filepath.Dir(exe)
	}

	return filepath.Dir(exe)
}
