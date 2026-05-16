package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Summary struct {
	Created int
	Touched int
	Dirs    int
	Failed  int
}

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "n", false, "show actions without creating files or directories")
	flag.BoolVar(&dryRun, "dry-run", false, "show actions without creating files or directories")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] <path1> <path2> ...\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	summary := &Summary{}

	for _, arg := range flag.Args() {
		targets := expandBraces(arg)
		for _, target := range targets {
			err := processTarget(target, dryRun, summary)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", target, err)
			}
		}
	}

	fmt.Println("\n--- Execution Summary ---")
	if dryRun {
		fmt.Println("(Dry-Run Mode: No physical changes were made)")
	}
	fmt.Printf("Directories processed: %d\n", summary.Dirs)
	fmt.Printf("Files created:         %d\n", summary.Created)
	fmt.Printf("Files touched:         %d\n", summary.Touched)
	if summary.Failed > 0 {
		fmt.Printf("Operations failed:     %d\n", summary.Failed)
	}
}

func expandBraces(input string) []string {
	start := strings.Index(input, "{")
	if start == -1 {
		return []string{input}
	}

	end := strings.Index(input, "}")
	if end == -1 || start >= end {
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
		expanded := prefix + strings.TrimSpace(opt) + suffix
		results = append(results, expandBraces(expanded)...)
	}
	return results
}

func processTarget(path string, dryRun bool, sum *Summary) error {
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		return createOnlyDir(path, dryRun, sum)
	}
	return touchFile(path, dryRun, sum)
}

func createOnlyDir(path string, dryRun bool, sum *Summary) error {
	if dryRun {
		fmt.Printf("Would create directory: %s\n", path)
		sum.Dirs++
		return nil
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		sum.Failed++
		return fmt.Errorf("error creating directory: %w", err)
	}
	fmt.Printf("Successfully created directory: %s\n", path)
	sum.Dirs++
	return nil
}

func touchFile(path string, dryRun bool, sum *Summary) error {
	dir := filepath.Dir(path)
	_, err := os.Stat(path)
	exists := err == nil
	if err != nil && !os.IsNotExist(err) {
		sum.Failed++
		return fmt.Errorf("error checking file: %w", err)
	}

	if dryRun {
		if exists {
			fmt.Printf("Would touch: %s\n", path)
			sum.Touched++
		} else {
			if dir != "." && dir != "" {
				fmt.Printf("Would create directories and file: %s\n", path)
			} else {
				fmt.Printf("Would create file: %s\n", path)
			}
			sum.Created++
		}
		return nil
	}

	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			sum.Failed++
			return fmt.Errorf("error creating directories: %w", err)
		}
	}

	if os.IsNotExist(err) {
		ext := filepath.Ext(path)
		perm := os.FileMode(0644)
		if ext == ".sh" {
			perm = os.FileMode(0755)
		}

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, perm)
		if err != nil {
			sum.Failed++
			return fmt.Errorf("error creating file: %w", err)
		}

		injectBoilerplate(file, path, dir)
		file.Close()
		fmt.Printf("Successfully created: %s\n", path)
		sum.Created++
	} else if err == nil {
		now := time.Now()
		err = os.Chtimes(path, now, now)
		if err != nil {
			sum.Failed++
			return fmt.Errorf("error updating timestamps: %w", err)
		}
		fmt.Printf("Successfully touched: %s\n", path)
		sum.Touched++
	} else {
		sum.Failed++
		return fmt.Errorf("error checking file: %w", err)
	}

	return nil
}

func injectBoilerplate(file *os.File, path string, dir string) {
	ext := filepath.Ext(path)
	base := filepath.Base(path)

	switch ext {
	case ".go":
		if base == "main.go" {
			file.WriteString("package main\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n")
		} else {
			pkgName := filepath.Base(dir)
			if pkgName == "." || pkgName == "" {
				pkgName = "main"
			} else {
				pkgName = strings.ReplaceAll(pkgName, "-", "_")
			}
			file.WriteString(fmt.Sprintf("package %s\n", pkgName))
		}
	case ".html":
		file.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n\t<meta charset=\"UTF-8\">\n\t<title>Document</title>\n</head>\n<body>\n\t\n</body>\n</html>\n")
	case ".sh":
		file.WriteString("#!/bin/bash\n\nset -e\n\n")
	case ".json":
		file.WriteString("{\n\t\n}\n")
	case ".md":
		title := strings.TrimSuffix(base, ext)
		file.WriteString(fmt.Sprintf("# %s\n\n", strings.Title(title)))
	case ".yml", ".yaml":
		file.WriteString("---\n")
	default:
		if base == "go.mod" {
			wd, err := os.Getwd()
			modName := "project"
			if err == nil {
				modName = filepath.Base(wd)
			}
			file.WriteString(fmt.Sprintf("module %s\n\ngo 1.24\n", modName))
		}
	}
}