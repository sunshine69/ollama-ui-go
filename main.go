package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sunshine69/ollama-ui-go/lib"
)

func main() {
	r := gin.Default()

	AcceptedUsers := map[string]string{}

	if err := json.Unmarshal([]byte(os.Getenv("ACCEPTED_USERS")), &AcceptedUsers); err != nil {
		secret, _ := lib.GenerateSecureRandomPassword(64)
		AcceptedUsers = map[string]string{"admin": secret}
		fmt.Printf(`[INFO] No ACCEPTED_USERS environment variable found. Generate default credentials. User: admin, jwtsecret: %s.`, secret)
		fmt.Printf(`[INFO] If you want to set your own then set env var 'ACCEPTED_USERS' with a json string in the format '{"your-user-name": "your-jwt-secret"}'. To login provide the username and the jwt token generated using the secret and the 'sub' field must be set to the username.`)
	}

	// Middleware for basic authentication
	r.Use(func(c *gin.Context) {
		doAbort := func() {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
		username, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			doAbort()
			return
		}
		// Check if provided credentials match the stored ones
		if secret, ok := AcceptedUsers[username]; ok {
			sub, err := lib.ValidateJWT(password, secret)
			if err != nil || username != sub {
				doAbort()
				return
			}
		} else {
			doAbort()
			return
		}
		c.Next()
	})

	r.GET("/model/:modelname", func(c *gin.Context) {
		modelName := c.Param("modelname")
		modelInfo, err := lib.GetOllamaModel(modelName)
		if err != nil {
			println("[DEBUG] [ERROR]: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch model information"})
			return
		}
		c.Data(http.StatusOK, "application/json", modelInfo)
	})

	r.GET("/models", func(c *gin.Context) {
		models, err := lib.GetOllamaModels()
		if err != nil {
			println("[DEBUG] [ERROR]: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
			return
		}
		c.Data(http.StatusOK, "application/json", models)
	})

	r.POST("/ask", func(c *gin.Context) {
		var ollamaRequest lib.OllamaRequest
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
		// fmt.Println("[DEBUG] requestString " + requestString)
		response, err := lib.AskOllamaAPI(requestString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
			return
		}
		// fmt.Println("[DEBUG] AI response " + string(response))
		c.Data(http.StatusOK, "application/json", response)
	})
	r.Static("static/", "static")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	r.Run(":8081")
}
