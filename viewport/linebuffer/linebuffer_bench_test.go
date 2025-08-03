package linebuffer

import (
	"strings"
	"testing"
)

// To run benchmarks:
// - All: go test -bench=. -benchmem -run=^$ ./viewport/linebuffer
// - Plain text only: go test -bench=BenchmarkNew_Plain -benchmem -run=^$ ./viewport/linebuffer
// - ANSI only: go test -bench=BenchmarkNew_ANSI -benchmem -run=^$ ./viewport/linebuffer
// - Unicode only: go test -bench=BenchmarkNew_Unicode -benchmem -run=^$ ./viewport/linebuffer
//
// Example of interpreting benchmark output:
// BenchmarkNew_Plain_1000-8    156124	      7883 ns/op	    8448 B/op	       3 allocs/op
// - 156124: benchmark ran 156,124 iterations to get a stable measurement
// - 7883 ns/op: each call to New() takes about 7.9 microseconds
// - 8448 B/op: each operation allocates about 8.4KB of memory
// - 3 allocs/op: each call to New() makes 3 distinct memory allocations

// BenchmarkNew_Plain benchmarks New() with plain text strings of various sizes
func BenchmarkNew_Plain_10(b *testing.B) {
	baseString := strings.Repeat("h", 10)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_Plain_100(b *testing.B) {
	baseString := strings.Repeat("h", 100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_Plain_1000(b *testing.B) {
	baseString := strings.Repeat("h", 1000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_Plain_10000(b *testing.B) {
	baseString := strings.Repeat("h", 10000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

// BenchmarkNew_ANSI benchmarks New() with ANSI-styled strings of various sizes
func BenchmarkNew_ANSI_10(b *testing.B) {
	baseString := strings.Repeat("\x1b[31mh\x1b[0m", 10)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_ANSI_100(b *testing.B) {
	baseString := strings.Repeat("\x1b[31mh\x1b[0m", 100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_ANSI_1000(b *testing.B) {
	baseString := strings.Repeat("\x1b[31mh\x1b[0m", 1000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_ANSI_10000(b *testing.B) {
	baseString := strings.Repeat("\x1b[31mh\x1b[0m", 10000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

// BenchmarkNew_Unicode benchmarks New() with Unicode strings of various sizes
func BenchmarkNew_Unicode_10(b *testing.B) {
	baseString := strings.Repeat("世", 10)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_Unicode_100(b *testing.B) {
	baseString := strings.Repeat("世", 100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_Unicode_1000(b *testing.B) {
	baseString := strings.Repeat("世", 1000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}

func BenchmarkNew_Unicode_10000(b *testing.B) {
	baseString := strings.Repeat("世", 10000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(baseString)
	}
}
