package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain builds the wcGo binary the comparison tests exec, so a cold
// `go test ./...` works with no manual build step. The binary lands at
// ../wcGo — exactly where the tests already look — and is removed after.
func TestMain(m *testing.M) {
	bin := filepath.FromSlash("../wcGo")
	// -o is resolved relative to build.Dir (the repo root), landing beside it.
	build := exec.Command("go", "build", "-o", "wcGo", "./cmd")
	build.Dir = ".."
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building wcGo for tests: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.Remove(bin)
	os.Exit(code)
}
