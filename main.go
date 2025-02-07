package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
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
}

var (
	ollamaURL             string
	ollamaAPIChatEndpoint string
	ollamaTagEndpoint     string
)

func init() {
	ollamaURL = os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://192.168.20.49:11434" // example
		ollamaAPIChatEndpoint = ollamaURL + "/api/chat"
		ollamaTagEndpoint = ollamaURL + "/api/tags"
	}
}

func main() {
	r := gin.Default()

	r.GET("/models", func(c *gin.Context) {
		models, err := getOllamaModels()
		if err != nil {
			println("[DEBUG] [ERROR]: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
			return
		}
		c.Data(http.StatusOK, "application/json", []byte(models))
	})

	r.POST("/ask", func(c *gin.Context) {
		var ollamaRequest OllamaRequest
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// Handle error
		}
		// fmt.Println(string(jsonData))
		if err := json.Unmarshal(jsonData, &ollamaRequest); err != nil {
			fmt.Printf("[DEBUG] Error: %s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		requestBody, err := json.Marshal(ollamaRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
			return
		}
		requestString := string(requestBody)
		fmt.Println("[DEBUG] requestString " + requestString)
		response, err := askOllamaAPI(requestString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
			return
		}
		fmt.Println("[DEBUG] AI response " + response)
		c.JSON(http.StatusOK, gin.H{"response": response})
	})
	r.Static("static/", "static")
	r.Run(":8080")
}

func askOllamaAPI(question string) (string, error) {
	// Create a POST request to the Ollama API
	req, err := http.NewRequest("POST", ollamaAPIChatEndpoint, strings.NewReader(question))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Return the response as a string
	return string(body), nil
}

func getOllamaModels() (string, error) {
	req, err := http.NewRequest("GET", ollamaTagEndpoint, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
