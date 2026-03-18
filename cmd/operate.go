package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gitlens/internal/ai"
)

var operateCmd = &cobra.Command{
	Use:   "operate <query>",
	Short: "Generate and execute a git command from natural language",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		provider, err := ai.New(Cfg)
		if err != nil {
			return err
		}

		prompt := ai.OperatePrompt(query)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		result, err := provider.Complete(ctx, prompt)
		if err != nil {
			return err
		}

		command, explanation, warning := parseOperateResponse(result)
		if command == "" {
			return fmt.Errorf("could not parse command from AI response:\n%s", result)
		}

		fmt.Printf("Command: %s\n", command)
		fmt.Printf("Explanation: %s\n", explanation)
		if warning != "" {
			fmt.Printf("\n⚠  WARNING: %s\n", warning)
		}
		fmt.Printf("\nRun this command? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" {
			fmt.Println("Aborted.")
			return nil
		}

		parts := strings.Fields(command)
		if len(parts) == 0 {
			return fmt.Errorf("empty command")
		}
		execCmd := exec.Command(parts[0], parts[1:]...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin
		return execCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(operateCmd)
}

// parseOperateResponse parses the 3-line AI response for operate.
// Line 1: command, Line 2: explanation, Line 3 (optional): WARNING: ...
func parseOperateResponse(response string) (command, explanation, warning string) {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	if len(lines) >= 1 {
		command = strings.TrimSpace(lines[0])
	}
	if len(lines) >= 2 {
		explanation = strings.TrimSpace(lines[1])
	}
	if len(lines) >= 3 {
		line := strings.TrimSpace(lines[2])
		if strings.HasPrefix(line, "WARNING:") {
			warning = strings.TrimPrefix(line, "WARNING:")
			warning = strings.TrimSpace(warning)
		}
	}
	return
}
