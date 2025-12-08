package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/brickster241/wc-Go/services"
)

// printResults prints the word count results based on the provided configuration
func printResults(cfg services.WcCLI, results []services.WCResult) {

	multipleLines := len(results) > 1
	var totalLines, totalWords, totalBytes, totalChars int
	wLines, wWords, wBytes, wChars := 8, 8, 8, 8
	for _, res := range results {

		totalLines += res.Lines
		totalWords += res.Words
		totalBytes += res.Bytes
		totalChars += res.Chars

		// Convert numbers so we can measure length
		lStr := strconv.FormatInt(int64(totalLines), 10)
		wStr := strconv.FormatInt(int64(totalWords), 10)
		cStr := strconv.FormatInt(int64(totalBytes), 10)
		mStr := strconv.FormatInt(int64(totalChars), 10)

		// Compute column widths
		if len(lStr) >= wLines {
			wLines = len(lStr) + 1
		}
		if len(wStr) >= wWords {
			wWords = len(wStr) + 1
		}
		if len(cStr) >= wBytes {
			wBytes = len(cStr) + 1
		}
		if len(mStr) >= wChars {
			wChars = len(mStr) + 1
		}
	}

	for _, res := range results {
		if cfg.Lines {
			fmt.Printf("%*d", wLines, res.Lines)
		}
		if cfg.Words {
			fmt.Printf("%*d", wWords, res.Words)
		}
		if cfg.Bytes {
			fmt.Printf("%*d", wBytes, res.Bytes)
		}
		if cfg.Chars {
			fmt.Printf("%*d", wChars, res.Chars)
		}
		// Print filename unless reading from stdin
		if res.FileName != "stdin" {
			fmt.Printf(" %s\n", res.FileName)
		} else {
			fmt.Printf("\n")
		}
	}

	// Print totals if multiple files were processed
	if multipleLines {
		if cfg.Lines {
			fmt.Printf("%*d", wLines, totalLines)
		}
		if cfg.Words {
			fmt.Printf("%*d", wWords, totalWords)
		}
		if cfg.Bytes {
			fmt.Printf("%*d", wBytes, totalBytes)
		}
		if cfg.Chars {
			fmt.Printf("%*d", wChars, totalChars)
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
