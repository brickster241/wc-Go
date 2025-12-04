package services

import (
	"bufio"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

// WCResult struct to hold the word count results for a file
type WCResult struct {
	FileName string // Name of the file
	Lines    int    // Number of lines
	Words    int    // Number of words
	Bytes    int    // Number of bytes
	Chars    int    // Number of characters
}

// ProcessFilesConcurrent processes multiple files concurrently and returns their word count results.
func ProcessFilesConcurrent(cfg wcCLI) ([]WCResult, error) {

	// IF no files, process STDIN directly (synchronously)
	if len(cfg.files) == 0 {
		r := processReader("stdin", os.Stdin, cfg)
		return []WCResult{r}, nil
	}

	results := make([]WCResult, len(cfg.files))
	errCh := make(chan error, len(cfg.files))
	doneCh := make(chan struct{})

	// Process each file in a separate goroutine
	for i, fname := range cfg.files {
		go func(i int, fname string) {
			f, err := os.Open(fname)
			if err != nil {
				errCh <- err
				return
			}
			defer f.Close()

			results[i] = processReader(fname, f, cfg)
			errCh <- nil
		}(i, fname)
	}

	// Wait for all goroutines to finish or the first error
	go func() {
		for range cfg.files {
			if err := <-errCh; err != nil {
				// Send final error and stop
				doneCh <- struct{}{}
				return
			}
		}
		doneCh <- struct{}{}
	}()

	// Wait for completion
	<-doneCh
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// processReader processes a single file reader and returns the WCResult.
func processReader(fileName string, reader io.Reader, cfg wcCLI) WCResult {

	result := WCResult{FileName: fileName}

	// Read raw bytes for byte count (fast)
	byteReader := bufio.NewReader(reader)

	// Read chunks, not LINES (because this will support huge lines too)
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := byteReader.Read(buf)
		if n > 0 {
			chunk := buf[:n]

			// Byte count
			if cfg.bytes {
				result.Bytes += n
			}
			if cfg.chars || cfg.words || cfg.lines {
				// Convert bytes to string (UTF-8 safe)
				s := string(chunk)

				// Line count
				if cfg.lines {
					result.Lines += countLines(s)
				}

				// Word count
				if cfg.words {
					result.Words += countWords(s)
				}

				// Character count
				if cfg.chars {
					result.Chars += countRunes(s) // UTF-8 aware character count
				}
			}
		}
		if err == io.EOF {
			break
		}

		if err != nil {
			// Handle read error (could log or return)
			break
		}
	}
	return result
}

// countLines counts the number of lines in a string.
func countLines(s string) int {
	count := 0
	for _, r := range s {
		if r == '\n' {
			count++
		}
	}
	return count
}

// countWords counts the number of words in a string.
func countWords(s string) int {
	inWord := false
	count := 0

	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else {
			if !inWord {
				count++
			}
			inWord = true
		}
	}
	return count
}

// countRunes counts the number of runes (characters) in a string.
func countRunes(s string) int {
	return utf8.RuneCountInString(s)
}
