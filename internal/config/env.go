package config

import (
	"os"
)

type Env struct {
	AppName            string
	AppPort            string
	AppEnv             string
	DBURL              string
	SessionSecret      string
	RecaptchaSecretKey string
	RecaptchaSiteKey   string
	BaseURL            string
}

func LoadEnv() *Env {
	return &Env{
		AppName:            getEnv("APP_NAME", "go-project"),
		AppPort:            getEnv("APP_PORT", "8080"),
		AppEnv:             getEnv("APP_ENV", "development"),
		DBURL:              getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		SessionSecret:      getEnv("SESSION_SECRET", "change-me-session-secret"),
		RecaptchaSecretKey: getEnv("RECAPTCHA_SECRET_KEY", ""),
		RecaptchaSiteKey:   getEnv("RECAPTCHA_SITE_KEY", ""),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// Helper para obtener entorno
func (e *Env) IsProduction() bool {
	return e.AppEnv == "production"
}

func (e *Env) IsDevelopment() bool {
	return e.AppEnv == "development"
}
