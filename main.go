package main

// build like this env CGO=0 go build -trimpath --tags "json1 fts5 secure_delete osusergo netgo sqlite_stat4 sqlite_foreign_keys" -ldflags="-X main.version=v1.0 -extldflags=-w -s"

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	// To bundle assets first build the binary - then get into this dir (where the go file has the rice findbox command) and run 'rice append --exec <path-to-bin>
	rice "github.com/GeertJohan/go.rice"
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

	http.HandleFunc(path_base+"/ollama/model/{model_name}", lib.HandleOllamaGetModel)

	http.HandleFunc(path_base+"/ollama/models", lib.HandleOllamaGetModels)

	http.HandleFunc(path_base+"/ollama/ask", lib.HandleOllamaChat)
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
		preferred_models := os.Getenv("PREFERRED_MODELS")
		if preferred_models == "" {
			preferred_models = `["qwen2.5-coder:32b", "gemma3-12b:custom", "qwq:32b-q4_K_M", "huihui_ai/qwen2.5-coder-abliterate:14b-instruct-q4_K_M", "huihui_ai/phi4-abliterated:latest"]`
		}
		if r.URL.Path == path_base+"/" {
			t.ExecuteTemplate(w, "index.html", map[string]any{"path_base": path_base, "preferred_models": preferred_models})
		} else {
			http.NotFound(w, r)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Listening on port %s\n", port)
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
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	})
}
