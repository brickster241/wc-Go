package tests

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"os"
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
	"nospace.txt",
	"unicode.txt",
}

// flags represents the combinations of command-line flags to test with the wc command.
var flags = [][]string{
	{},     // no flags -> default wc (lines, words, bytes)
	{"-l"}, // lines only
	{"-w"}, // words only
	{"-c"}, // bytes only
	{"-m"}, // characters only
}

// randomLine generates a random line with a mix of letters, digits, punctuation, and emojis
func randomLine() string {
	length, _ := rand.Int(rand.Reader, big.NewInt(200))
	n := length.Int64() + 100 // ensure at least 100 bytes

	buf := make([]rune, n)
	for i := range buf {

		choice, _ := rand.Int(rand.Reader, big.NewInt(5))
		switch choice.Int64() {
		case 0:
			// random lowercase letter
			buf[i] = rune('a' + randByte()%26)
		case 1:
			// random uppercase letter
			buf[i] = rune('A' + randByte()%26)
		case 2:
			// random digit
			buf[i] = rune('0' + randByte()%10)
		case 3:
			// random punctuation
			punctuations := []rune{' ', '\n', '\t', '.', ',', '!', '?', ';', ':', '-', '_', '(', ')', '[', ']', '{', '}', '"', '\''}
			buf[i] = punctuations[randByte()%byte(len(punctuations))]
		case 4:
			// emoji & unicode block
			emojis := []rune{'😀', '😃', '😄', '😅', '😆', '😉', '😊', '😎', '😍', '🤖', '🚀', '🌍', '🔥', '🎉', '🍕', '📚', '🍀', '🐍'}
			buf[i] = emojis[randByte()%byte(len(emojis))]
		}
	}
	return string(buf) + "\n" // ensure file ends with newline
}

// randByte generates a random byte
func randByte() byte {
	b := make([]byte, 1)
	rand.Read(b)
	return b[0]
}

// TestLargeRandomFile generates a large random file and tests the wcGo implementation against the standard wc command.
func TestLargeRandomFile(t *testing.T) {
	wcGo := filepath.FromSlash("../wcGo") // Path to the wcGo binary

	// Create a large random file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large_random.txt")

	t.Log("Generating large test file....")
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Write random lines until we reach ~100 MB
	size := int64(0)
	target := int64(100 * 1024 * 1024) // 100 MB
	for size < target {
		line := randomLine()
		n, _ := f.WriteString(line)
		size += int64(n)
	}

	// Close the file to flush writes
	f.Close()

	t.Logf("Generated test file of size %d MB.", size/1024/1024)
	// Iterate over each combination of flags
	for _, fl := range flags {
		t.Run(testFile+" "+strings.Join(fl, " "), func(t *testing.T) {

			// run wc
			wcArgs := append(fl, testFile)
			want, err := exec.Command("wc", wcArgs...).CombinedOutput()

			if err != nil {
				t.Fatalf("wc command failed for %s with flags %v: %v", testFile, fl, err)
			}

			// run wcGo
			wcGoArgs := append(fl, testFile)
			got, err := exec.Command(wcGo, wcGoArgs...).CombinedOutput()

			if err != nil {
				t.Fatalf("wcGo command failed for %s with flags %v: %v", testFile, fl, err)
			}

			// Compare outputs
			if !bytes.Equal(want, got) {
				t.Fatalf("\nMismatch!\nwc:    %q\nwcGo:  %q", want, got)
			}
		})
	}
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
