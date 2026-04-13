package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	RoleID   int    `json:"role_id"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey  []byte
	expiration time.Duration
}

func NewJWTManager(secretKey string, expirationHours int) *JWTManager {
	return &JWTManager{
		secretKey:  []byte(secretKey),
		expiration: time.Duration(expirationHours) * time.Hour,
	}
}

// Generate - Genera un nuevo token JWT
func (m *JWTManager) Generate(userID int, email, name string, roleID int, roleName string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		Name:     name,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tu-proyecto",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Validate - Valida y extrae claims del token
func (m *JWTManager) Validate(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}

// Refresh - Genera un nuevo token a partir de uno existente (si no expiró)
func (m *JWTManager) Refresh(tokenString string) (string, error) {
	claims, err := m.Validate(tokenString)
	if err != nil {
		return "", err
	}

	// Generar nuevo token con los mismos claims
	return m.Generate(claims.UserID, claims.Email, claims.Name, claims.RoleID, claims.RoleName)
}

// ExtractFromRequest - Extrae token del header Authorization
func (m *JWTManager) ExtractFromRequest(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no se encontró header Authorization")
	}

	// Formato: "Bearer <token>"
	var tokenString string
	_, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenString)
	if err != nil {
		return "", fmt.Errorf("formato de Authorization inválido")
	}

	return tokenString, nil
}
