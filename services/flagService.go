package services

import (
	"flag"
)

// wcCLI struct to hold the state of command-line flags and file names
type WcCLI struct {
	Words bool     // -w flag (word count)
	Lines bool     // -l flag (line count)
	Bytes bool     // -c flag (byte count)
	Chars bool     // -m flag (character count)
	Files []string // List of input files
}

// GetCLIFlags parses command-line flags and returns a wcCLI struct
func GetCLIFlags() WcCLI {

	// Define command-line flags
	words := flag.Bool("w", false, "The number of words in each input file is written to the standard output.")
	lines := flag.Bool("l", false, "The number of lines in each input file is written to the standard output")
	chars := flag.Bool("m", false, "The number of characters in each input file is written to the standard output.  If the current locale does not support multibyte characters, this is equivalent to the -c option.")
	bytes := flag.Bool("c", false, "The number of bytes in each input file is written to the standard output")

	flag.Parse() // Parse the command-line flags

	files := flag.Args() // Remaining arguments are treated as input files

	// Initialize wcCLI struct based on parsed flags
	return WcCLI{
		Words: *words,
		Lines: *lines,
		Bytes: *bytes,
		Chars: *chars,
		Files: files,
	}
}
