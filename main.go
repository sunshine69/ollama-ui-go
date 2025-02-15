package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	// To bundle assets first build the binary - then get into this dir (where the go file has the rice findbox command) and run 'rice append --exec <path-to-bin>
	rice "github.com/GeertJohan/go.rice"
	"github.com/ollama/ollama/api"
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

		client, err := api.ClientFromEnvironment()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		req := &api.ChatRequest{
			Model:    ollamaRequest.Model,
			Messages: ollamaRequest.Messages,
			Stream:   &ollamaRequest.Stream,
			Options:  ollamaRequest.Options,
			Format:   json.RawMessage(ollamaRequest.Format),
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		respFunc := func(resp api.ChatResponse) error {
			// fmt.Print(resp.Message.Content)
			fmt.Fprint(w, resp.Message.Content)
			flusher.Flush()
			return nil
		}

		err = client.Chat(ctx, req, respFunc)
		if err != nil {
			http.Error(w, "Failed to process chat request", http.StatusInternalServerError)
			return
		}
	})

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	if r.URL.Path == "/" || !strings.HasPrefix(r.URL.Path, path_base+"/ollama") {
	// 		http.StripPrefix("/", http.FileServer(rice.MustFindBox("static").HTTPBox())).ServeHTTP(w, r)
	// 	} else {
	// 		http.NotFound(w, r)
	// 	}
	// })
	t := template.New("tmpl")
	templateBox := rice.MustFindBox("templates")
	templateBox.Walk("/", func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		if info.IsDir() {
			return nil
		}
		fname := filepath.Base(path)
		t = template.Must(t.New(fname).Parse(templateBox.MustString(fname)))
		return nil
	})
	http.HandleFunc(path_base+"/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path_base+"/" {
			t.ExecuteTemplate(w, "index.html", map[string]any{"path_base": path_base})
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
