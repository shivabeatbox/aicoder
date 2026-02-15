package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AkshayNayak/ticketflow/action/internal/agent"
)

func main() {
	cfg := agent.Config{
		Provider:          getInput("PROVIDER", "claude"),
		APIKey:            requireInput("API_KEY"),
		Model:             getInput("MODEL", ""),
		TicketKey:         requireInput("TICKET_KEY"),
		TicketTitle:       requireInput("TICKET_TITLE"),
		TicketDescription: requireInput("TICKET_DESCRIPTION"),
		Workspace:         getEnv("GITHUB_WORKSPACE", "."),
	}

	log.Printf("Sprint Code Agent starting...")
	log.Printf("Ticket: %s - %s", cfg.TicketKey, cfg.TicketTitle)
	log.Printf("Provider: %s | Model: %s", cfg.Provider, cfg.Model)
	log.Printf("Workspace: %s", cfg.Workspace)

	a, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	result, err := a.Run(context.Background())
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	log.Printf("Agent completed successfully!")
	log.Printf("Files changed: %s", strings.Join(result.FilesChanged, ", "))
	log.Printf("Summary: %s", result.Summary)

	// Write outputs for GitHub Actions
	writeOutput("files_changed", strings.Join(result.FilesChanged, ","))
	writeOutput("summary", result.Summary)
}

// requireInput reads a GitHub Actions input (INPUT_ env var) and exits if not set.
func requireInput(name string) string {
	val := os.Getenv("INPUT_" + strings.ToUpper(name))
	if val == "" {
		log.Fatalf("Required input %s is not set", name)
	}
	return val
}

// getInput reads a GitHub Actions input with a default fallback.
func getInput(name, defaultVal string) string {
	val := os.Getenv("INPUT_" + strings.ToUpper(name))
	if val == "" {
		return defaultVal
	}
	return val
}

// getEnv reads an environment variable with a default fallback.
func getEnv(name, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
}

// writeOutput writes a value to the GitHub Actions output file.
func writeOutput(name, value string) {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		return
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		log.Printf("Warning: could not write output %s: %v", name, err)
		return
	}
	defer f.Close()

	// Use delimiter for multi-line values
	delimiter := "EOF_TICKETFLOW"
	fmt.Fprintf(f, "%s<<%s\n%s\n%s\n", name, delimiter, value, delimiter)
}
