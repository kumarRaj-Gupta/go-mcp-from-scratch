package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// RequestPayload represents the incoming JSON-RPC request from STDIN.
// Example: {"jsonrpc": "2.0", "id": 123, "method": "execute", "params": {...}}
type RequestPayload struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"` // Use interface{} to handle null or string/number IDs
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// ExecutionParams represents the parameters for the 'execute' method.
// These are extracted from RequestPayload.Params when the method is 'execute'.
type ExecutionParams struct {
	ToolName  string   `json:"tool_name"`
	Arguments funcArgs `json:"arguments"` // funcArgs is defined in mcp_logic.go
}

// ResponsePayload represents the outgoing JSON-RPC response to STDOUT.
// Example: {"jsonrpc": "2.0", "id": 123, "result": {...}} or {"jsonrpc": "2.0", "id": 123, "error": {...}}
type ResponsePayload struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"` // Omit if empty
	Error   interface{} `json:"error,omitempty"`  // Omit if empty
}

// runStdioServer is the main loop for the STDIO server.
func runStdioServer() {
	// Create a new scanner to read line by line from standard input
	scanner := bufio.NewScanner(os.Stdin)

	// Loop indefinitely, processing one JSON message per line
	for scanner.Scan() {
		line := scanner.Bytes()

		// 1. Parse the incoming JSON request
		var request RequestPayload
		if err := json.Unmarshal(line, &request); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
			continue
		}

		responsePayload := make(map[string]interface{})

		// 2. Process the MCP methods
		switch request.Method {
		case "initialize":
			responsePayload = mcpInitialize()

		case "execute":
			var execParams ExecutionParams
			// Unmarshal the 'params' field into the specific ExecutionParams struct
			if err := json.Unmarshal(request.Params, &execParams); err != nil {
				responsePayload = map[string]interface{}{"error": "Invalid parameters structure for 'execute'"}
			} else {
				// Call the core execution logic
				responsePayload = mcpExecute(execParams.ToolName, execParams.Arguments)
			}
		default:
			// Handle unsupported methods
			responsePayload = map[string]interface{}{"error": fmt.Sprintf("Unsupported method: %s", request.Method)}
		}

		// 3. Send the final JSON-RPC response back via STDOUT
		if request.ID != nil {

			// Check if the response is an error or a result based on its content
			var finalResponse ResponsePayload
			finalResponse.Jsonrpc = "2.0"
			finalResponse.ID = request.ID

			if _, isError := responsePayload["error"]; isError {
				// If the payload contains an "error" key, treat it as a JSON-RPC error
				// Note: For simplicity, we are placing the whole error map under the Result for now,
				// but a proper JSON-RPC error would use the 'error' field in the ResponsePayload.
				// For MCP compatibility, we place the full responsePayload under Result.
				finalResponse.Result = responsePayload
			} else {
				finalResponse.Result = responsePayload
			}

			// Marshal the final JSON-RPC object
			jsonOutput, err := JsonOutput(finalResponse)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling response: %v\n", err)
				continue
			}

			// Write JSON output followed by a newline, and flush immediately
			fmt.Fprintf(os.Stdout, "%s\n", jsonOutput)
			os.Stdout.Sync()
		}

		fmt.Fprintf(os.Stderr, "-> Processed method: %s (ID: %v)\n", request.Method, request.ID)
	}

	// Check for scanner errors after the loop
	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error reading from standard input: %v\n", err)
	}
}

// main is the entry point for the STDIO server binary
func main() {
	fmt.Println("Starting MCP STDIO Server. Awaiting input on STDIN...")
	runStdioServer()
}
