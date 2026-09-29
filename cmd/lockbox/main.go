package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var stdinReader = bufio.NewReader(os.Stdin)

func readMasterPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		raw, err := term.ReadPassword(fd)
		fmt.Println()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(raw)), nil
	}
	line, err := stdinReader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readNewMasterPassword() (string, error) {
	pw, err := readMasterPassword("🔑 Create Master Password: ")
	if err != nil {
		return "", err
	}
	if pw == "" {
		return "", fmt.Errorf("🚨 Master password cannot be blank.")
	}
	confirm, err := readMasterPassword("🔑 Confirm Master Password: ")
	if err != nil {
		return "", err
	}
	if pw != confirm {
		return "", fmt.Errorf("🚨 Passwords do not match.")
	}
	return pw, nil
}
