package utils

import (
	"regexp"
	"strings"
)

// ValidatePasswordStrength - Valida fortaleza de contraseña
func ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < 6 {
		return false, "La contraseña debe tener al menos 6 caracteres"
	}
	if len(password) > 72 {
		return false, "La contraseña no puede exceder 72 caracteres"
	}
	if strings.Contains(password, " ") {
		return false, "La contraseña no puede contener espacios"
	}
	return true, ""
}

// GetPasswordStrength - Retorna nivel de fortaleza (0-4)
func GetPasswordStrength(password string) int {
	strength := 0

	if len(password) >= 8 {
		strength++
	}
	if len(password) >= 12 {
		strength++
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	if hasUpper && hasLower {
		strength++
	}
	if hasNumber {
		strength++
	}
	if hasSpecial {
		strength++
	}

	if strength > 4 {
		strength = 4
	}
	return strength
}
