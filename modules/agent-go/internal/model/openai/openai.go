package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// Model implements model.LLM interface for OpenAI
type Model struct {
	client    openai.Client
	modelName string
}

// NewModel creates a new OpenAI model adapter
func NewModel(apiKey string, modelName string) (*Model, error) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Model{
		client:    client,
		modelName: modelName,
	}, nil
}

// Name returns the model name
func (m *Model) Name() string {
	return m.modelName
}

// GenerateContent implements model.LLM interface
func (m *Model) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		// Convert genai messages to OpenAI format
		messages := m.convertMessages(req)

		// Convert tools from req.Config.Tools (genai format)
		var tools []openai.ChatCompletionToolParam
		if req.Config != nil && len(req.Config.Tools) > 0 {
			tools = m.convertGenaiTools(req.Config.Tools)
			log.Printf("🔧 OpenAI: Converted %d tools", len(tools))
			for _, t := range tools {
				log.Printf("   - Tool: %s", t.Function.Name)
			}
		}

		// Build request params
		params := openai.ChatCompletionNewParams{
			Model:    m.modelName,
			Messages: messages,
		}
		if len(tools) > 0 {
			params.Tools = tools
		}

		// Call OpenAI
		resp, err := m.client.Chat.Completions.New(ctx, params)
		if err != nil {
			yield(nil, fmt.Errorf("OpenAI API error: %w", err))
			return
		}

		// Convert response to ADK format
		llmResp := m.convertResponse(resp)
		yield(llmResp, nil)
	}
}

// convertMessages converts genai.Content to OpenAI messages
func (m *Model) convertMessages(req *model.LLMRequest) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion

	// Add system instruction if present
	if req.Config != nil && req.Config.SystemInstruction != nil {
		sysText := extractText(req.Config.SystemInstruction)
		if sysText != "" {
			messages = append(messages, openai.SystemMessage(sysText))
		}
	}

	// Add conversation messages
	for _, content := range req.Contents {
		// Check for function calls/responses first
		for _, part := range content.Parts {
			if part.FunctionCall != nil {
				// This is a tool call from assistant
				argsJSON, _ := json.Marshal(part.FunctionCall.Args)
				toolCallID := part.FunctionCall.ID
				if toolCallID == "" {
					toolCallID = "call_" + part.FunctionCall.Name
				}
				messages = append(messages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: []openai.ChatCompletionMessageToolCallParam{
							{
								ID:   toolCallID,
								Type: "function",
								Function: openai.ChatCompletionMessageToolCallFunctionParam{
									Name:      part.FunctionCall.Name,
									Arguments: string(argsJSON),
								},
							},
						},
					},
				})
				continue
			}
			if part.FunctionResponse != nil {
				// This is a tool response
				respJSON, _ := json.Marshal(part.FunctionResponse.Response)
				toolCallID := part.FunctionResponse.ID
				if toolCallID == "" {
					toolCallID = "call_" + part.FunctionResponse.Name
				}
				messages = append(messages, openai.ToolMessage(string(respJSON), toolCallID))
				continue
			}
		}

		// Regular text message
		text := extractText(content)
		if text != "" {
			switch content.Role {
			case "user":
				messages = append(messages, openai.UserMessage(text))
			case "model", "assistant":
				messages = append(messages, openai.AssistantMessage(text))
			}
		}
	}

	return messages
}

// convertGenaiTools converts genai.Tool definitions to OpenAI format
func (m *Model) convertGenaiTools(genaiTools []*genai.Tool) []openai.ChatCompletionToolParam {
	var result []openai.ChatCompletionToolParam

	for _, tool := range genaiTools {
		if tool == nil {
			continue
		}
		for _, fn := range tool.FunctionDeclarations {
			if fn == nil {
				continue
			}

			// Convert parameters - ADK uses ParametersJsonSchema (any), not Parameters (*genai.Schema)
			paramsSchema := shared.FunctionParameters{}
			if fn.ParametersJsonSchema != nil {
				// ParametersJsonSchema can be map[string]any or *jsonschema.Schema
				// Convert to map by JSON marshaling/unmarshaling
				jsonBytes, err := json.Marshal(fn.ParametersJsonSchema)
				if err == nil {
					var schemaMap map[string]any
					if err := json.Unmarshal(jsonBytes, &schemaMap); err == nil {
						for k, v := range schemaMap {
							paramsSchema[k] = v
						}
						log.Printf("   Tool %s schema: %v", fn.Name, paramsSchema)
					} else {
						log.Printf("   Tool %s: failed to unmarshal schema: %v", fn.Name, err)
					}
				} else {
					log.Printf("   Tool %s: failed to marshal schema: %v", fn.Name, err)
				}
			} else if fn.Parameters != nil {
				// Fallback to Parameters if set
				paramsSchema = m.convertSchemaToParams(fn.Parameters)
			}

			result = append(result, openai.ChatCompletionToolParam{
				Function: shared.FunctionDefinitionParam{
					Name:        fn.Name,
					Description: openai.String(fn.Description),
					Parameters:  paramsSchema,
				},
			})
		}
	}

	return result
}

// convertSchemaToParams converts genai.Schema to OpenAI FunctionParameters
func (m *Model) convertSchemaToParams(schema *genai.Schema) shared.FunctionParameters {
	params := shared.FunctionParameters{
		"type": string(schema.Type),
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]any)
		for name, prop := range schema.Properties {
			propDef := map[string]any{
				"type": string(prop.Type),
			}
			if prop.Description != "" {
				propDef["description"] = prop.Description
			}
			props[name] = propDef
			log.Printf("   Schema prop: %s (%s) - %s", name, prop.Type, prop.Description)
		}
		params["properties"] = props
	}

	if len(schema.Required) > 0 {
		params["required"] = schema.Required
		log.Printf("   Required: %v", schema.Required)
	}

	return params
}

// convertResponse converts OpenAI response to ADK LLMResponse
func (m *Model) convertResponse(resp *openai.ChatCompletion) *model.LLMResponse {
	if len(resp.Choices) == 0 {
		return &model.LLMResponse{
			Content: genai.NewContentFromText("", "model"),
		}
	}

	choice := resp.Choices[0]
	var parts []*genai.Part

	// Handle text content
	if choice.Message.Content != "" {
		parts = append(parts, &genai.Part{Text: choice.Message.Content})
	}

	// Handle tool calls
	for _, toolCall := range choice.Message.ToolCalls {
		var args map[string]any
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

		log.Printf("🔧 OpenAI tool call: %s with args: %v", toolCall.Function.Name, args)

		parts = append(parts, &genai.Part{
			FunctionCall: &genai.FunctionCall{
				ID:   toolCall.ID,
				Name: toolCall.Function.Name,
				Args: args,
			},
		})
	}

	// Convert finish reason
	var finishReason genai.FinishReason
	switch choice.FinishReason {
	case "stop":
		finishReason = genai.FinishReasonStop
	case "tool_calls":
		finishReason = genai.FinishReasonOther
	case "length":
		finishReason = genai.FinishReasonMaxTokens
	}

	return &model.LLMResponse{
		Content: &genai.Content{
			Role:  "model",
			Parts: parts,
		},
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(resp.Usage.PromptTokens),
			CandidatesTokenCount: int32(resp.Usage.CompletionTokens),
			TotalTokenCount:      int32(resp.Usage.TotalTokens),
		},
		TurnComplete: true,
		FinishReason: finishReason,
	}
}

// extractText extracts text from genai.Content
func extractText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var text string
	for _, part := range content.Parts {
		if part.Text != "" {
			text += part.Text
		}
	}
	return text
}
