# wcGo

A high-performance, concurrent implementation of the Unix `wc` utility written in Go. wcGo processes files and streams with goroutines, efficient chunked I/O, and proper UTF-8 handling to deliver fast, accurate word, line, byte, and character counts.

## Features

- **Streaming Processing**: Processes files in 32KB chunks, enabling memory-efficient handling of arbitrarily large files without loading them entirely into memory
- **Concurrent Processing**: Leverages goroutines to process multiple files simultaneously, distributing I/O and computation across available system resources
- **Correct UTF-8 Handling**: Properly decodes and counts multi-byte UTF-8 characters (runes) with full Unicode support, including emoji and international text
- **Safe Rune Decoding Across Chunk Boundaries**: Intelligently handles incomplete UTF-8 sequences at chunk edges, carrying over partial runes to the next chunk to prevent character corruption
- **stdin Support**: Read from standard input when no files are specified, making wcGo compatible with Unix pipes and shell redirection
- **Full POSIX Compatibility**: Supports all standard flags (`-l`, `-w`, `-c`, `-m`) and produces identical output to GNU wc

## Installation

### Build from Source

Ensure you have Go 1.21 or later installed.

```bash
git clone https://github.com/brickster241/wc-Go.git
cd wc-Go
go build -o wcGo ./cmd
```

The binary will be created in the current directory.

### Add to PATH (Optional)

Make wcGo globally accessible from anywhere:

```bash
# Option 1: Copy to a system directory (requires sudo)
sudo cp wcGo /usr/local/bin/

# Option 2: Add the build directory to your PATH
export PATH="$PATH:/path/to/wc-Go"
echo 'export PATH="$PATH:/path/to/wc-Go"' >> ~/.zshrc  # for zsh
```

### Verify Installation

```bash
wcGo --help
```

## Usage

wcGo replicates the behavior of the standard Unix `wc` utility with identical command-line syntax and output formatting.

### Basic Examples

Count lines, words, and bytes in a file (default behavior):

```bash
wcGo file.txt
```

Output example:
```
      10      50     312 file.txt
```

This shows 10 lines, 50 words, and 312 bytes.

### Flag-Specific Counts

Count only lines:

```bash
wcGo -l file.txt
```

Count only words:

```bash
wcGo -w file.txt
```

Count only bytes:

```bash
wcGo -c file.txt
```

Count only characters (runes):

```bash
wcGo -m file.txt
```

Combine multiple flags to show specific metrics:

```bash
wcGo -l -w -c file.txt
```

### Multiple Files

Process multiple files concurrently; wcGo automatically parallelizes across goroutines:

```bash
wcGo file1.txt file2.txt file3.txt
```

Output will include counts for each file and a totals line:

```
      10      50     312 file1.txt
      15      75     425 file2.txt
       8      40     210 file3.txt
      33     165     947 total
```

### Reading from stdin

Use wcGo with pipes and input redirection:

```bash
cat file.txt | wcGo
```

```bash
wcGo < file.txt
```

```bash
echo "hello world" | wcGo -w
```

## Flags Reference

| Flag | Description |
|------|-------------|
| `-l` | Count lines (number of newline characters) |
| `-w` | Count words (contiguous sequences of non-whitespace characters) |
| `-c` | Count bytes (total size in bytes, not characters) |
| `-m` | Count characters (Unicode runes, accounting for multi-byte characters) |

**Default Behavior**: When no flags are specified, wcGo outputs lines, words, and bytes (equivalent to `-l -w -c`).

## Technical Details

### Chunked Streaming Architecture

wcGo does not load entire files into memory. Instead, it:

1. Reads files in **32KB chunks** using a buffered reader
2. Processes each chunk independently to compute word, line, byte, and character counts
3. Aggregates counts across chunks
4. This allows processing of arbitrarily large files with constant memory usage

### Handling Incomplete UTF-8 Runes Across Boundaries

UTF-8 is a variable-length encoding where characters can span 1–4 bytes. When a chunk boundary occurs mid-rune, wcGo:

1. Attempts to decode each byte sequence using Go's `utf8.DecodeRune()`
2. When a rune is incomplete (fewer bytes than needed for a full character), it returns `utf8.RuneError`
3. wcGo carries over incomplete bytes to the next chunk, prepending them before processing
4. This ensures that multi-byte Unicode characters spanning chunk boundaries are correctly counted without loss or corruption

### Concurrency for Multiple Files

When multiple files are provided:

1. Each file is opened and processed in a separate goroutine
2. A buffered error channel coordinates completion across goroutines
3. Results are aggregated and printed in order with a total line
4. Goroutines are not started sequentially; they all launch immediately and execute in parallel

This design maximizes throughput on multi-core systems by avoiding synchronous file I/O bottlenecks.

## Testing

wcGo includes comprehensive tests that verify correctness against GNU wc.

### Test Files

Sample test files are located in the `testdata/` directory:

- `empty.txt` – Empty file (validates edge case handling)
- `simple.txt` – Simple ASCII text
- `multiline.txt` – Multiple lines of varying lengths
- `nospace.txt` – Text without spaces (word boundary testing)
- `unicode.txt` – Mix of Unicode characters, emoji, and international text
- `base_test.txt` – Baseline test file

### Running Tests

Build wcGo first:

```bash
go build -o wcGo ./cmd
```

Run the test suite:

```bash
cd tests
go test -v
```

This will:
- Process each test file with various flag combinations
- Compare wcGo output byte-for-byte with GNU wc
- Report any mismatches

### Comparing with GNU wc

To manually verify wcGo against wc:

```bash
wc testdata/unicode.txt
wcGo testdata/unicode.txt
```

The output should be identical.

### Large File Testing

The test suite includes a `TestLargeRandomFile` that:

1. Generates a large random file (500 MB in size) with mixed content (ASCII, digits, punctuation, emoji)
2. Runs wcGo and wc on this file
3. Compares output to ensure correctness on real-world large data
4. **Automatically cleans up** the temporary file after testing

This verifies that wcGo maintains accuracy and efficiency with large datasets.

## Roadmap & Future Improvements

- **Parallel Chunk Processing**: Process chunks of a single large file in parallel (currently sequential per file)
- **Performance Metrics**: Built-in benchmarking and timing output
- **Additional Output Formats**: JSON, CSV, or TSV output options
- **Recursive Directory Processing**: `-R` flag to recursively count all files in a directory tree
- **Filtering and Exclusion**: Pattern-based file inclusion/exclusion for batch processing
- **Streaming Statistics**: Real-time progress indicators for very large files
