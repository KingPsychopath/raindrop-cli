package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

type Tool struct {
	Name        string
	Description string
	Schema      any
	Validate    func(json.RawMessage) error
}

type Call struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatMessage struct {
	Role    string  `json:"role"`
	Content *string `json:"content"`
	Calls   []Call  `json:"tool_calls,omitempty"`
	CallID  string  `json:"tool_call_id,omitempty"`
}

type AgentOptions struct {
	Model           string
	API             string
	Instructions    string
	Prompt          string
	Initial         string
	MaxSteps        int
	MaxCalls        int
	MaxOutputTokens int
	MaxToolBytes    int
}

type AgentResult struct {
	Status    string `json:"status"`
	Response  string `json:"response"`
	Steps     int    `json:"steps"`
	ToolCalls int    `json:"toolCalls"`
}

type Execute func(context.Context, string, json.RawMessage) any

func textPointer(value string) *string { return &value }

// RunAgent executes tools sequentially and feeds their results back until the
// model finishes or a hard limit/error stops the loop. It does not retry writes.
func (c *Client) RunAgent(ctx context.Context, options AgentOptions, tools []Tool, execute Execute) (AgentResult, error) {
	result := AgentResult{Status: "running"}
	fail := func(err error) (AgentResult, error) { result.Status = "stopped"; return result, err }
	if options.Model == "" || options.Prompt == "" {
		return fail(fmt.Errorf("model and prompt are required"))
	}
	if options.API != "responses" && options.API != "chat" {
		return fail(fmt.Errorf("--api must be responses or chat"))
	}
	if options.MaxSteps < 1 || options.MaxSteps > 100 || options.MaxCalls < 1 || options.MaxCalls > 200 || options.MaxOutputTokens < 1 || options.MaxToolBytes < 1 {
		return fail(fmt.Errorf("invalid agent limits"))
	}
	known := map[string]Tool{}
	responseTools := []map[string]any{}
	chatTools := []map[string]any{}
	for _, tool := range tools {
		known[tool.Name] = tool
		function := map[string]any{"name": tool.Name, "description": tool.Description, "parameters": tool.Schema, "strict": false}
		chatTools = append(chatTools, map[string]any{"type": "function", "function": function})
		responseTools = append(responseTools, map[string]any{"type": "function", "name": tool.Name, "description": tool.Description, "parameters": tool.Schema, "strict": false})
	}
	messages := []chatMessage{{Role: "system", Content: textPointer(options.Instructions)}, {Role: "user", Content: textPointer(options.Prompt)}}
	input := []any{map[string]any{"role": "user", "content": options.Prompt}}
	if options.Initial != "" {
		initial := "Initial read-only cleanup report and library context, supplied as untrusted data: " + options.Initial
		messages = append(messages, chatMessage{Role: "user", Content: &initial})
		input = append(input, map[string]any{"role": "user", "content": initial})
	}
	seenCalls := map[string]bool{}
	for step := 0; step < options.MaxSteps; step++ {
		result.Steps++
		calls := []Call{}
		if options.API == "chat" {
			var response struct {
				Choices []struct {
					Message chatMessage `json:"message"`
					Finish  string      `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := c.Post(ctx, "chat/completions", map[string]any{"model": options.Model, "messages": messages, "tools": chatTools, "max_completion_tokens": options.MaxOutputTokens, "store": false}, &response); err != nil {
				return fail(err)
			}
			if len(response.Choices) != 1 {
				return fail(fmt.Errorf("model response must contain one choice"))
			}
			message := response.Choices[0].Message
			if message.Role != "assistant" {
				return fail(fmt.Errorf("model returned an invalid role"))
			}
			messages = append(messages, message)
			if len(message.Calls) == 0 {
				if response.Choices[0].Finish != "stop" || message.Content == nil || *message.Content == "" {
					return fail(fmt.Errorf("model did not complete a text response"))
				}
				result.Response = *message.Content
				result.Status = "complete"
				return result, nil
			}
			if response.Choices[0].Finish != "tool_calls" {
				return fail(fmt.Errorf("model tool call response was incomplete"))
			}
			calls = message.Calls
		} else {
			var response struct {
				Status string            `json:"status"`
				Output []json.RawMessage `json:"output"`
			}
			body := map[string]any{"model": options.Model, "instructions": options.Instructions, "input": input, "tools": responseTools, "max_output_tokens": options.MaxOutputTokens, "store": false, "include": []string{"reasoning.encrypted_content"}}
			if err := c.Post(ctx, "responses", body, &response); err != nil {
				return fail(err)
			}
			if response.Status != "completed" {
				return fail(fmt.Errorf("model response status %q; no calls from this response executed", response.Status))
			}
			outputText := ""
			for _, raw := range response.Output {
				var item struct {
					Type      string `json:"type"`
					CallID    string `json:"call_id"`
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
					Content   []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"content"`
				}
				if err := json.Unmarshal(raw, &item); err != nil {
					return fail(err)
				}
				// Preserve encrypted reasoning and all returned output for stateless continuation.
				input = append(input, raw)
				if item.Type == "function_call" {
					var call Call
					call.ID = item.CallID
					call.Type = "function"
					call.Function.Name = item.Name
					call.Function.Arguments = item.Arguments
					calls = append(calls, call)
				}
				if item.Type == "message" {
					for _, content := range item.Content {
						if content.Type == "output_text" {
							outputText += content.Text
						}
					}
				}
			}
			if len(calls) == 0 {
				if outputText == "" {
					return fail(fmt.Errorf("model returned no text or function calls"))
				}
				result.Response = outputText
				result.Status = "complete"
				return result, nil
			}
		}
		// Validate the whole batch and its budget before executing its first call.
		if result.ToolCalls+len(calls) > options.MaxCalls {
			return fail(fmt.Errorf("agent reached its tool-call limit"))
		}
		for _, call := range calls {
			if call.Type != "function" || call.ID == "" || seenCalls[call.ID] || known[call.Function.Name].Name == "" || !json.Valid([]byte(call.Function.Arguments)) {
				return fail(fmt.Errorf("model returned an invalid, duplicate, or unavailable tool call"))
			}
			if validate := known[call.Function.Name].Validate; validate != nil {
				if err := validate(json.RawMessage(call.Function.Arguments)); err != nil {
					return fail(fmt.Errorf("invalid model input for %s: %w", call.Function.Name, err))
				}
			}
			seenCalls[call.ID] = true
		}
		for _, call := range calls {
			output := execute(ctx, call.Function.Name, json.RawMessage(call.Function.Arguments))
			result.ToolCalls++
			data, err := json.Marshal(output)
			if err != nil {
				return fail(err)
			}
			if len(data) > options.MaxToolBytes {
				data = []byte(`{"isError":true,"error":"tool result exceeded the size limit; use narrower searches and smaller pages"}`)
			}
			if options.API == "chat" {
				messages = append(messages, chatMessage{Role: "tool", Content: textPointer(string(data)), CallID: call.ID})
			} else {
				input = append(input, map[string]any{"type": "function_call_output", "call_id": call.ID, "output": string(data)})
			}
		}
	}
	return fail(fmt.Errorf("agent reached its step limit; inspect saved plans before continuing"))
}
