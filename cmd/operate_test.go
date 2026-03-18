package cmd

import "testing"

func TestParseOperateResponse(t *testing.T) {
	cases := []struct {
		input                           string
		command, explanation, warning string
	}{
		{
			"git rebase -i HEAD~3\nReorder the last 3 commits interactively\nWARNING: rewrites commit history",
			"git rebase -i HEAD~3", "Reorder the last 3 commits interactively", "rewrites commit history",
		},
		{
			"git log --oneline\nShow commit history\n",
			"git log --oneline", "Show commit history", "",
		},
	}
	for _, tc := range cases {
		cmd, exp, warn := parseOperateResponse(tc.input)
		if cmd != tc.command || exp != tc.explanation || warn != tc.warning {
			t.Errorf("parseOperateResponse(%q) = (%q, %q, %q), want (%q, %q, %q)",
				tc.input, cmd, exp, warn, tc.command, tc.explanation, tc.warning)
		}
	}
}
