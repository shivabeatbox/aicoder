package provider

import (
	"context"
	"encoding/json"
	"fmt"
)

// Provider is the common interface for all LLM providers.
type Provider interface {
	// Chat sends a conversation to the LLM and returns its response.
	// Supports system prompts, multi-turn messages, and tool use.
	Chat(ctx context.Context, params ChatParams) (*ChatResponse, error)
}

// ChatParams holds the parameters for a chat request.
type ChatParams struct {
	System   string
	Messages []Message
	Tools    []Tool
	MaxTokens int
}

// Role represents the role of a message sender.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a single message in the conversation.
type Message struct {
	Role    Role
	Content []ContentBlock
}

// ContentBlock is a union type for message content.
type ContentBlock struct {
	Type string // "text", "tool_use", "tool_result"

	// For text blocks
	Text string

	// For tool_use blocks
	ToolUseID   string
	ToolName    string
	ToolInput   json.RawMessage

	// For tool_result blocks
	ToolResultID string
	ToolResult   string
	IsError      bool
}

// Tool defines a tool the LLM can call.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{} // JSON Schema properties
	Required    []string
}

// StopReason indicates why the LLM stopped generating.
type StopReason string

const (
	StopReasonEndTurn StopReason = "end_turn"
	StopReasonToolUse StopReason = "tool_use"
	StopReasonMaxTokens StopReason = "max_tokens"
)

// ChatResponse holds the LLM's response.
type ChatResponse struct {
	Content    []ContentBlock
	StopReason StopReason
}

// Helper constructors

func NewTextBlock(text string) ContentBlock {
	return ContentBlock{Type: "text", Text: text}
}

func NewToolUseBlock(id, name string, input json.RawMessage) ContentBlock {
	return ContentBlock{Type: "tool_use", ToolUseID: id, ToolName: name, ToolInput: input}
}

func NewToolResultBlock(toolUseID, result string, isError bool) ContentBlock {
	return ContentBlock{Type: "tool_result", ToolResultID: toolUseID, ToolResult: result, IsError: isError}
}

func UserMessage(content ...ContentBlock) Message {
	return Message{Role: RoleUser, Content: content}
}

func AssistantMessage(content ...ContentBlock) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// NewProvider creates a provider instance based on the provider name.
func NewProvider(name, apiKey, model string) (Provider, error) {
	switch name {
	case "claude", "anthropic":
		return NewClaude(apiKey, model), nil
	case "openai", "gpt":
		return NewOpenAI(apiKey, model), nil
	case "gemini", "google":
		return NewGemini(apiKey, model), nil
	default:
		return nil, fmt.Errorf("unknown provider %q — supported: claude, openai, gemini", name)
	}
}
