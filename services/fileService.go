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
func ProcessFilesConcurrent(cfg WcCLI) ([]WCResult, error) {

	// IF no files, process STDIN directly (synchronously)
	if len(cfg.Files) == 0 {
		r := processReader("stdin", os.Stdin, cfg)
		return []WCResult{r}, nil
	}

	results := make([]WCResult, len(cfg.Files))
	errCh := make(chan error, len(cfg.Files))
	doneCh := make(chan struct{})

	// Process each file in a separate goroutine
	for i, fname := range cfg.Files {
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
		for range cfg.Files {
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
func processReader(fileName string, reader io.Reader, cfg WcCLI) WCResult {

	result := WCResult{FileName: fileName}

	// Read raw bytes for byte count (fast)
	byteReader := bufio.NewReader(reader)

	// Read chunks, not LINES (because this will support huge lines too)
	buf := make([]byte, 32*1024) // 32KB buffer

	// Track if we are in a word for word counting
	inWord := false

	// Carry over bytes for incomplete UTF-8 characters
	var carry []byte

	for {
		n, err := byteReader.Read(buf)
		if n > 0 {
			// Convert bytes to string (UTF-8 safe)
			chunk := buf[:n]
			s := string(chunk)

			// Byte count
			if cfg.Bytes {
				result.Bytes += n
			}

			// Line count
			if cfg.Lines {
				result.Lines += countLines(s)
			}

			// Word count
			if cfg.Words {
				words := countWords(s, &inWord)
				result.Words += words
			}

			// Character count
			if cfg.Chars {
				runeCount, leftover := decodeAndCountRunes(chunk, carry)
				result.Chars += runeCount
				carry = leftover
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
func countWords(s string, inWord *bool) int {
	count := 0

	// A word is defined as a sequence of non-space characters
	for _, r := range s {
		if unicode.IsSpace(r) {
			*inWord = false
		} else {
			if !*inWord {
				count++
			}
			*inWord = true
		}
	}
	return count
}

// decodeAndCountRunes decodes UTF-8 runes from a byte slice and counts them, handling incomplete runes.
func decodeAndCountRunes(chunk []byte, carry []byte) (runes int, newCarry []byte) {
	// Prepend carry to the chunk
	buf := append(carry, chunk...)

	i := 0
	for i < len(buf) {
		r, size := utf8.DecodeRune(buf[i:])
		if r == utf8.RuneError && size == 1 {
			// Incomplete rune at the end -> save it for the next chunk
			break
		}

		runes++
		i += size

	}
	newCarry = append([]byte{}, buf[i:]...) // Save incomplete bytes
	return runes, newCarry
}
