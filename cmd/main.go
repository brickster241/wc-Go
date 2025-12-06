package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/brickster241/wc-Go/services"
)

// printResults prints the word count results based on the provided configuration
func printResults(cfg services.WcCLI, results []services.WCResult) {

	multipleLines := len(results) > 1
	var totalLines, totalWords, totalBytes, totalChars int

	for _, res := range results {
		if cfg.Lines {
			fmt.Printf("%8d", res.Lines)
		}
		if cfg.Words {
			fmt.Printf("%8d", res.Words)
		}
		if cfg.Bytes {
			fmt.Printf("%8d", res.Bytes)
		}
		if cfg.Chars {
			fmt.Printf("%8d", res.Chars)
		}

		// Print filename unless reading from stdin
		if res.FileName != "stdin" {
			fmt.Printf(" %s\n", filepath.Base(res.FileName))
		} else {
			fmt.Printf("\n")
		}

		totalLines += res.Lines
		totalWords += res.Words
		totalBytes += res.Bytes
		totalChars += res.Chars
	}

	// Print totals if multiple files were processed
	if multipleLines {
		if cfg.Lines {
			fmt.Printf("%8d", totalLines)
		}
		if cfg.Words {
			fmt.Printf("%8d", totalWords)
		}
		if cfg.Bytes {
			fmt.Printf("%8d", totalBytes)
		}
		if cfg.Chars {
			fmt.Printf("%8d", totalChars)
		}

		fmt.Printf(" total\n")
	}
}

func main() {

	// 1. Parse Flags
	cfg := services.GetCLIFlags()

	// If no flags are provided, wc defaults to printing ALL: -l -w -c
	if !cfg.Words && !cfg.Lines && !cfg.Bytes && !cfg.Chars {
		cfg.Words = true
		cfg.Lines = true
		cfg.Bytes = true
	}

	// 2. Process Files (concurrently if multiple files are provided)
	results, err := services.ProcessFilesConcurrent(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		os.Exit(1)
	}

	// 3. Print Results
	printResults(cfg, results)
}
