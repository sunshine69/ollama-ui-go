module github.com/sunshine69/ollama-ui-go

go 1.24.0

replace github.com/sunshine69/ollama-ui-go/lib => ./lib

require (
	github.com/GeertJohan/go.rice v1.0.3
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/ollama/ollama v0.6.1
)

require github.com/daaku/go.zipexe v1.0.2 // indirect
