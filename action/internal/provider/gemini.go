package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

type geminiProvider struct {
	client *genai.Client
	model  string
}

func NewGemini(apiKey, model string) Provider {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return &geminiProvider{client: nil, model: model}
	}
	return &geminiProvider{client: client, model: model}
}

func (g *geminiProvider) Chat(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	if g.client == nil {
		return nil, fmt.Errorf("gemini client failed to initialize")
	}

	// Convert tools to Gemini format
	var funcDecls []*genai.FunctionDeclaration
	for _, t := range params.Tools {
		props := make(map[string]*genai.Schema)
		for name, val := range t.Parameters {
			props[name] = convertToGeminiSchema(val)
		}

		funcDecls = append(funcDecls, &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
			Parameters: &genai.Schema{
				Type:       genai.TypeObject,
				Properties: props,
				Required:   t.Required,
			},
		})
	}

	geminiTools := []*genai.Tool{{FunctionDeclarations: funcDecls}}

	// Convert messages to Gemini format
	var contents []*genai.Content

	for _, msg := range params.Messages {
		role := genai.RoleUser
		if msg.Role == RoleAssistant {
			role = genai.RoleModel
		}

		var parts []*genai.Part
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				parts = append(parts, &genai.Part{Text: block.Text})
			case "tool_use":
				args := make(map[string]any)
				json.Unmarshal(block.ToolInput, &args)
				parts = append(parts, genai.NewPartFromFunctionCall(block.ToolName, args))
			case "tool_result":
				response := make(map[string]any)
				response["result"] = block.ToolResult
				if block.IsError {
					response["error"] = true
				}
				parts = append(parts, genai.NewPartFromFunctionResponse(block.ToolResultID, response))
			}
		}

		if len(parts) > 0 {
			contents = append(contents, &genai.Content{
				Parts: parts,
				Role:  role,
			})
		}
	}

	config := &genai.GenerateContentConfig{
		Tools: geminiTools,
	}
	if params.System != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: params.System}},
		}
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	// Convert response
	var content []ContentBlock
	hasToolCalls := false

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				content = append(content, NewTextBlock(part.Text))
			}
			if part.FunctionCall != nil {
				hasToolCalls = true
				argsRaw, _ := json.Marshal(part.FunctionCall.Args)
				id := fmt.Sprintf("call_%s", part.FunctionCall.Name)
				content = append(content, NewToolUseBlock(id, part.FunctionCall.Name, argsRaw))
			}
		}
	}

	stopReason := StopReasonEndTurn
	if hasToolCalls {
		stopReason = StopReasonToolUse
	}

	return &ChatResponse{Content: content, StopReason: stopReason}, nil
}

func convertToGeminiSchema(val interface{}) *genai.Schema {
	m, ok := val.(map[string]interface{})
	if !ok {
		return &genai.Schema{Type: genai.TypeString}
	}

	schema := &genai.Schema{}
	if t, ok := m["type"].(string); ok {
		schema.Type = genai.Type(t)
	}
	if d, ok := m["description"].(string); ok {
		schema.Description = d
	}
	return schema
}
