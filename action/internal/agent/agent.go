package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AkshayNayak/ticketflow/action/internal/provider"
)

// Config holds the configuration for the agent.
type Config struct {
	Provider          string // "claude", "openai", "gemini"
	APIKey            string
	Model             string
	TicketKey         string
	TicketTitle       string
	TicketDescription string
	Workspace         string
	MaxTurns          int
}

// Result holds the outcome of an agent run.
type Result struct {
	Summary      string
	FilesChanged []string
}

// Agent orchestrates the AI-powered implementation loop.
type Agent struct {
	config   Config
	provider provider.Provider
	tracker  *ChangeTracker
}

// New creates a new Agent with the given configuration.
func New(cfg Config) (*Agent, error) {
	if cfg.MaxTurns == 0 {
		cfg.MaxTurns = 50
	}
	if cfg.Provider == "" {
		cfg.Provider = "claude"
	}

	p, err := provider.NewProvider(cfg.Provider, cfg.APIKey, cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &Agent{
		config:   cfg,
		provider: p,
		tracker:  NewChangeTracker(),
	}, nil
}

// Run executes the agent loop: sends messages to the LLM, handles tool calls,
// and repeats until the model stops requesting tools or max turns is reached.
func (a *Agent) Run(ctx context.Context) (*Result, error) {
	systemPrompt := BuildSystemPrompt(a.config.TicketKey, a.config.TicketTitle, a.config.TicketDescription)

	repoTree := buildRepoTree(a.config.Workspace)
	initialMessage := BuildInitialUserMessage(repoTree)

	messages := []provider.Message{
		provider.UserMessage(provider.NewTextBlock(initialMessage)),
	}
	tools := ToolDefinitions()

	var summaryParts []string

	for turn := 0; turn < a.config.MaxTurns; turn++ {
		log.Printf("[turn %d] Sending request to %s (%s)...", turn+1, a.config.Provider, a.config.Model)

		response, err := a.provider.Chat(ctx, provider.ChatParams{
			System:    systemPrompt,
			Messages:  messages,
			Tools:     tools,
			MaxTokens: 8192,
		})
		if err != nil {
			return nil, fmt.Errorf("API error on turn %d: %w", turn+1, err)
		}

		log.Printf("[turn %d] Stop reason: %s, content blocks: %d", turn+1, response.StopReason, len(response.Content))

		// Process response content blocks
		var assistantBlocks []provider.ContentBlock
		var toolResultBlocks []provider.ContentBlock

		for _, block := range response.Content {
			assistantBlocks = append(assistantBlocks, block)

			switch block.Type {
			case "text":
				log.Printf("[turn %d] Text: %s", turn+1, truncate(block.Text, 200))
				summaryParts = append(summaryParts, block.Text)

			case "tool_use":
				log.Printf("[turn %d] Tool call: %s", turn+1, block.ToolName)

				result, isError := HandleToolCall(a.config.Workspace, block.ToolName, block.ToolInput, a.tracker)
				log.Printf("[turn %d] Tool result (%s): %s", turn+1, block.ToolName, truncate(result, 200))

				toolResultBlocks = append(toolResultBlocks, provider.NewToolResultBlock(block.ToolUseID, result, isError))
			}
		}

		// Add assistant response to conversation
		messages = append(messages, provider.AssistantMessage(assistantBlocks...))

		// If there were tool calls, send results back as a user message
		if len(toolResultBlocks) > 0 {
			messages = append(messages, provider.UserMessage(toolResultBlocks...))
		}

		// Stop if the model is done (no more tool calls)
		if response.StopReason == provider.StopReasonEndTurn {
			log.Printf("Agent completed after %d turns", turn+1)
			break
		}
	}

	summary := strings.Join(summaryParts, "\n")
	if summary == "" {
		summary = "Agent completed implementation."
	}

	return &Result{
		Summary:      summary,
		FilesChanged: a.tracker.Files(),
	}, nil
}

// buildRepoTree generates a directory tree of the workspace (up to 3 levels deep).
func buildRepoTree(workspace string) string {
	var b strings.Builder
	buildTreeRecursive(workspace, "", &b, 0, 3)
	if b.Len() == 0 {
		return "(empty repository)"
	}
	return b.String()
}

func buildTreeRecursive(dir, prefix string, b *strings.Builder, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var visible []os.DirEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
			continue
		}
		visible = append(visible, e)
	}

	for i, entry := range visible {
		isLast := i == len(visible)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := entry.Name()
		if entry.IsDir() {
			fmt.Fprintf(b, "%s%s%s/\n", prefix, connector, name)
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}
			buildTreeRecursive(dir+"/"+name, newPrefix, b, depth+1, maxDepth)
		} else {
			fmt.Fprintf(b, "%s%s%s\n", prefix, connector, name)
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
