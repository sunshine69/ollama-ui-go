package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaRequest struct {
	// Prompt   string      `json:"prompt"`
	Model    string      `json:"model"`
	Stream   bool        `json:"stream"`
	Messages []AIMessage `json:"messages"`
	Images   []string    `json:"images"`
	Format   string      `json:"format"`
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
		OllamaURL = "http://192.168.20.49:11434" // example
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
