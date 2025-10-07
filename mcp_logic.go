package main

import (
	"encoding/json"
	"fmt"
)

// type for function's arguments

type funcArgs map[string]float64

func add(argMap funcArgs) interface{} {
	n1, n1Exists := argMap["a"]
	n2, n2Exists := argMap["b"]
	if !n1Exists || !n2Exists {
		return map[string]interface{}{"error": fmt.Sprintf("Either n1 or n2 is not provided. n1:%v n2%v", n1, n2)}
	}
	return n1 + n2
}
func sub(argMap funcArgs) interface{} {
	n1, n1Exists := argMap["a"]
	n2, n2Exists := argMap["b"]
	if !n1Exists || !n2Exists {
		return map[string]interface{}{"error": fmt.Sprintf("Either n1 or n2 is not provided. n1:%v n2%v", n1, n2)}
	}
	return n1 - n2
}
func mul(argMap funcArgs) interface{} {
	n1, n1Exists := argMap["a"]
	n2, n2Exists := argMap["b"]
	if !n1Exists || !n2Exists {
		return map[string]interface{}{"error": fmt.Sprintf("Either n1 or n2 is not provided. n1:%v n2%v", n1, n2)}
	}
	return n1 * n2
}
func div(argMap funcArgs) interface{} {
	n1, n1Exists := argMap["a"]
	n2, n2Exists := argMap["b"]
	if !n1Exists || !n2Exists {
		return map[string]interface{}{"error": fmt.Sprintf("Either n1 or n2 is not provided. n1:%v n2%v", n1, n2)}
	}
	if n2 == 0 {
		return map[string]interface{}{"error": "Divisior cannot be zero"}
	}
	return n1 / n2
}

type funcMCP func(funcArgs) interface{}

var funcMap = map[string]funcMCP{
	"add": add,
	"sub": sub,
	"mul": mul,
	"div": div,
}

// TOOL Definitions

type Tool_defs struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	INPUT_SCHEMA InputSchema `json:"input_schema"`
}
type Parameters struct {
	Type string `json:"type"`
	Desc string `json:"description"`
}
type InputSchema struct {
	Type       string                `json:"type"`
	Parameters map[string]Parameters `json:"parameters"`
	Required   []string              `json:"required"`
}

// Varies for different MCP Cients
var TOOL_DEFNITIONS = []Tool_defs{
	{
		Name:        "add",
		Description: "Adds two numbers together and returns result",
		INPUT_SCHEMA: InputSchema{
			Type: "OBJECT",
			Parameters: map[string]Parameters{
				"a": {Type: "number", Desc: "First Number"},
				"b": {Type: "number", Desc: "Second Number"},
			},
			Required: []string{"a", "b"},
		},
	},
	{
		Name:        "sub",
		Description: "Subtracts second number from the first and returns result",
		INPUT_SCHEMA: InputSchema{
			Type: "OBJECT",
			Parameters: map[string]Parameters{
				"a": {Type: "number", Desc: "First Number"},
				"b": {Type: "number", Desc: "Second Number"},
			},
			Required: []string{"a", "b"},
		},
	},
	{
		Name:        "mul",
		Description: "Multiplies two numbers together and returns result",
		INPUT_SCHEMA: InputSchema{
			Type: "OBJECT",
			Parameters: map[string]Parameters{
				"a": {Type: "number", Desc: "First Number"},
				"b": {Type: "number", Desc: "Second Number"},
			},
			Required: []string{"a", "b"},
		},
	},
	{
		Name:        "div",
		Description: "Divides second number from the first and returns result",
		INPUT_SCHEMA: InputSchema{
			Type: "OBJECT",
			Parameters: map[string]Parameters{
				"a": {Type: "number", Desc: "First Number"},
				"b": {Type: "number", Desc: "Second Number"},
			},
			Required: []string{"a", "b"},
		},
	},
}

func JsonOutput(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("couldn't process data into json. data:%v", data)
	}
	return string(jsonData), nil
}

func mcpInitialize() map[string]interface{} {
	return map[string]interface{}{"capabilites": map[string]interface{}{"tools": TOOL_DEFNITIONS}}
}

func mcpExecute(tool string, args map[string]float64) map[string]interface{} {
	toolToCall, found := funcMap[tool]
	if !found {
		return map[string]interface{}{"error": "No such tools. Please recheck. "}
	}
	returnValue := toolToCall(args)

	if err, ok := returnValue.(map[string]interface{}); ok {
		if _, keyErrorExists := err["error"]; keyErrorExists {
			return map[string]interface{}{"error": err["error"]}
		}
	}

	return map[string]interface{}{"result": map[string]interface{}{"value": returnValue}}
}
