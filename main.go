package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sunshine69/ollama-ui-go/lib"
)

var AcceptedUsers = map[string]string{}

func main() {
	path_base := os.Getenv("PATH_BASE")

	if err := json.Unmarshal([]byte(os.Getenv("ACCEPTED_USERS")), &AcceptedUsers); err != nil {
		secret, _ := lib.GenerateSecureRandomPassword(64)
		AcceptedUsers = map[string]string{"admin": secret}
		fmt.Printf("[INFO] No ACCEPTED_USERS environment variable found. Generate default credentials. User: admin, jwtsecret: '%s'\n", secret)
		fmt.Println(`[INFO] If you want to set your own then set env var 'ACCEPTED_USERS' with a json string in the format '{"your-user-name": "your-jwt-secret"}'. To login provide the username and the jwt token generated using the secret and the 'sub' field must be set to the username.`)
	}

	http.HandleFunc(path_base+"/ollama/model/", func(w http.ResponseWriter, r *http.Request) {
		modelName := r.URL.Path[len(path_base+"/ollama/model/"):]
		modelInfo, err := lib.GetOllamaModel(modelName)
		if err != nil {
			http.Error(w, "Failed to fetch model information", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(modelInfo)
	})

	http.HandleFunc(path_base+"/ollama/models", func(w http.ResponseWriter, r *http.Request) {

		models, err := lib.GetOllamaModels()
		if err != nil {
			http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(models)
	})

	http.HandleFunc(path_base+"/ollama/ask", func(w http.ResponseWriter, r *http.Request) {

		var ollamaRequest lib.OllamaRequest
		jsonData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		fmt.Println(string(jsonData))
		if err := json.Unmarshal(jsonData, &ollamaRequest); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		requestBody, err := json.Marshal(ollamaRequest)
		if err != nil {
			http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
			return
		}
		requestString := string(requestBody)
		response, err := lib.AskOllamaAPI(requestString)
		if err != nil {
			http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || !strings.HasPrefix(r.URL.Path, path_base+"/ollama") {
			http.StripPrefix("/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	http.ListenAndServe(":"+port, isAuthorized(http.DefaultServeMux))
}

func basicAuth(w http.ResponseWriter, r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	if secret, ok := AcceptedUsers[username]; ok {
		sub, err := lib.ValidateJWT(password, secret)
		if err != nil || username != sub {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return false
		}
	} else {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func isAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if basicAuth(w, r) {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	})
}
