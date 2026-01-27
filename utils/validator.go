package utils

import (
	"html"
	"regexp"
	"strings"
)

func ValidateName(name string) bool {
	// CORREGIDO: Solo letras (incluyendo acentos españoles), espacios, apóstrofes y guiones
	re := regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ\s'-]{2,50}$`)
	return re.MatchString(name)
}

func ValidateEmail(email string) bool {
	// Expresión regular simple para email
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func ValidatePhone(phone string) bool {
	// CORREGIDO: Solo números, +, - y espacios
	re := regexp.MustCompile(`^[0-9+\-\s]{8,15}$`)
	if !re.MatchString(phone) {
		return false
	}

	// Contar solo dígitos (ignorar +, - y espacios)
	digitCount := 0
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}

	// Verificar que tenga entre 8 y 15 dígitos
	return digitCount >= 8 && digitCount <= 15
}

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	// Remover múltiples espacios
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	// Escapar caracteres HTML (MEJOR que QueryEscape)
	input = html.EscapeString(input)
	return input
}
