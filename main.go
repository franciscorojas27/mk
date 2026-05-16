package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mk <path/to/file>")
		os.Exit(1)
	}

	rawInput := os.Args[1]
	targets := expandBraces(rawInput)

	for _, target := range targets {
		err := touchFile(target)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", target, err)
		}
	}
}

func expandBraces(input string) []string {
	start := strings.Index(input, "{")
	end := strings.Index(input, "}")

	if start == -1 || end == -1 || start >= end {
		input = strings.ReplaceAll(input, "{", "")
		input = strings.ReplaceAll(input, "}", "")
		return []string{input}
	}

	prefix := input[:start]
	suffix := input[end+1:]
	inside := input[start+1 : end]

	options := strings.Split(inside, ",")
	var results []string

	for _, opt := range options {
		results = append(results, prefix+strings.TrimSpace(opt)+suffix)
	}

	return results
}

func touchFile(path string) error {
	dir := filepath.Dir(path)

	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creating directories: %w", err)
		}
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		file.Close()
		fmt.Printf("Successfully created: %s\n", path)
	} else if err == nil {
		now := time.Now()
		err = os.Chtimes(path, now, now)
		if err != nil {
			return fmt.Errorf("error updating timestamps: %w", err)
		}
		fmt.Printf("Successfully touched: %s\n", path)
	} else {
		return fmt.Errorf("error checking file: %w", err)
	}

	return nil
}