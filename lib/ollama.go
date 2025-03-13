package lib

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ollama/ollama/api"
)

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaRequest struct {
	// Prompt   string      `json:"prompt"`
	Model     string                 `json:"model"`
	Stream    bool                   `json:"stream"`
	Messages  []api.Message          `json:"messages"`
	Format    string                 `json:"format"`
	Options   map[string]interface{} `json:"options"`
	KeepAlive string                 `json:"keep_alive"`
	Raw       bool                   `json:"raw"`
	Tools     api.Tools              `json:"tools"`
}

var (
	OllamaHost            string
	ollamaAPIChatEndpoint string
	ollamaTagEndpoint     string
)

func init() {
	if OllamaHost == "" {
		OllamaHost = os.Getenv("OLLAMA_HOST")
	}
	if OllamaHost == "" {
		OllamaHost = "http://localhost:11434" // example
	}
	parseOllamaEndpoint()
}

func parseOllamaEndpoint() {
	ollamaAPIChatEndpoint = OllamaHost + "/api/chat"
	ollamaTagEndpoint = OllamaHost + "/api/tags"
}

func AskOllamaAPI(question string) (*http.Response, error) {
	// Create a POST request to the Ollama API
	req, err := http.NewRequest("POST", ollamaAPIChatEndpoint, strings.NewReader(question))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// defer resp.Body.Close()

	// Read the response body
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// Return the response as a string
	return resp, nil
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
	req, err := http.NewRequest("POST", OllamaHost+"/api/show", strings.NewReader(payload))
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

type ToolFunctionResponse struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string   `json:"type"`
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type        string   `json:"type"`
				Description string   `json:"description"`
				Enum        []string `json:"enum,omitempty"`
			} `json:"properties"`
		} `json:"parameters"`
		Arguments map[string]interface{} `json:"arguments"`
	} `json:"function"`
}

func ParseToolCalls(inputString string) (toolfFunctions []ToolFunctionResponse, err error) {
	// Find the start and end indices of the string_data part.
	start := strings.Index(inputString, "<|tool_call|>")
	if start == -1 {
		return toolfFunctions, fmt.Errorf("tool_call tag not found")
	}
	end := strings.Index(inputString, "<|/tool_call|>")
	if end == -1 {
		return toolfFunctions, fmt.Errorf("/tool_call/ tag not found")
	}

	// Extract the string_data part.
	data := inputString[start+len("<|tool_call|>") : end]
	fmt.Fprintf(os.Stderr, "[DEBUG] data: %s\n", data)
	err = json.Unmarshal([]byte(data), &toolfFunctions)
	return toolfFunctions, err
}

func FlattenArgument(arugments map[string]any) []string {
	var args []string
	for _, v := range arugments {
		args = append(args, v.(string))
	}
	return args
}
