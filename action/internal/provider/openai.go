package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type openaiProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAI(apiKey, model string) Provider {
	if model == "" {
		model = "gpt-4o"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &openaiProvider{client: &client, model: model}
}

func (o *openaiProvider) Chat(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	// Convert tools
	tools := make([]openai.ChatCompletionToolParam, len(params.Tools))
	for i, t := range params.Tools {
		schema := shared.FunctionParameters{
			"type":       "object",
			"properties": t.Parameters,
		}
		if len(t.Required) > 0 {
			schema["required"] = t.Required
		}
		tools[i] = openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        t.Name,
				Description: param.NewOpt(t.Description),
				Parameters:  schema,
			},
		}
	}

	// Convert messages
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(params.Messages)+1)

	if params.System != "" {
		messages = append(messages, openai.SystemMessage(params.System))
	}

	for _, msg := range params.Messages {
		switch msg.Role {
		case RoleUser:
			for _, block := range msg.Content {
				switch block.Type {
				case "text":
					messages = append(messages, openai.UserMessage(block.Text))
				case "tool_result":
					messages = append(messages, openai.ToolMessage(block.ToolResult, block.ToolResultID))
				}
			}
		case RoleAssistant:
			var toolCalls []openai.ChatCompletionMessageToolCallParam
			var textContent string

			for _, block := range msg.Content {
				switch block.Type {
				case "text":
					textContent = block.Text
				case "tool_use":
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID: block.ToolUseID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      block.ToolName,
							Arguments: string(block.ToolInput),
						},
					})
				}
			}

			assistantMsg := openai.ChatCompletionAssistantMessageParam{}
			if textContent != "" {
				assistantMsg.Content.OfString = param.NewOpt(textContent)
			}
			if len(toolCalls) > 0 {
				assistantMsg.ToolCalls = toolCalls
			}
			messages = append(messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantMsg})
		}
	}

	resp, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    o.model,
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	choice := resp.Choices[0]
	var content []ContentBlock

	if choice.Message.Content != "" {
		content = append(content, NewTextBlock(choice.Message.Content))
	}

	for _, tc := range choice.Message.ToolCalls {
		inputRaw := json.RawMessage(tc.Function.Arguments)
		content = append(content, NewToolUseBlock(tc.ID, tc.Function.Name, inputRaw))
	}

	stopReason := StopReasonEndTurn
	if choice.FinishReason == "tool_calls" || choice.FinishReason == "function_call" {
		stopReason = StopReasonToolUse
	} else if choice.FinishReason == "length" {
		stopReason = StopReasonMaxTokens
	}

	return &ChatResponse{Content: content, StopReason: stopReason}, nil
}
