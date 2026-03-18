package ai

import (
	"fmt"
	"sort"
	"strings"
)

// ExplainPrompt builds the prompt for the explain command.
func ExplainPrompt(diff, query string) string {
	if query != "" {
		return fmt.Sprintf("Given the following git diff, answer this question: %s\n\n```diff\n%s\n```", query, diff)
	}
	return fmt.Sprintf("Summarize the following git diff in clear, concise prose. Focus on what changed and why it matters.\n\n```diff\n%s\n```", diff)
}

// DraftPrompt builds the prompt for the draft command.
func DraftPrompt(diff, context string, commitTypes map[string]string) string {
	// Sort for deterministic output
	keys := make([]string, 0, len(commitTypes))
	for k := range commitTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	types := make([]string, 0, len(commitTypes))
	for _, k := range keys {
		types = append(types, fmt.Sprintf("  %s: %s", k, commitTypes[k]))
	}
	contextLine := ""
	if context != "" {
		contextLine = fmt.Sprintf("\nAdditional context from the author: %s\n", context)
	}
	return fmt.Sprintf(`Generate a conventional commit message for the following diff.%s
Commit types:
%s

Rules:
- Format: <type>(<optional scope>): <description>
- Description must be imperative mood, lowercase, no period
- Keep it under 72 characters
- Output ONLY the commit message, nothing else

`+"```diff\n%s\n```", contextLine, strings.Join(types, "\n"), diff)
}

// OperatePrompt builds the prompt for the operate command.
func OperatePrompt(query string) string {
	return fmt.Sprintf(`Generate a git command for the following request: %s

Respond with exactly 3 lines:
1. The git command (e.g. "git rebase -i HEAD~3")
2. A one-sentence explanation
3. Either "WARNING: <reason>" if the command is destructive, or an empty line

Output only these lines, no markdown, no extra text.`, query)
}
