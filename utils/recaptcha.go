package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// Tus claves de reCAPTCHA
	recaptchaSecret = "6LdjX1gsAAAAAF72XunM29l-VWjQvnB0UheAhCc6"
	recaptchaURL    = "https://www.google.com/recaptcha/api/siteverify"
)

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func ValidateRecaptcha(token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	// Preparar datos para la petición
	data := url.Values{}
	data.Set("secret", recaptchaSecret)
	data.Set("response", token)

	// Hacer petición POST
	resp, err := http.PostForm(recaptchaURL, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Parsear respuesta JSON
	var recaptchaResp RecaptchaResponse
	err = json.Unmarshal(body, &recaptchaResp)
	if err != nil {
		return false, err
	}

	return recaptchaResp.Success, nil
}
