package app

import (
	"bufio"
	"bytes"
	"testing"
)

// 151.7 ns/op	       0 B/op	       0 allocs/op
func BenchmarkParseDate(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParseDate",
	}

	in := []byte("2023/11/24 ! (code) and then some more bytes")
	for i := 0; i < b.N; i++ {
		_, _, _ = s.ParseDate(in)
	}
}

// 20.94 ns/op	       0 B/op	       0 allocs/op
func BenchmarkParseTxPending(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParseTxPending",
		row:      0,
		col:      0,
	}
	in := []byte("! (code) and then some more bytes")
	for i := 0; i < b.N; i++ {
		s.ParseTxPending(in)
	}
}

// 42.65 ns/op	       4 B/op	       1 allocs/op
func BenchmarkParseTxCode(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParseTxCode",
		row:      0,
		col:      0,
	}
	in := []byte("(code) and some extra bytes")
	for i := 0; i < b.N; i++ {
		s.ParseTxCode(in)
	}
}

// 195.6 ns/op	      64 B/op	       2 allocs/op
func BenchmarkParseTxDesc(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParseTxDesc",
		row:      0,
		col:      0,
	}
	in := []byte("and #then some #more bytes")
	for i := 0; i < b.N; i++ {
		s.ParseTxDesc(in)
	}
}

// 81.10 ns/op	      24 B/op	       1 allocs/op
func BenchmarkParseAcctName(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParseAcctName",
	}
	in := []byte("account:name:with:depth	and then some more")
	for i := 0; i < b.N; i++ {
		s.ParseAcctName(in)
	}
}

// 20.11 ns/op	       0 B/op	       0 allocs/op
func BenchmarkParsePostNeg(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParsePostNeg",
	}
	in := []byte("- $12 CAD ; some comment")
	for i := 0; i < b.N; i++ {
		s.ParsePostNeg(in)
	}
}

// 33.49 ns/op	       0 B/op	       0 allocs/op
func BenchmarkParsePostPrefix(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParsePostPrefix",
	}
	in := []byte("$ 12 ; some comment")
	for i := 0; i < b.N; i++ {
		s.ParsePostPrefix(in)
	}
}

// 196.6 ns/op	      40 B/op	       2 allocs/op
func BenchmarkParseDecimal(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParsePostPrefix",
	}
	in := []byte("12,345.67 and tail bytes")
	for i := 0; i < b.N; i++ {
		s.ParseDecimal(in)
	}
}

// 127.7 ns/op	       4 B/op	       1 allocs/op
func BenchmarkParsePostfix(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParsePostfix",
	}
	in := []byte("NADA @ $34.54")
	for i := 0; i < b.N; i++ {
		s.ParsePostfix(in)
	}
}

// 254.1 ns/op	      40 B/op	       2 allocs/op
func BenchmarkParsePrice(b *testing.B) {
	s := Scanner{
		filename: "BenchmarkParsePrice",
	}
	in := []byte("@ $34.54")
	for i := 0; i < b.N; i++ {
		s.ParsePrice(in)
	}
}

// 96.52 ns/op	      40 B/op	       2 allocs/op
func BenchmarkFastDecimal(b *testing.B) {
	tok := []byte("123.32")
	nfractional := 2
	for i := 0; i < b.N; i++ {
		fastNewDecimal(tok, nfractional)
	}
}

// 568.7 ns/op	     368 B/op	       3 allocs/op
func BenchmarkParseTransaction(b *testing.B) {
	buf := bytes.NewBufferString("2023/11/20 ! (code) description")

	s := Scanner{
		filename: "BenchmarkParseTransaction",
		Scanner:  bufio.NewScanner(buf),
	}

	s.Scan()
	line := s.Bytes()

	for i := 0; i < b.N; i++ {
		s.ParseTransaction(line)
	}

}

// 18.86 ns/op
func BenchmarkTrimLeftSpace(b *testing.B) {
	str := bytes.NewBufferString(" some text bytes").Bytes()
	for i := 0; i < b.N; i++ {
		bytes.TrimLeft(str, " \t")
	}
}

// 6.745 ns/op
func BenchmarkTrimSpace(b *testing.B) {
	str := bytes.NewBufferString(" some text bytes").Bytes()
	for i := 0; i < b.N; i++ {
		bytes.TrimSpace(str)
	}
}

func BenchmarkParser(b *testing.B) {

	// 14917465 ns/op	 5164081 B/op	   55019 allocs/op
	// file := "./test/bench1.journal"

	// 83309 ns/op	    9648 B/op	      21 allocs/op
	file := "./test/bench2.journal"
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseJournal(file)
	}
}
