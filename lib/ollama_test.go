package lib

import (
	"log"
	"testing"
)

func TestRunLuaFile(t *testing.T) {
	o, err := RunLuaFile("test.lua", []byte("this is input data"))
	log.Println("output: ", string(o))
	log.Println("error: ", err)
}
