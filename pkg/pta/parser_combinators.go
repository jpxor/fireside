package pta

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// attempted to write a combinator parser...
// it got complicated and hard to debug really quick,
// and didn't perform as well as the procedural approach.
//
// but I kept some of the pieces that might come in
// handy later on

var ErrNoMatch = fmt.Errorf("no match")
var ErrEOF = fmt.Errorf("unexpected end of file")

type TokenScanner func(in []byte) (tok []byte, tail []byte, err error)

func OR(scanners ...TokenScanner) TokenScanner {
	return func(in []byte) (tok, tail []byte, err error) {
		for _, scanner := range scanners {
			tok, tail, err = scanner(in)
			if err == nil {
				return
			}
		}
		return nil, nil, ErrNoMatch
	}
}

func SEQ(scanners ...TokenScanner) TokenScanner {
	return func(in []byte) ([]byte, []byte, error) {
		var tail []byte = in
		var tok []byte
		var err error
		var i int
		for _, scanner := range scanners {
			tok, tail, err = scanner(tail)
			if err != nil {
				return nil, nil, err
			}
			i += len(tok)
		}
		return in[:i], in[i:], nil
	}
}

func UNTIL(scanner TokenScanner) TokenScanner {
	return func(in []byte) (tok []byte, tail []byte, err error) {
		for i := 0; i < len(in); i++ {
			tok, tail, err = scanner(in[i:])
			if err == ErrEOF {
				return
			}
			if err == ErrNoMatch {
				continue
			}
			tok = in[:i]
			tail = in[i:]
			return
		}
		err = ErrEOF
		return
	}
}

func CAPTURE(left, right TokenScanner) TokenScanner {
	return func(in []byte) (tok []byte, tail []byte, err error) {
		var ltok []byte
		var rtok []byte
		var newtail []byte
		ltok, tail, err = left(in)
		if err != nil {
			return
		}
		for {
			rtok, newtail, err = right(tail)
			if err == ErrEOF {
				return
			}
			if err == ErrNoMatch {
				tail = tail[1:]
				continue
			}
			tail = newtail
			break
		}
		i := len(ltok)
		j := len(in) - len(tail) - len(rtok)
		tok = in[i:j]
		return
	}
}

func ScanNDigits(n int) TokenScanner {
	return func(bytes []byte) ([]byte, []byte, error) {
		if len(bytes) < n {
			return nil, nil, ErrEOF
		}
		for i := 0; i < n; i++ {
			r := rune(bytes[i])
			if !unicode.IsDigit(r) {
				return nil, nil, ErrNoMatch
			}
		}
		return bytes[:n], bytes[n:], nil
	}
}

func ScanByte(b byte) TokenScanner {
	return func(bytes []byte) ([]byte, []byte, error) {
		if len(bytes) == 0 {
			return nil, nil, ErrEOF
		}
		if bytes[0] == b {
			return bytes[0:1], bytes[1:], nil
		}
		return nil, nil, ErrNoMatch
	}
}

func ScanByteFrom(matches []byte) TokenScanner {
	return func(bytes []byte) ([]byte, []byte, error) {
		if len(bytes) == 0 {
			return nil, nil, ErrEOF
		}
		for i := 0; i < len(matches); i++ {
			if bytes[0] == matches[i] {
				return bytes[0:1], bytes[1:], nil
			}
		}
		return nil, nil, ErrNoMatch
	}
}

func ScanSpace(bytes []byte) ([]byte, []byte, error) {
	if len(bytes) == 0 {
		return nil, nil, ErrEOF
	}
	r, w := utf8.DecodeRune(bytes)
	if !unicode.IsSpace(r) {
		return nil, nil, ErrNoMatch
	}
	for i := w; i < len(bytes); i += w {
		r, w = utf8.DecodeRune(bytes[i:])
		if !unicode.IsSpace(r) {
			return bytes[:i], bytes[i:], nil
		}
	}
	return bytes, nil, nil
}

var validSeparators = []byte{'-', '/', '.'}
var scanDate = SEQ(
	ScanNDigits(4),
	ScanByteFrom(validSeparators),
	ScanNDigits(2),
	ScanByteFrom(validSeparators),
	ScanNDigits(2),
)

// took ~17% longer to parse a date using this matcher
func MatchDate_unused(in []byte) bool {
	_, _, err := scanDate(in)
	return err == nil
}
