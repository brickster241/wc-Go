// wcbench measures wcGo against the system wc on generated corpora and
// prints a markdown table. Every number in the README comes from here —
// regenerate them with:
//
//	go build -o wcGo ./cmd && go run ./wcbench
//
// Corpora are deterministic (fixed seed), generated under wcbench/corpus/
// (gitignored) and reused across runs. The system wc runs with LC_ALL
// pinned to en_US.UTF-8: in the POSIX locale wc -m silently counts bytes,
// which would make the comparison flattering and false.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"time"
	"unicode/utf8"
)

const (
	dir        = "wcbench/corpus"
	asciiMB    = 512
	utf8MB     = 256
	manyFiles  = 100
	manyFileMB = 5
	runs       = 5
)

func main() {
	must(os.MkdirAll(dir, 0o755))
	ascii := gen("ascii.txt", asciiMB, genASCII)
	heavy := gen("utf8.txt", utf8MB, genUTF8)
	torture := gen("torture.txt", 64, genTorture)
	many := genMany()

	fmt.Printf("host: %s/%s, %d cores · corpora: %dMB ascii, %dMB utf-8, %d×%dMB files · median of %d\n\n",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), asciiMB, utf8MB, manyFiles, manyFileMB, runs)

	fmt.Println("| benchmark | wc | wcGo | speedup |")
	fmt.Println("|---|---|---|---|")
	for _, m := range []struct{ name, flag string }{
		{"lines (-l)", "-l"}, {"words (-w)", "-w"}, {"bytes (-c)", "-c"}, {"chars (-m)", "-m"},
	} {
		row(m.name+", ascii", m.flag, asciiMB, ascii)
	}
	for _, m := range []struct{ name, flag string }{
		{"words (-w)", "-w"}, {"chars (-m)", "-m"},
	} {
		row(m.name+", utf-8 heavy", m.flag, utf8MB, heavy)
	}
	row("chars (-m), boundary torture", "-m", 64, torture)
	rowMany(many)

	// correctness spot-check on the torture corpus — a benchmark that gets
	// the wrong answer fast is not a benchmark
	w := out("wc", "-m", torture)
	g := out("./wcGo", "-m", torture)
	fmt.Printf("\ntorture corpus agreement: wc=%q wcGo=%q — %v\n", field(w), field(g), field(w) == field(g))
}

// row times one flag over one corpus for both tools and prints the table row.
func row(name, flag string, mb int, path string) {
	wcT := median(func() { run("wc", flag, path) })
	goT := median(func() { run("./wcGo", flag, path) })
	fmt.Printf("| %s | %s | **%s** | %.2f× |\n",
		name, rate(mb, wcT), rate(mb, goT), wcT.Seconds()/goT.Seconds())
}

// rowMany is the concurrency case: wcGo fans out a goroutine per file; wc
// walks them sequentially. Identical output format, very different clock.
func rowMany(files []string) {
	wcT := median(func() { run("wc", append([]string{"-l", "-w", "-c"}, files...)...) })
	goT := median(func() { run("./wcGo", append([]string{"-l", "-w", "-c"}, files...)...) })
	mb := manyFiles * manyFileMB
	fmt.Printf("| %d×%dMB files, -l -w -c | %s | **%s** | %.2f× |\n",
		manyFiles, manyFileMB, rate(mb, wcT), rate(mb, goT), wcT.Seconds()/goT.Seconds())
}

func run(bin string, args ...string) {
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	cmd.Stdout = nil
	must(cmd.Run())
}

func out(bin string, args ...string) string {
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	b, err := cmd.Output()
	must(err)
	return string(b)
}

func field(s string) string {
	var f string
	fmt.Sscan(s, &f)
	return f
}

func median(f func()) time.Duration {
	ts := make([]time.Duration, runs)
	for i := range ts {
		t0 := time.Now()
		f()
		ts[i] = time.Since(t0)
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })
	return ts[runs/2]
}

func rate(mb int, d time.Duration) string {
	return fmt.Sprintf("%.0f MB/s", float64(mb)/d.Seconds())
}

// --- corpus generators, all seeded ---

func gen(name string, mb int, line func(*rand.Rand) []byte) string {
	path := filepath.Join(dir, name)
	if st, err := os.Stat(path); err == nil && st.Size() > int64(mb)*1024*1024-4096 {
		return path // reuse
	}
	fmt.Fprintf(os.Stderr, "generating %s (%dMB)…\n", path, mb)
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	rng := rand.New(rand.NewSource(42))
	size := 0
	for size < mb*1024*1024 {
		b := line(rng)
		n, _ := f.Write(b)
		size += n
	}
	return path
}

func genASCII(rng *rand.Rand) []byte {
	n := 60 + rng.Intn(120)
	b := make([]byte, 0, n+1)
	for i := 0; i < n; i++ {
		if rng.Intn(6) == 0 {
			b = append(b, ' ')
		} else {
			b = append(b, byte('a'+rng.Intn(26)))
		}
	}
	return append(b, '\n')
}

func genUTF8(rng *rand.Rand) []byte {
	// mixed 1–4 byte runes: latin, accents, devanagari, kana, emoji
	pools := [][]rune{
		[]rune("the quick brown fox "),
		[]rune("éàüßñçøđł "),
		[]rune("अभ्यासकरना "),
		[]rune("すばやいきつね "),
		[]rune("🚀🌍🔥📚🍀"),
	}
	b := make([]byte, 0, 200)
	for i := 0; i < 30; i++ {
		p := pools[rng.Intn(len(pools))]
		b = utf8.AppendRune(b, p[rng.Intn(len(p))])
	}
	return append(b, '\n')
}

// genTorture emits fixed-width 3-byte runes so rune boundaries land on every
// possible offset mod 32768 over the file — the chunk-boundary worst case.
func genTorture(rng *rand.Rand) []byte {
	b := make([]byte, 0, 3*1024)
	for i := 0; i < 1024; i++ {
		b = utf8.AppendRune(b, rune(0x0900+rng.Intn(0x7F))) // 3-byte Devanagari block
	}
	return append(b, '\n')
}

func genMany() []string {
	files := make([]string, manyFiles)
	for i := range files {
		files[i] = gen(fmt.Sprintf("many_%02d.txt", i), manyFileMB, genASCII)
	}
	return files
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
