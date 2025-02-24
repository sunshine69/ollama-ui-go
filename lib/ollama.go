package lib

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaRequest struct {
	// Prompt   string      `json:"prompt"`
	Model    string                 `json:"model"`
	Stream   bool                   `json:"stream"`
	Messages []AIMessage            `json:"messages"`
	Images   []string               `json:"images"`
	Format   string                 `json:"format"`
	Options  map[string]interface{} `json:"options"`
}

var (
	OllamaURL             string
	ollamaAPIChatEndpoint string
	ollamaTagEndpoint     string
)

func init() {
	if OllamaURL == "" {
		OllamaURL = os.Getenv("OLLAMA_URL")
	}
	if OllamaURL == "" {
		OllamaURL = "http://localhost:11434" // example
	}
	parseOllamaEndpoint()
}

func parseOllamaEndpoint() {
	ollamaAPIChatEndpoint = OllamaURL + "/api/chat"
	ollamaTagEndpoint = OllamaURL + "/api/tags"
}

func AskOllamaAPI(question string) ([]byte, error) {
	// Create a POST request to the Ollama API
	req, err := http.NewRequest("POST", ollamaAPIChatEndpoint, strings.NewReader(question))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	// Return the response as a string
	return body, nil
}

func GetOllamaModels() ([]byte, error) {
	req, err := http.NewRequest("GET", ollamaTagEndpoint, nil)
	if err != nil {
		return []byte(""), err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}
	return body, nil
}

func GetOllamaModel(modelName string) ([]byte, error) {
	fmt.Println("[DEBUG] modelName: " + modelName)
	payload := fmt.Sprintf(`{"model": "%s"}`, modelName)
	req, err := http.NewRequest("POST", OllamaURL+"/api/show", strings.NewReader(payload))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}
	return body, nil
}

func GenerateSecureRandomPassword(length int) (string, error) {
	if length < 12 {
		return "", fmt.Errorf("password length must be at least 12 characters")
	}

	const lettersAndDigits = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%&*"
	password := make([]byte, length)

	_, err := rand.Read(password)
	if err != nil {
		return "", fmt.Errorf("error generating password: %v", err)
	}

	for i := range password {
		password[i] = lettersAndDigits[int(password[i])%len(lettersAndDigits)]
	}

	return string(password), nil
}

func ValidateJWT(jwtToken, storedPasswordHash string) (string, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the key for validation
		return []byte(storedPasswordHash), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string), nil
	}
	return "", fmt.Errorf("Invalid password")
}
