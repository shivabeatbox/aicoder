package agent

import "fmt"

// BuildSystemPrompt constructs the system prompt for the Claude agent.
func BuildSystemPrompt(ticketKey, ticketTitle, ticketDescription string) string {
	return fmt.Sprintf(`You are an expert software engineer. Your task is to implement a Jira ticket by modifying the codebase in the current repository.

## Ticket
Key: %s
Title: %s
Description:
%s

## Instructions
1. Start by exploring the repository structure using list_directory to understand the codebase layout.
2. Read relevant files to understand existing code patterns, conventions, and architecture.
3. Implement the changes described in the ticket.
4. Follow the existing code style and patterns you observe in the repository.
5. Write clean, production-ready code.
6. Only modify or create files directly related to the ticket requirements.
7. If the ticket requires new dependencies, mention them but do not run install commands.

## Guidelines
- Prefer editing existing files over creating new ones when possible.
- Use edit_file for targeted changes to existing files.
- Use write_file only for new files or complete rewrites.
- Always read a file before editing it.
- Keep changes minimal and focused on the ticket requirements.
- Do not add unnecessary comments, documentation, or boilerplate.`, ticketKey, ticketTitle, ticketDescription)
}

// BuildInitialUserMessage constructs the first user message with repo context.
func BuildInitialUserMessage(repoTree string) string {
	return fmt.Sprintf(`Here is the current repository structure:

%s

Please implement the Jira ticket described in the system prompt. Start by exploring the codebase to understand its structure, then make the necessary changes.`, repoTree)
}
