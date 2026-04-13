package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// ValidateName - Valida nombres (letras, espacios, acentos, ñ)
func ValidateName(name string) bool {
	if len(name) < 2 || len(name) > 100 {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}

// ValidateEmail - Valida formato de email
func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 100 {
		return false
	}
	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return regex.MatchString(email)
}

// ValidatePhone - Valida número de teléfono
func ValidatePhone(phone string) bool {
	clean := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r == '+' {
			return r
		}
		return -1
	}, phone)

	if strings.HasPrefix(clean, "+") {
		if len(clean) < 9 || len(clean) > 16 {
			return false
		}
		digits := clean[1:]
		regex := regexp.MustCompile(`^[0-9]+$`)
		return regex.MatchString(digits)
	}

	if len(clean) < 8 || len(clean) > 15 {
		return false
	}
	regex := regexp.MustCompile(`^[0-9]+$`)
	return regex.MatchString(clean)
}

// SanitizeInput - Limpia entrada de usuario
func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)

	htmlEscaper := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
		"`", "&#96;",
	)
	input = htmlEscaper.Replace(input)

	input = strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\r' && r != '\t') {
			return -1
		}
		return r
	}, input)

	return input
}

// IsValidUUID - Valida formato UUID
func IsValidUUID(uuid string) bool {
	regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return regex.MatchString(strings.ToLower(uuid))
}
