package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const PORT = 8080

func sseExecuteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")
	w.Header().Add("Access-Control-Allowed-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is valid", http.StatusMethodNotAllowed)
		return
	}

	// Unmarshall the Request Body
	// executionParams
	var executionParams ExecutionParams

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &executionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponsePayload := mcpExecute(executionParams.ToolName, executionParams.Arguments)
	// Format the result as a single SSE event-stream
	// Expected Format: [data]\n\n
	jsonResult, err := JsonOutput(ResponsePayload)
	if err != nil {
		fmt.Printf("Error Marshalling result:%v", err)
		fmt.Fprintf(w, "data: %s\n\n", `{"error":"Internal Server Error Marshalling the Json Result"}`)
	}

	sseEvent := fmt.Sprintf("data: %s\n\n", jsonResult)

	_, err = w.Write([]byte(sseEvent))
	if err != nil {
		fmt.Printf("Error writing data to http.ResponseWriter: %v", err)
		fmt.Fprintf(w, "data: %s\n\n", `{"error":"Error writing data to http.ResponseWriter"}`)
	}

	// Now we must flush the buffer to make sure the output is immediately sent.
	if Flusher, ok := w.(http.Flusher); ok {
		Flusher.Flush()
	}

	fmt.Printf("Executed Tool %s(SSE)\n", executionParams.ToolName)

}

func toolsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the Tool Definitions
	ResponsePayload := mcpInitialize()
	// Convert to Json
	jsonBody, err := json.Marshal(ResponsePayload)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", `{"error":"Error Marshalling the Json"}`)
		fmt.Printf("Error Marshalling the Json: %v", err)
	}
	// Write the Json Response
	outputString := fmt.Sprintf("data: %s\n\n", jsonBody)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(outputString))
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", `{"error":"Error writing to http.ResponseWriter"}`)
		fmt.Printf("Error writing to http.ResponseWriter: %v", err)
	}
	fmt.Print("Tool Definitions sent.")
}

func RunSSEServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/tools", toolsHandler)
	mux.HandleFunc("/execute", sseExecuteHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", PORT),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 15,
	}

	fmt.Printf("Starting the MCP Server (SSE) at localhost:%v", PORT)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error Starting the Server: %v", err)
		return
	}
}
