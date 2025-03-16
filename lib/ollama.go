package lib

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cjoudrey/gluahttp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kohkimakimoto/gluayaml"
	"github.com/ollama/ollama/api"
	"github.com/sunshine69/gluare"

	u "github.com/sunshine69/golang-tools/utils"
	gopherjson "github.com/sunshine69/gopher-json"
	lua "github.com/yuin/gopher-lua"
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

func HandleOllamaGetModels(w http.ResponseWriter, r *http.Request) {
	models, err := GetOllamaModels()
	if err != nil {
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(models)
}

func HandleOllamaChat(w http.ResponseWriter, r *http.Request) {
	var ollamaRequest OllamaRequest
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
	// dur, err := time.ParseDuration(ollamaRequest.KeepAlive)
	// if err != nil {
	// 	http.Error(w, "Invalid keep alive duration", http.StatusBadRequest)
	// 	return
	// }
	// keep_alive := api.Duration{Duration: dur}

	req := &api.ChatRequest{
		Model:    ollamaRequest.Model,
		Messages: ollamaRequest.Messages,
		Stream:   &ollamaRequest.Stream,
		Options:  ollamaRequest.Options,
		Format:   json.RawMessage(ollamaRequest.Format),
		Tools:    ollamaRequest.Tools,
		// KeepAlive: &keep_alive,
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	respFunc := func(resp api.ChatResponse) error {
		// Not sure why resp.Message.ToolCalls isd always empty list. ollama bug?
		// The response Message Content is in the format <|tool_call|>content<|tool_call|> where content is a json which ahs the AI response. We need to parse this and make a decision on how to handle it.
		if len(resp.Message.ToolCalls) > 0 { // yeah some model support it. Might be ollama does not understand other model tags
			for _, toolCall := range resp.Message.ToolCalls {
				fmt.Fprintf(os.Stderr, "[DEBUG] func name: %s Args: %s\n", toolCall.Function.Name, toolCall.Function.Arguments.String())
				if _, err := os.Stat("lua-tools/" + toolCall.Function.Name + ".lua"); os.IsNotExist(err) {
					fmt.Println("Failed to lookup lua file", err)
					fmt.Fprint(w, resp.Message.Content)
				} else {
					jsonin := u.JsonDumpByte(toolCall.Function.Arguments, "")
					output, _ := RunLuaFile("lua-tools/"+toolCall.Function.Name+".lua", jsonin)
					fmt.Fprint(w, output)
				}
			}
		} else { // Non standard tools call - like gemma3; they give it in the response text
			if toolsFuncs, err := ParseToolCalls(resp.Message.Content); err == nil {
				for _, toolCall := range toolsFuncs {
					fmt.Fprintf(os.Stderr, "[DEBUG] func name: %s Args: %s\n", toolCall.Function.Name, toolCall.Function.Arguments)
					if _, err := os.Stat("lua-tools/" + toolCall.Function.Name + ".lua"); os.IsNotExist(err) {
						fmt.Println("Failed to lookup lua file", err)
						fmt.Fprint(w, resp.Message.Content)
					} else {
						jsonin := u.JsonDumpByte(toolCall.Function.Arguments, "")
						output, _ := RunLuaFile("lua-tools/"+toolCall.Function.Name+".lua", jsonin)
						fmt.Fprint(w, output)
					}
				}
			} else {
				// fmt.Fprintln(os.Stderr, "Failed to parse tool calls "+err.Error())
				_, err := fmt.Fprint(w, resp.Message.Content)
				if err != nil {
					cancel()
					return err
				}
			}
		}
		flusher.Flush()
		return nil
	}

	err = client.Chat(ctx1, req, respFunc)
	if err != nil {
		http.Error(w, "Failed to process chat request", http.StatusInternalServerError)
		return
	}
}

func HandleOllamaGetModel(w http.ResponseWriter, r *http.Request) {
	// path_base := os.Getenv("PATH_BASE")
	// modelName := r.URL.Path[len(path_base+"/ollama/model/"):]
	modelName := r.PathValue("model_name")
	modelInfo, err := GetOllamaModel(modelName)
	if err != nil {
		http.Error(w, "Failed to fetch model information", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(modelInfo)
}

func RunLuaFile(luaFileName string, inputData []byte) ([]byte, error) {
	old := os.Stdout     // keep backup of the real stdout
	oldStdin := os.Stdin // keep backup of the real stdin

	r, w, _ := os.Pipe()
	outC := make(chan []byte)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.Bytes()
	}()
	os.Stdout = w

	inR, inW, _ := os.Pipe()
	os.Stdin = inR

	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("re", gluare.Loader)
	L.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
	L.PreloadModule("yaml", gluayaml.Loader)
	L.PreloadModule("json", gopherjson.Loader)

	// Write input data to stdin
	byteCount, err := inW.Write(inputData)
	if err != nil {
		os.Stdout = old
		os.Stdin = oldStdin
		fmt.Println("Failed to write to stdin", err.Error())
		return nil, err
	}
	inW.Close()

	err = L.DoFile(luaFileName)
	if err != nil {
		return nil, err
	}

	w.Close()
	os.Stdout = old
	os.Stdin = oldStdin
	fmt.Println("byteCount: ", byteCount)
	out := <-outC
	return out, nil
}
