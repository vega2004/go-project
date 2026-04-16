package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type RecaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	Action      string    `json:"action"`
	Score       float64   `json:"score"`
}

func ValidateRecaptcha(token, secretKey string, minScore float64, expectedAction string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token de reCAPTCHA vacío")
	}

	if secretKey == "" {
		return true, nil
	}

	data := url.Values{}
	data.Set("secret", secretKey)
	data.Set("response", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://www.google.com/recaptcha/api/siteverify", data)
	if err != nil {
		return false, fmt.Errorf("error llamando API de reCAPTCHA: %w", err)
	}
	defer resp.Body.Close()

	var recaptchaResp RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&recaptchaResp); err != nil {
		return false, fmt.Errorf("error decodificando respuesta reCAPTCHA: %w", err)
	}

	// LOGS PARA DEPURAR
	fmt.Printf("[RECAPTCHA DEBUG] Success: %v\n", recaptchaResp.Success)
	fmt.Printf("[RECAPTCHA DEBUG] Hostname: %s\n", recaptchaResp.Hostname)
	fmt.Printf("[RECAPTCHA DEBUG] ErrorCodes: %v\n", recaptchaResp.ErrorCodes)
	fmt.Printf("[RECAPTCHA DEBUG] Score: %f\n", recaptchaResp.Score)

	if !recaptchaResp.Success {
		return false, fmt.Errorf("verificación reCAPTCHA fallida: %v", recaptchaResp.ErrorCodes)
	}

	if minScore > 0 {
		if recaptchaResp.Score < minScore {
			return false, fmt.Errorf("score demasiado bajo: %.2f (mínimo: %.2f)", recaptchaResp.Score, minScore)
		}
		if expectedAction != "" && recaptchaResp.Action != expectedAction {
			return false, fmt.Errorf("acción incorrecta: esperada '%s', obtenida '%s'", expectedAction, recaptchaResp.Action)
		}
	}

	return true, nil
}
