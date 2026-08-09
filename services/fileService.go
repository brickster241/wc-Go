package services

import (
	"bufio"
	"bytes"
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
			// -c alone never needs the bytes, only how many there are —
			// stat and stop, the same shortcut coreutils wc takes. This is
			// the difference between O(1) and reading a 512MB file.
			if cfg.Bytes && !cfg.Lines && !cfg.Words && !cfg.Chars {
				if st, err := os.Stat(fname); err == nil && st.Mode().IsRegular() {
					results[i] = WCResult{FileName: fname, Bytes: int(st.Size())}
					errCh <- nil
					return
				}
			}
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

	// Bytes of a rune that straddles the chunk boundary. Every counter that
	// decodes runes (words, chars) must see the stream through this carry, or
	// a multi-byte character split across two Reads decodes as garbage twice.
	var carry []byte

	for {
		n, err := byteReader.Read(buf)
		if n > 0 {
			// Byte count is over the RAW chunk — the carry shuffles bytes
			// between iterations but never changes how many there were.
			if cfg.Bytes {
				result.Bytes += n
			}

			// All rune-aware counters share one carry-corrected view: the
			// joined buffer minus any incomplete rune at its tail. Counters
			// work on the []byte directly — a string(valid) here would copy
			// 32KB per chunk, a tax every single mode pays.
			joined := append(carry, buf[:n]...)
			valid, rest := splitIncompleteTail(joined)
			carry = rest

			if cfg.Lines {
				result.Lines += bytes.Count(valid, newline)
			}
			if cfg.Words {
				result.Words += countWords(valid, &inWord)
			}
			if cfg.Chars {
				// RuneCount advances one byte per undecodable byte — the
				// same policy wc applies in a UTF-8 locale: every valid
				// rune counts 1, every invalid byte counts 1.
				result.Chars += utf8.RuneCount(valid)
			}
		}
		if err == io.EOF {
			// A rune left unfinished by EOF is dropped, not counted —
			// verified against wc(1): `printf 'ab\xc3' | wc -m` says 2.
			break
		}

		if err != nil {
			// Handle read error (could log or return)
			break
		}
	}
	return result
}

// splitIncompleteTail splits buf into the prefix safe to decode now and a
// trailing fragment that may be the start of a rune finished by the next
// chunk. The distinction that matters: an INVALID byte mid-stream is data
// (wc counts it and moves on one byte); only an incomplete sequence at the
// very END is deferred. Conflating the two — treating every RuneError as
// "wait for more bytes" — silently drops the rest of the chunk, which is
// exactly the bug the differential suite caught here.
func splitIncompleteTail(buf []byte) (valid, rest []byte) {
	tail := len(buf)
	// A rune is at most UTFMax bytes, so only the last few can be incomplete.
	for back := 1; back <= utf8.UTFMax && back <= len(buf); back++ {
		i := len(buf) - back
		if utf8.RuneStart(buf[i]) {
			if !utf8.FullRune(buf[i:]) {
				tail = i // defer this fragment to the next chunk
			}
			break
		}
	}
	return buf[:tail], append([]byte{}, buf[tail:]...)
}

var newline = []byte{'\n'}

// countWords counts words: maximal runs of non-space characters.
//
// The hot path is byte-wise: ASCII classifies from a 128-entry table with no
// rune decoding at all, and only a high byte (>= 0x80) pays for
// utf8.DecodeRuneInString + unicode.IsSpace. On plain-text corpora this is
// the difference between trailing wc and beating it — measured by wcbench,
// not assumed.
func countWords(s []byte, inWord *bool) int {
	count := 0
	for i := 0; i < len(s); {
		b := s[i]
		var space bool
		if b < utf8.RuneSelf {
			space = asciiSpace[b]
			i++
		} else {
			r, size := utf8.DecodeRune(s[i:])
			space = unicode.IsSpace(r)
			i += size
		}
		if space {
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

// asciiSpace mirrors unicode.IsSpace over the ASCII range: '\t' '\n' '\v'
// '\f' '\r' and ' '.
var asciiSpace = [utf8.RuneSelf]bool{
	'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
}
