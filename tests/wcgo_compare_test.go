package tests

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testFiles = []string{
	"empty.txt",
	"base_test.txt",
	"simple.txt",
	"multiline.txt",
	"unicode.txt",
}

// flags represents the combinations of command-line flags to test with the wc command.
var flags = [][]string{
	{},     // no flags -> default wc (lines, words, bytes)
	{"-l"}, // lines only
	{"-w"}, // words only
	{"-c"}, // bytes only
	{"-m"}, // characters only
	{"-l", "-w", "-c"},
	{"-l", "-w", "-m"},
}

// TestWCComparison compares the output of the standard wc command with the custom wcGo implementation
func TestWCComparison(t *testing.T) {
	wcGo := filepath.FromSlash("../wcGo") // Path to the wcGo binary

	for _, file := range testFiles {
		path := filepath.Join("..", "testdata", file) // Path to the test file

		// Iterate over each combination of flags
		for _, fl := range flags {
			t.Run(file+" "+strings.Join(fl, " "), func(t *testing.T) {

				// run wc
				wcArgs := append(fl, path)
				want, err := exec.Command("wc", wcArgs...).CombinedOutput()

				if err != nil {
					t.Fatalf("wc command failed for %s with flags %v: %v", file, fl, err)
				}

				// run wcGo
				wcGoArgs := append(fl, path)
				got, err := exec.Command(wcGo, wcGoArgs...).CombinedOutput()

				if err != nil {
					t.Fatalf("wcGo command failed for %s with flags %v: %v", file, fl, err)
				}

				// Compare outputs
				if !bytes.Equal(want, got) {
					t.Fatalf("\nMismatch!\nwc:    %q\nwcGo:  %q", want, got)
				}

			})
		}
	}
}
