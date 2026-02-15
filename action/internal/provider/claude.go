package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type claudeProvider struct {
	client *anthropic.Client
	model  string
}

func NewClaude(apiKey, model string) Provider {
	if model == "" {
		model = "claude-sonnet-4-5-20250929"
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &claudeProvider{client: &client, model: model}
}

func (c *claudeProvider) Chat(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	maxTokens := params.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8192
	}

	// Convert tools
	tools := make([]anthropic.ToolUnionParam, len(params.Tools))
	for i, t := range params.Tools {
		tools[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: t.Parameters,
					Required:   t.Required,
				},
			},
		}
	}

	// Convert messages
	messages := make([]anthropic.MessageParam, 0, len(params.Messages))
	for _, msg := range params.Messages {
		blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.Content))
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				blocks = append(blocks, anthropic.NewTextBlock(block.Text))
			case "tool_use":
				// Tool use blocks are part of assistant messages — handled via ToParam
			case "tool_result":
				blocks = append(blocks, anthropic.NewToolResultBlock(block.ToolResultID, block.ToolResult, block.IsError))
			}
		}

		if msg.Role == RoleUser {
			messages = append(messages, anthropic.NewUserMessage(blocks...))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(blocks...))
		}
	}

	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: int64(maxTokens),
		System: []anthropic.TextBlockParam{
			{Text: params.System},
		},
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	// Convert response
	var content []ContentBlock
	for _, block := range resp.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			content = append(content, NewTextBlock(variant.Text))
		case anthropic.ToolUseBlock:
			inputRaw, _ := json.Marshal(variant.Input)
			content = append(content, NewToolUseBlock(variant.ID, variant.Name, inputRaw))
		}
	}

	stopReason := StopReasonEndTurn
	if resp.StopReason == "tool_use" {
		stopReason = StopReasonToolUse
	} else if resp.StopReason == "max_tokens" {
		stopReason = StopReasonMaxTokens
	}

	return &ChatResponse{Content: content, StopReason: stopReason}, nil
}
