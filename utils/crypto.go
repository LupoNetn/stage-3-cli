package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateRandomString generates a URL-safe random string
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GeneratePKCEParams creates the state, verifier, and challenge for OAuth2 PKCE
func GeneratePKCEParams(stateInt, verifierInt int) (string, string, string, error) {
	state, err := GenerateRandomString(stateInt)
	if err != nil {
		return "", "", "", err
	}

	codeVerifier, err := GenerateRandomString(verifierInt)
	if err != nil {
		return "", "", "", err
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return state, codeVerifier, codeChallenge, nil
}
