// Build like this env CGO=0 go build -trimpath --tags "json1 fts5 secure_delete osusergo netgo sqlite_stat4 sqlite_foreign_keys" -ldflags="-X main.version=v1.0 -extldflags=-w -s" -buildmode=plugin ai-tools.go

package main

// Get_current_weather is a dummy function that returns a string.
// To make it simpler every function in the plugin should have the same signature.
// The function can take any number of arguments and return a string.
// In the UI set the Options field like this
// { "tools": [
//
//	{
//	  "type": "function",
//	  "function": {
//	    "name": "Get_current_weather",
//	    "description": "Get the current weather for a location",
//	    "parameters": {
//	      "type": "object",
//	      "properties": {
//	        "location": {
//	          "type": "string",
//	          "description": "The location to get the weather for, e.g. San Francisco, CA"
//	        },
//	        "format": {
//	          "type": "string",
//	          "description": "The format to return the weather in, e.g. 'celsius' or 'fahrenheit'",
//	          "enum": ["celsius", "fahrenheit"]
//	        }
//	      },
//	      "required": ["location", "format"]
//	    }
//	  }
//	}
//
// ]
// }
func Get_current_weather(myargs ...string) string {
	out := "Function Get_current_weather called with arguments: "
	for _, value := range myargs {
		out += " '" + value + "'"
	}
	return out
}
