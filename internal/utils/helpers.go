package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// GenerateRandomToken - Genera token aleatorio seguro (32 bytes = 64 chars hex)
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSecureToken - Genera token URL-safe (base64)
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando token seguro: %w", err)
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(bytes), "="), nil
}

// FormatDate - Formatea fecha a string
func FormatDate(t time.Time, format string) string {
	switch format {
	case "date":
		return t.Format("02/01/2006")
	case "datetime":
		return t.Format("02/01/2006 15:04:05")
	case "time":
		return t.Format("15:04:05")
	case "iso":
		return t.Format("2006-01-02T15:04:05Z")
	default:
		return t.Format("02/01/2006 15:04:05")
	}
}

// TruncateText - Trunca texto a longitud máxima
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

// Slugify - Convierte texto a slug URL-friendly
func Slugify(text string) string {
	text = strings.ToLower(text)
	text = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == ' ' || r == '_' || r == '-':
			return '-'
		default:
			return -1
		}
	}, text)

	for strings.Contains(text, "--") {
		text = strings.ReplaceAll(text, "--", "-")
	}
	return strings.Trim(text, "-")
}

// IsEmpty - Verifica si un string está vacío o solo espacios
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// DefaultString - Retorna valor por defecto si string está vacío
func DefaultString(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// Capitalize - Primera letra mayúscula
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
