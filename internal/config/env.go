package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
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
	// Cargar archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No se pudo cargar archivo .env, usando variables de entorno del sistema")
	}

	env := &Env{
		AppName:            getEnv("APP_NAME", "go-project"),
		AppPort:            getEnv("APP_PORT", "8080"),
		AppEnv:             getEnv("APP_ENV", "development"),
		DBURL:              getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		SessionSecret:      getEnv("SESSION_SECRET", "change-me-session-secret"),
		RecaptchaSecretKey: getEnv("RECAPTCHA_SECRET_KEY", ""),
		RecaptchaSiteKey:   getEnv("RECAPTCHA_SITE_KEY", ""),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
	}

	// Log para verificar
	log.Printf("[CONFIG] RecaptchaSiteKey: %s", env.RecaptchaSiteKey)
	log.Printf("[CONFIG] RecaptchaSecretKey: %s", maskKey(env.RecaptchaSecretKey))

	return env
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func (e *Env) IsProduction() bool {
	return e.AppEnv == "production"
}

func (e *Env) IsDevelopment() bool {
	return e.AppEnv == "development"
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
