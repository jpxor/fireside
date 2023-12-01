package app

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func matchErrs(a error, b error) bool {
	if a != nil && b != nil {
		return true
	}
	if a == nil && b == nil {
		return true
	}
	return false
}

func TestParseDate(t *testing.T) {
	type Case struct {
		in   []byte
		out  time.Time
		tail []byte
		err  error
	}

	date := func(in string) (date time.Time) {
		date, _ = time.Parse("2006/01/02", in)
		return
	}

	printDate := func(in time.Time) string {
		return in.Format("2006/01/02")
	}

	cases := []Case{
		{in: []byte(""), out: NotDate, tail: []byte(""), err: nil},
		{in: []byte("not_date"), out: NotDate, tail: []byte("not_date"), err: nil},
		{in: []byte("2023/11/24"), out: date("2023/11/24"), tail: []byte(""), err: nil},
		{in: []byte("2023-11-24"), out: date("2023/11/24"), tail: []byte(""), err: nil},
		{in: []byte("2023/11/24 tailing bytes"), out: date("2023/11/24"), tail: []byte("tailing bytes"), err: nil},
		{in: []byte("2023/24/11 bad date"), out: NotDate, tail: []byte(""), err: fmt.Errorf("bad date")},
		{in: []byte("2023/11/24nospace"), out: NotDate, tail: []byte(""), err: fmt.Errorf("no space")},
	}

	s := Scanner{
		filename: "TestParseDate",
		row:      0,
		col:      0,
	}

	for _, test := range cases {
		s.row += 1
		s.col = 0

		out, tail, err := s.ParseDate(test.in)

		if out.Compare(test.out) != 0 {
			t.Errorf("dates do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got date: %s\n", printDate(out))
			fmt.Printf("expected: %s\n", printDate(test.out))
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParseTxPending(t *testing.T) {
	type Case struct {
		in   []byte
		out  bool
		tail []byte
		err  error
	}

	cases := []Case{
		{in: []byte(""), out: false, tail: []byte(""), err: nil},
		{in: []byte("some bytes"), out: false, tail: []byte("some bytes"), err: nil},
		{in: []byte("!nospace"), out: false, tail: []byte(""), err: fmt.Errorf("no space")},
		{in: []byte("!"), out: true, tail: []byte(""), err: nil},
		{in: []byte("! some bytes"), out: true, tail: []byte("some bytes"), err: nil},
	}

	s := Scanner{
		filename: "TestParseTxPending",
		row:      0,
		col:      0,
	}

	for _, test := range cases {
		s.row += 1
		s.col = 0

		out, tail, err := s.ParseTxPending(test.in)

		if out != test.out {
			t.Errorf("pendings do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got date: %t\n", out)
			fmt.Printf("expected: %t\n", test.out)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParseTxCode(t *testing.T) {
	type Case struct {
		in   []byte
		out  string
		tail []byte
		err  error
	}

	cases := []Case{
		{in: []byte(""), out: "", tail: []byte(""), err: nil},
		{in: []byte("some bytes"), out: "", tail: []byte("some bytes"), err: nil},
		{in: []byte("(c)nospace"), out: "", tail: []byte(""), err: fmt.Errorf("no space")},
		{in: []byte("()"), out: "", tail: []byte(""), err: nil},
		{in: []byte("(c)"), out: "c", tail: []byte(""), err: nil},
		{in: []byte("(code)"), out: "code", tail: []byte(""), err: nil},
		{in: []byte("(code with space)"), out: "code with space", tail: []byte(""), err: nil},
		{in: []byte("(code) with tail"), out: "code", tail: []byte("with tail"), err: nil},
	}

	s := Scanner{
		filename: "TestParseTxCode",
		row:      0,
		col:      0,
	}

	for _, test := range cases {
		s.row += 1
		s.col = 0

		out, tail, err := s.ParseTxCode(test.in)

		if out != test.out {
			t.Errorf("pendings do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got date: %s\n", out)
			fmt.Printf("expected: %s\n", test.out)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParseTxDesc(t *testing.T) {
	type Case struct {
		in   []byte
		out  string
		tags []string
		err  error
	}

	cases := []Case{
		{in: []byte(""), out: "", tags: nil, err: nil},
		{in: []byte("description"), out: "description", tags: nil, err: nil},
		{in: []byte("#tag"), out: "#tag", tags: []string{"tag"}, err: nil},
		{in: []byte("description #tag"), out: "description #tag", tags: []string{"tag"}, err: nil},
		{in: []byte("#tag description"), out: "#tag description", tags: []string{"tag"}, err: nil},
		{in: []byte("desc #tag desc"), out: "desc #tag desc", tags: []string{"tag"}, err: nil},
		{in: []byte("#a #b #c"), out: "#a #b #c", tags: []string{"a", "b", "c"}, err: nil},
		{in: []byte("#tag, desc"), out: "#tag, desc", tags: []string{"tag"}, err: nil},
	}

	s := Scanner{
		filename: "TestParseTxDesc",
		row:      0,
		col:      0,
	}

	for _, test := range cases {
		s.row += 1
		s.col = 0

		out, tags, err := s.ParseTxDesc(test.in)

		if out != test.out {
			t.Errorf("descriptions do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got desc: %s\n", out)
			fmt.Printf("expected: %s\n", test.out)
		}

		if !reflect.DeepEqual(tags, test.tags) {
			t.Errorf("tags do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tags: %s\n", tags)
			fmt.Printf("expected: %s\n", test.tags)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParseAcctName(t *testing.T) {
	type Case struct {
		in   []byte
		out  string
		tail []byte
		err  error
	}

	cases := []Case{
		{in: []byte(""), out: "", tail: []byte(""), err: nil},
		{in: []byte("name"), out: "name", tail: []byte(""), err: nil},
		{in: []byte("name\n"), out: "name", tail: []byte(""), err: nil},
		{in: []byte("name\r\n"), out: "name", tail: []byte(""), err: nil},
		{in: []byte("name\t"), out: "name", tail: []byte(""), err: nil},
		{in: []byte("name \t"), out: "name", tail: []byte(""), err: nil},
		{in: []byte("name\t "), out: "name", tail: []byte(""), err: nil},
		{in: []byte("name name"), out: "name name", tail: []byte(""), err: nil},
		{in: []byte("name\ttail"), out: "name", tail: []byte("tail"), err: nil},
		{in: []byte("name\t\ttail"), out: "name", tail: []byte("tail"), err: nil},
		{in: []byte("name  tail"), out: "name", tail: []byte("tail"), err: nil},
		{in: []byte("name \ttail"), out: "name", tail: []byte("tail"), err: nil},
	}

	s := Scanner{
		filename: "TestParseAcctName",
		row:      0,
		col:      0,
	}

	for i, test := range cases {
		s.row += 1
		s.col = 0

		out, tail, err := s.ParseAcctName(test.in)

		if out != test.out {
			t.Errorf("names do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got name: '%s'\n", out)
			fmt.Printf("expected: '%s'\n", test.out)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParsePostNeg(t *testing.T) {
	type Case struct {
		in   []byte
		out  bool
		tail []byte
		err  error
	}

	cases := []Case{
		{in: []byte(""), out: false, tail: []byte(""), err: nil},
		{in: []byte("-"), out: true, tail: []byte(""), err: nil},
		{in: []byte("- $12"), out: true, tail: []byte("$12"), err: nil},
		{in: []byte("-$12"), out: true, tail: []byte("$12"), err: nil},
		{in: []byte("$12"), out: false, tail: []byte("$12"), err: nil},
	}

	s := Scanner{
		filename: "TestParsePostNeg",
		row:      0,
		col:      0,
	}

	for _, test := range cases {
		s.row += 1
		s.col = 0

		out, tail, err := s.ParsePostNeg(test.in)

		if out != test.out {
			t.Errorf("outputs do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got neg : %t\n", out)
			fmt.Printf("expected: %t\n", test.out)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match")
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParsePostPrefix(t *testing.T) {
	type Case struct {
		in   []byte
		out  string
		tail []byte
		err  error
	}

	cases := []Case{
		{in: []byte(""), out: "", tail: []byte(""), err: nil},
		{in: []byte("12"), out: "", tail: []byte("12"), err: nil},
		{in: []byte("$"), out: "$", tail: []byte(""), err: nil},
		{in: []byte("$12"), out: "$", tail: []byte("12"), err: nil},
		{in: []byte("R$12"), out: "R$", tail: []byte("12"), err: nil},
		{in: []byte("¥12"), out: "¥", tail: []byte("12"), err: nil},
		{in: []byte("kr.12"), out: "kr.", tail: []byte("12"), err: nil},
		{in: []byte("₹12"), out: "₹", tail: []byte("12"), err: nil},
		{in: []byte("₪12"), out: "₪", tail: []byte("12"), err: nil},
		{in: []byte("₩12"), out: "₩", tail: []byte("12"), err: nil},
		{in: []byte("RM12"), out: "RM", tail: []byte("12"), err: nil},
		{in: []byte("kr12"), out: "kr", tail: []byte("12"), err: nil},
		{in: []byte("₱12"), out: "₱", tail: []byte("12"), err: nil},
		{in: []byte("R12"), out: "R", tail: []byte("12"), err: nil},
		{in: []byte("fr.12"), out: "fr.", tail: []byte("12"), err: nil},
		{in: []byte("元12"), out: "元", tail: []byte("12"), err: nil},
		{in: []byte("£12"), out: "£", tail: []byte("12"), err: nil},
		{in: []byte("₿12"), out: "₿", tail: []byte("12"), err: nil},
		{in: []byte("$ 12"), out: "$", tail: []byte("12"), err: nil},
		{in: []byte("R$ 12"), out: "R$", tail: []byte("12"), err: nil},
		{in: []byte("¥ 12"), out: "¥", tail: []byte("12"), err: nil},
		{in: []byte("kr 12"), out: "kr", tail: []byte("12"), err: nil},
		{in: []byte("$ 12 USD"), out: "$", tail: []byte("12 USD"), err: nil},
		{in: []byte("12$ CAD"), out: "", tail: []byte("12$ CAD"), err: nil},
	}

	s := Scanner{
		filename: "TestParsePostPrefix",
		row:      0,
		col:      0,
	}

	for i, test := range cases {
		s.row += 1
		s.col = 0

		out, tail, err := s.ParsePostPrefix(test.in)

		if out != test.out {
			t.Errorf("symbols do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got sym : '%s'\n", out)
			fmt.Printf("expected: '%s'\n", test.out)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParsePostAmount(t *testing.T) {
	type Case struct {
		in     []byte
		out    decimal.Decimal
		decsym string
		tail   []byte
		err    error
	}

	amount := func(s string) decimal.Decimal {
		amount, _ := decimal.NewFromString(s)
		return amount
	}

	cases := []Case{
		{in: []byte(""), out: decimal.Zero, decsym: "", tail: []byte(""), err: fmt.Errorf("bad format")},
		{in: []byte("1"), out: amount("1"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("1otherbytes"), out: amount("1"), decsym: "", tail: []byte("otherbytes"), err: nil},
		{in: []byte("1 otherbytes"), out: amount("1"), decsym: "", tail: []byte("otherbytes"), err: nil},
		{in: []byte("01"), out: amount("1"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10"), out: amount("10"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10.0"), out: amount("10"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10.10"), out: amount("10.1"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10,10"), out: amount("10.1"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10,101"), out: amount("10101"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10_101"), out: amount("10101"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10,101.99"), out: amount("10101.99"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("10,101.99"), out: amount("10101.99"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("111_222_333"), out: amount("111222333"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("4.567"), out: amount("4567"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("4.567,89"), out: amount("4567.89"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("4.567,890"), out: amount("4567890"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("4 $ 56"), out: amount("4.56"), decsym: " $ ", tail: []byte(""), err: nil},
		{in: []byte("4$56"), out: amount("4.56"), decsym: "$", tail: []byte(""), err: nil},
		{in: []byte("4 ₿ 56"), out: amount("4.56"), decsym: " ₿ ", tail: []byte(""), err: nil},
		{in: []byte("44.23.56"), out: amount("4423.56"), decsym: "", tail: []byte(""), err: nil},
		{in: []byte("69 AAA @ $10"), out: amount("69"), decsym: "", tail: []byte("AAA @ $10"), err: nil},
		{in: []byte("no value"), out: decimal.Zero, decsym: "", tail: []byte(""), err: fmt.Errorf("bad format")},
		{in: []byte("4 56"), out: decimal.Zero, decsym: "", tail: []byte(""), err: fmt.Errorf("bad format")},
	}

	s := Scanner{
		filename: "ParseDecimal",
		row:      0,
		col:      0,
	}

	for i, test := range cases {
		s.row += 1
		s.col = 0

		out, decsym, tail, err := s.ParseDecimal(test.in)

		if !out.Equal(test.out) {
			t.Errorf("values do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got valu : '%s'\n", out)
			fmt.Printf("expected: '%s'\n", test.out)
		}

		if decsym != test.decsym {
			t.Errorf("symbols do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got sym : '%s'\n", decsym)
			fmt.Printf("expected: '%s'\n", test.decsym)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}
	}
}

func TestParsePostfix(t *testing.T) {
	type Case struct {
		in   []byte
		sym  string
		code string
		tail []byte
	}

	cases := []Case{
		{in: []byte(""), sym: "", code: "", tail: []byte("")},
		{in: []byte("$"), sym: "$", code: "", tail: []byte("")},
		{in: []byte("Kč"), sym: "Kč", code: "", tail: []byte("")},
		{in: []byte("€"), sym: "€", code: "", tail: []byte("")},
		{in: []byte("Ft"), sym: "Ft", code: "", tail: []byte("")},
		{in: []byte(".د.م."), sym: ".د.م.", code: "", tail: []byte("")},
		{in: []byte("zł"), sym: "zł", code: "", tail: []byte("")},
		{in: []byte("p."), sym: "p.", code: "", tail: []byte("")},
		{in: []byte("﷼"), sym: "﷼", code: "", tail: []byte("")},
		{in: []byte("kr"), sym: "kr", code: "", tail: []byte("")},
		{in: []byte("฿"), sym: "฿", code: "", tail: []byte("")},
		{in: []byte("₺"), sym: "₺", code: "", tail: []byte("")},
		{in: []byte("₫"), sym: "₫", code: "", tail: []byte("")},
		{in: []byte("R$"), sym: "R$", code: "", tail: []byte("")},
		{in: []byte("USD"), sym: "", code: "USD", tail: []byte("")},
		{in: []byte("$ CAD"), sym: "$", code: "CAD", tail: []byte("")},
		{in: []byte("€ EUR"), sym: "€", code: "EUR", tail: []byte("")},
		{in: []byte("€ EUR and then more"), sym: "€", code: "EUR", tail: []byte("and then more")},
		{in: []byte("AAPL @ 28.23$ CAD"), sym: "", code: "AAPL", tail: []byte("@ 28.23$ CAD")},
		{in: []byte("BB @ $4.99 CAD"), sym: "", code: "BB", tail: []byte("@ $4.99 CAD")},
		{in: []byte("C @ 45.23USD"), sym: "", code: "C", tail: []byte("@ 45.23USD")},
		{in: []byte("shoes"), sym: "", code: "shoes", tail: []byte("")},
		{in: []byte("\"quoted name\""), sym: "", code: "\"quoted name\"", tail: []byte("")},
		{in: []byte("'quoted name'"), sym: "", code: "'quoted name'", tail: []byte("")},
	}

	s := Scanner{
		filename: "TestParsePostPrefix",
		row:      0,
		col:      0,
	}

	for i, test := range cases {
		s.row += 1
		s.col = 0

		sym, code, tail := s.ParsePostfix(test.in)

		if sym != test.sym {
			t.Errorf("symbols do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got sym : '%s'\n", sym)
			fmt.Printf("expected: '%s'\n", test.sym)
		}

		if code != test.code {
			t.Errorf("codes do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got code: '%s'\n", code)
			fmt.Printf("expected: '%s'\n", test.code)
		}

		if !bytes.Equal(tail, test.tail) {
			t.Errorf("tails do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got tail: %s\n", tail)
			fmt.Printf("expected: %s\n", test.tail)
		}
	}
}

func TestParsePrice(t *testing.T) {
	type Case struct {
		in      []byte
		val     decimal.Decimal
		Type    CommodityType
		perUnit bool
		tail    []byte
		err     error
	}

	amount := func(s string) decimal.Decimal {
		amount, _ := decimal.NewFromString(s)
		return amount
	}

	cases := []Case{
		{in: []byte(""), val: decimal.Zero, tail: []byte("")},
		{in: []byte("text"), val: decimal.Zero, err: fmt.Errorf("bad format")},
		{in: []byte("@text"), val: decimal.Zero, err: fmt.Errorf("bad format")},
		{in: []byte("@@text"), val: decimal.Zero, err: fmt.Errorf("bad format")},
		{in: []byte("@1"), val: amount("1"), perUnit: true},
		{in: []byte("@@1"), val: amount("1"), perUnit: false},
		{in: []byte("@ 1"), val: amount("1"), perUnit: true},
		{in: []byte("@@ 1"), val: amount("1"), perUnit: false},
		{in: []byte("@ $1"), val: amount("1"), perUnit: true, Type: CommodityType{Prefix: "$"}},
		{in: []byte("@@ $1"), val: amount("1"), perUnit: false, Type: CommodityType{Prefix: "$"}},
		{in: []byte("@ 1$"), val: amount("1"), perUnit: true, Type: CommodityType{Postfix: "$"}},
		{in: []byte("@@ 1$"), val: amount("1"), perUnit: false, Type: CommodityType{Postfix: "$"}},
	}

	s := Scanner{
		filename: "TestParsePrice",
		row:      0,
		col:      0,
	}

	for i, test := range cases {
		s.row += 1
		s.col = 0

		val, Type, perUnit, tail, err := s.ParsePrice(test.in)

		if !matchErrs(err, test.err) {
			t.Errorf("errs do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.in)
			fmt.Printf("got err : %v\n", err)
			fmt.Printf("expected: %v\n", test.err)
		}

		if err == nil {

			if !val.Equal(test.val) {
				t.Errorf("values do not match (#%d)", i)
				fmt.Printf("in      : %s\n", test.in)
				fmt.Printf("got valu: '%s'\n", val)
				fmt.Printf("expected: '%s'\n", test.val)
			}

			if !reflect.DeepEqual(Type, test.Type) {
				t.Errorf("types do not match (#%d)", i)
				fmt.Printf("in      : %s\n", test.in)
				fmt.Printf("got type: '%+v'\n", Type)
				fmt.Printf("expected: '%+v'\n", test.Type)
			}

			if perUnit != test.perUnit {
				t.Errorf("perunits do not match (#%d)", i)
				fmt.Printf("in      : %s\n", test.in)
				fmt.Printf("got peru: '%+v'\n", perUnit)
				fmt.Printf("expected: '%+v'\n", test.perUnit)
			}

			if !bytes.Equal(tail, test.tail) {
				t.Errorf("tails do not match (#%d)", i)
				fmt.Printf("in      : %s\n", test.in)
				fmt.Printf("got tail: %s\n", tail)
				fmt.Printf("expected: %s\n", test.tail)
			}

			if !matchErrs(err, test.err) {
				t.Errorf("errs do not match (#%d)", i)
				fmt.Printf("in      : %s\n", test.in)
				fmt.Printf("got err : %v\n", err)
				fmt.Printf("expected: %v\n", test.err)
			}
		}
	}
}

func TestFastNewDecimal(t *testing.T) {

	type Case struct {
		token       string
		nfractional int
	}

	// decimal.NewFromString doesn't accept
	// any symbols outside of digits and
	// the decimal point
	filter := regexp.MustCompile("[^0-9.]")

	cases := []Case{
		{"123", 0},
		{"123.4", 1},
		{"123.45", 2},
		{"123.456", 3},
		{"1,234", 0},
		{"1,234.5", 1},
		{"1,234.56", 2},
		{"1,234.567", 3},
		{"1,234,567", 0},
	}

	for i, test := range cases {
		expected, err := decimal.NewFromFormattedString(test.token, filter)
		dec := fastNewDecimal([]byte(test.token), test.nfractional)

		if err != nil {
			t.Errorf("failed to get expected value (#%d)", i)
		}

		if !expected.Equal(dec) {
			t.Errorf("values do not match (#%d)", i)
			fmt.Printf("in      : %s\n", test.token)
			fmt.Printf("got valu: '%s'\n", dec)
			fmt.Printf("expected: '%s'\n", expected)
		}
	}
}

func TestErrJournal(t *testing.T) {
	file := "./test/err.journal"
	_, _, err := ParseJournal(file)

	if err == nil {
		t.Errorf("umm... was expecting more")
		return
	}
	fmt.Println(err)
}

func TestSanity(t *testing.T) {
	file := "./test/test.journal"
	_, txs, err := ParseJournal(file)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	date := func(str string) (d time.Time) {
		d, _ = time.Parse("2006/01/02", str)
		return
	}

	// there are 2 transaction in the file, but they should
	// both produce the same tx struct:
	expectedTx := Transaction{
		Date:        date("2023/11/20"),
		Description: "This is the tx description",
		Postings: []Posting{
			{
				Account: "assets:cash",
				Amount:  fastNewDecimal([]byte("420"), 0),
				Commodity: Commodity{
					Type: CommodityType{
						Prefix: "$",
					},
				},
			}, {
				Account: "income:employer",
				Amount:  fastNewDecimal([]byte("420"), 0).Neg(),
				Commodity: Commodity{
					Type: CommodityType{
						Prefix: "$",
					},
				},
			}, {
				Account: "stocks in",
				Amount:  fastNewDecimal([]byte("69"), 0),
				Commodity: Commodity{
					Type: CommodityType{
						Code: "AAA",
					},
					BookValue: fastNewDecimal([]byte("10"), 0),
					ValueType: CommodityType{
						Prefix: "$",
					},
				},
			}, {
				Account: "stocks out",
				Amount:  fastNewDecimal([]byte("69"), 0).Neg(),
				Commodity: Commodity{
					Type: CommodityType{
						Code: "AAA",
					},
				},
			},
		},
	}

	// postings must be sorted, so order does not affect the
	// results
	sort.Slice(expectedTx.Postings, func(i, j int) bool {
		si := expectedTx.Postings[i].Account
		sj := expectedTx.Postings[j].Account
		return strings.Compare(si, sj) < 0
	})

	for _, tx := range txs {

		sort.Slice(tx.Postings, func(i, j int) bool {
			si := tx.Postings[i].Account
			sj := tx.Postings[j].Account
			return strings.Compare(si, sj) < 0
		})

		// allow to check visually
		fmt.Println(tx)

		if !expectedTx.Date.Equal(tx.Date) {
			t.Error("dates don't match")
		}

		if expectedTx.Description != tx.Description {
			t.Error("descriptions don't match")
		}

		if len(expectedTx.Postings) != len(tx.Postings) {
			t.Error("len postins don't match")
		}

		for i := 0; i < len(tx.Postings); i++ {
			post := tx.Postings[i]
			expectedPost := expectedTx.Postings[i]

			if expectedPost.Account != post.Account {
				t.Error("account names don't match")
			}

			if !expectedPost.Amount.Equal(post.Amount) {
				t.Error("post amounts don't match")
			}

			if !expectedPost.Commodity.BookValue.Equal(post.Commodity.BookValue) {
				t.Error("book values don't match")
			}

			if expectedPost.Commodity.Type != post.Commodity.Type {
				t.Error("commodity types don't match")
			}
		}
	}

}
