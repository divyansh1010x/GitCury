package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmAction asks the user to confirm an action with yes/no prompt
func ConfirmAction(message string, defaultYes bool) bool {
	// Skip confirmation in non-interactive environments
	if os.Getenv("GITCURY_NONINTERACTIVE") == "1" {
		return defaultYes
	}

	prompt := message + " "
	if defaultYes {
		prompt += "[Y/n]: "
	} else {
		prompt += "[y/N]: "
	}

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		Warning("[CONFIRM]: Error reading input, using default: " + err.Error())
		return defaultYes
	}

	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

// ConfirmActionWithDetails asks for confirmation and shows detailed information
func ConfirmActionWithDetails(action string, details []string, defaultYes bool) bool {
	fmt.Println("📝 " + action)

	if len(details) > 0 {
		fmt.Println("Details:")
		for _, detail := range details {
			fmt.Println("  • " + detail)
		}
		fmt.Println()
	}

	return ConfirmAction("Do you want to continue?", defaultYes)
}

// PromptForInput asks the user for text input with an optional default value
func PromptForInput(message string, defaultValue string) string {
	// Show default value in prompt if provided
	prompt := message
	if defaultValue != "" {
		prompt += fmt.Sprintf(" [default: %s]", defaultValue)
	}
	prompt += ": "

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		Warning("[PROMPT]: Error reading input, using default: " + err.Error())
		return defaultValue
	}

	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	return input
}

// PromptForSelection asks the user to select from a list of options
func PromptForSelection(message string, options []string, defaultIndex int) (string, int) {
	if len(options) == 0 {
		return "", -1
	}

	// Default to first item if default index is invalid
	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}

	fmt.Println(message)
	for i, option := range options {
		marker := " "
		if i == defaultIndex {
			marker = "*"
		}
		fmt.Printf("  %s %d) %s\n", marker, i+1, option)
	}

	selectedIndex := -1

	for selectedIndex < 0 || selectedIndex >= len(options) {
		input := PromptForInput("Enter number", fmt.Sprintf("%d", defaultIndex+1))

		// Try to parse as number
		var parsedIndex int
		_, err := fmt.Sscanf(input, "%d", &parsedIndex)
		if err != nil || parsedIndex < 1 || parsedIndex > len(options) {
			fmt.Printf("Please enter a number between 1 and %d\n", len(options))
			continue
		}

		selectedIndex = parsedIndex - 1
	}

	return options[selectedIndex], selectedIndex
}

// ShowProgressiveConfirmation shows a multi-step confirmation dialog
// where each step depends on the previous one
func ShowProgressiveConfirmation(steps []string, actions []func() error) error {
	if len(steps) != len(actions) {
		return fmt.Errorf("mismatch between steps (%d) and actions (%d)", len(steps), len(actions))
	}

	for i, step := range steps {
		if !ConfirmAction(fmt.Sprintf("Step %d/%d: %s", i+1, len(steps), step), true) {
			return fmt.Errorf("action cancelled by user at step %d", i+1)
		}

		Info(fmt.Sprintf("Executing step %d/%d: %s", i+1, len(steps), step))
		if err := actions[i](); err != nil {
			Error(fmt.Sprintf("Step %d failed: %s", i+1, err.Error()))
			return err
		}

		Success(fmt.Sprintf("Step %d/%d completed successfully", i+1, len(steps)))
	}

	return nil
}
