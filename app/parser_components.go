package app

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

var NotDate time.Time

const START_OF_COMMENT byte = ';'

// extend the scanner to add helpful state about the
// file being scanned (helps write nice parse errors)
type Scanner struct {
	*bufio.Scanner
	filename string
	row      int
	col      int
}

func (s *Scanner) Scan() bool {
	s.row += 1
	s.col = 0
	return s.Scanner.Scan()
}

// the caller must guarantee advancing is valid
func (s *Scanner) advance(line []byte, i int) (tok, tail []byte) {
	tok = line[:i]
	tail = bytes.TrimSpace(line[i:])
	s.col += len(line) - len(tail)
	return
}

func (s *Scanner) wrap(err error) error {
	return fmt.Errorf("%s:%d:%d: %s", s.filename, s.row, s.col, err)
}

func isDigit(r byte) bool {
	return unicode.IsDigit(rune(r))
}

func tidy(line []byte) (retLine []byte, empty, hadComment bool) {
	if i := bytes.IndexByte(line, START_OF_COMMENT); i != -1 {
		line = line[:i]
		hadComment = true
	}
	retLine = line
	empty = len(bytes.TrimSpace(line)) == 0
	return
}

// Note: the decimal.NewFromString function was slow and required
// an alloc to convert bytes to a string. Here we calculate the
// significand and exponential components of the decimal ourselves
// without a string conversion. Benchmarks show we tripled the
// speed and halved the number of allocations
func fastNewDecimal(tok []byte, nfractional int) decimal.Decimal {
	parseSignificand := func(tok []byte) (value int64) {
		// non-digit characters can be found in the token, these
		// are the thousandths and decimal separators which must
		// be skipped when computing the significand
		for i := 0; i < len(tok); i++ {
			if unicode.IsDigit(rune(tok[i])) {
				value = 10*value + int64(tok[i]-'0')
			}
		}
		return value
	}
	significand := parseSignificand(tok)
	exp := int32(-nfractional)
	return decimal.New(significand, exp)
}

func matchDate(tok []byte) bool {
	return len(tok) >= 10 &&
		isDigit(tok[0]) && isDigit(tok[1]) && isDigit(tok[2]) && isDigit(tok[3]) && // year
		!isDigit(tok[4]) && // seperator
		isDigit(tok[5]) && isDigit(tok[6]) && // month
		!isDigit(tok[7]) && // seperator
		isDigit(tok[8]) && isDigit(tok[9]) // day
}

// Date must have the format: YYYY/MM/DD, but it can use any not digit separator
// The date is the first token in a transaction line. Failing to match the date
// format simply means the line does not belong to a transaction, not an error
func (s *Scanner) ParseDate(tok []byte) (out time.Time, tail []byte, err error) {
	if !matchDate(tok) {
		return NotDate, tok, ErrNoMatch
	}
	if len(tok) > 10 { // must be followed by space, tab, or newline
		if !unicode.IsSpace(rune(tok[10])) {
			s.advance(tok, 10)
			return NotDate, []byte{}, s.wrap(fmt.Errorf("date must be followed by space or newline"))
		}
	}
	tok[4] = '/'
	tok[7] = '/'
	out, err = time.Parse("2006/01/02", string(tok[:10]))
	if err != nil {
		return NotDate, []byte{}, s.wrap(err)
	}
	_, tail = s.advance(tok, 10)
	return
}

// optional: '!' means the transaction is pending
func (s *Scanner) ParseTxPending(tok []byte) (out bool, tail []byte, err error) {
	if len(tok) > 0 && tok[0] == '!' {
		if len(tok) > 1 {
			// must be followed by space, tab, or newline
			if !unicode.IsSpace(rune(tok[1])) {
				s.advance(tok, 1)
				return false, []byte{}, s.wrap(fmt.Errorf("'!' must be followed by space or newline"))
			}
		}
		out = true
		_, tail = s.advance(tok, 1)
		return
	}
	return false, tok, nil
}

// optional:  any string within brakets: (code)
func (s *Scanner) ParseTxCode(tok []byte) (out string, tail []byte, err error) {
	if len(tok) > 1 && tok[0] == '(' {
		lbi := bytes.IndexByte(tok, ')')
		if lbi == -1 {
			return "", []byte{}, s.wrap(fmt.Errorf("missing closing bracket ')"))
		}
		if len(tok) > lbi+1 {
			// must be followed by space, tab, or newline
			if !unicode.IsSpace(rune(tok[lbi+1])) {
				s.advance(tok, lbi)
				return "", []byte{}, s.wrap(fmt.Errorf("'(code)' must be followed by space or newline"))
			}
		}
		out = string(tok[1:lbi])
		_, tail = s.advance(tok, len(out)+2)
		return
	}
	return "", tok, nil
}

// optional: the rest of the line makes up the description (tidy line to rm comments!)
// but #tags embeded in the description are also extracted
func (s *Scanner) ParseTxDesc(tok []byte) (out string, tags []string, err error) {
	out = string(tok)

	tmp := out
	if tagCount := strings.Count(out, "#"); tagCount > 0 {
		tags = make([]string, tagCount)

		for tid := 0; tid < tagCount; tid++ {
			i := strings.IndexByte(tmp, '#')
			tmp = tmp[i+1:]
			end := 0
			for end < len(tmp) &&
				!unicode.IsSpace(rune(tmp[end])) &&
				!unicode.IsPunct(rune(tmp[end])) {
				end++
			}
			tags[tid] = tmp[:end]
			tmp = tmp[end:]
		}
	}
	return
}

// required: all postings following the first transaction line
// must start with space
func (s *Scanner) ParseIndent(tok []byte) (tail []byte, err error) {
	if (len(tok) > 0) &&
		unicode.IsSpace(rune(tok[0])) {
		_, tail = s.advance(tok, 1)
		return
	}
	tail = tok
	err = s.wrap(fmt.Errorf("bad format: expected indent or newline"))
	return
}

// account name ends with double space, tab, or EOL
// the name is allowed to contan spaces
func (s *Scanner) ParseAcctName(tok []byte) (out string, tail []byte, err error) {

	endOfNameIndex := func(line []byte) int {
		for i := 1; i < len(line); i++ {
			if line[i-1] == ' ' && (line[i] == ' ' || line[i] == '\t') {
				return i - 1
			}
			if line[i] == '\n' || line[i] == '\t' || line[i] == '\r' {
				return i
			}
		}
		return len(line)
	}

	i := endOfNameIndex(tok)
	tok, tail = s.advance(tok, i)
	out = string(tok)
	return
}

// optional: have of postings will have a '-' symbol showing the flow of
// money out of the account
func (s *Scanner) ParsePostNeg(tok []byte) (out bool, tail []byte, err error) {
	if len(tok) > 0 && tok[0] == '-' {
		out = true
		_, tail = s.advance(tok, 1)
		return
	}
	return false, tok, nil
}

// optional: any string preceeding the first digit of the amount will be
// captured as a prefix
func (s *Scanner) ParsePostPrefix(tok []byte) (out string, tail []byte, err error) {
	var r rune
	var w int
	var i int
	for i = 0; i < len(tok); i += w {
		r, w = utf8.DecodeRune(tok[i:])
		if unicode.IsSpace(r) || unicode.IsDigit(r) {
			break
		}
	}
	tok, tail = s.advance(tok, i)
	out = string(tok)
	return
}

func (s *Scanner) ParseDecimal(line []byte) (out decimal.Decimal, decsym string, tail []byte, err error) {
	tok := line

	// must ignore commodity price if present
	r := bytes.Index(line, []byte{'@'})
	if r != -1 {
		tok = line[:r]
	}

	r = bytes.LastIndexFunc(tok, unicode.IsDigit)
	if r == -1 {
		err = s.wrap(fmt.Errorf("failed for parse decimal: '%s'", tok))
		return
	}
	tok, tail = s.advance(line, r+1)

	// number of digits to the right of decimal separator
	nfractional := 0

	// potential index of the separator
	r = bytes.LastIndexFunc(tok, func(r rune) bool {
		return !unicode.IsDigit(r)
	})

	if r != -1 {
		// assume there are only 1 or 2 digits to the right of the decimal sep (.00)
		// therefore, any more and we assume its a thousandths sep and can ignore it
		if len(tok)-r <= 3 {
			nfractional = len(tok) - r - 1

			// decimal separator may be a currency symbol !
			// and a symbol may span more than one byte
			if !unicode.IsPunct(rune(tok[r])) {
				l := 1 + bytes.LastIndexFunc(tok[:r], unicode.IsDigit)

				// but it cant be empty
				if len(bytes.TrimSpace(tok[l:r+1])) == 0 {
					err = s.wrap(fmt.Errorf("bad format: space in value '%s'", tok))
					return
				}
				decsym = string(tok[l : r+1])
			}
		}
	}

	out = fastNewDecimal(tok, nfractional)
	return
}

// here's a complex one, maybe requires rethink, or refactor?
//
// optional: the postfix should end at the end of the line, or at '@' symbol
func (s *Scanner) ParsePostfix(tok []byte) (sym string, code string, tail []byte) {

	isAllUpper := func(b []byte) (ok bool) {
		ok = true
		for i := 0; i < len(b); i++ {
			ok = ok && unicode.IsUpper(rune(b[i]))
		}
		return
	}

	isAllLetter := func(b []byte) (ok bool) {
		ok = true
		for i := 0; i < len(b); i++ {
			ok = ok && unicode.IsLetter(rune(b[i]))
		}
		return
	}

	isQuotedStr := func(b []byte) bool {
		return (len(b) > 2) &&
			((b[0] == '"' && b[len(b)-1] == '"') || (b[0] == '\'' && b[len(b)-1] == '\''))
	}

	// this is just a simple heuristic for trying to distinguish
	// a currency symbol from its code:
	//   - currency codes are: 3 upper case letters,
	//   - currency symbols are generally not letters, except some are! but they
	//     are lower/mixed case, mixed with punctuation, or are less than 3 chars,
	//   - stock tickers (codes) are: all upper case letters, but could be just
	//     one letter,
	//   - any non-currency object should be parsed as a code (ie: hours, shoes, cards)
	//     but is must not contain any spaces or be quoted
	isCode := func(b []byte) bool {
		return isAllLetter(b) && (isAllUpper(b) || len(b) >= 3)
	}

	// equivalent to bytes.SplitN where N=1 and the separator is
	// any amount of whitespace
	split1 := func(tok []byte) (head, tail []byte) {
		for i := 0; i < len(tok); {
			r, w := utf8.DecodeRune(tok[i:])
			if unicode.IsSpace(r) {
				head = tok[:i]
				tail = bytes.TrimSpace(tok[i+w:])
				return
			}
			i += w
		}
		return tok, tail
	}

	if len(tok) == 0 {
		return
	}

	r := bytes.IndexByte(tok, '@')
	if r != -1 {
		tok, tail = s.advance(tok, r)
	}

	if isQuotedStr(tok) {
		code = string(tok)

	} else {
		// split symbol and code
		left, right := split1(tok)
		if len(right) == 0 {
			// symbol OR code
			if isCode(left) {
				code = string(left)
			} else {
				sym = string(left)
			}
		} else {
			// symbol AND code
			right, tail = split1(right)
			sym = string(left)
			code = string(right)
		}
	}
	return
}

func (s *Scanner) ParseCommodity(tok []byte) (amount decimal.Decimal, com Commodity, tail []byte, err error) {
	if len(tok) == 0 {
		return
	}
	com.Type.Prefix, tail, err = s.ParsePostPrefix(tok)
	if err != nil {
		return
	}
	amount, com.Type.Decimal, tail, err = s.ParseDecimal(tail)
	if err != nil {
		return
	}
	com.Type.Postfix, com.Type.Code, tail = s.ParsePostfix(tail)
	return
}

func (s *Scanner) ParsePrice(tok []byte) (value decimal.Decimal, comType CommodityType, perUnit bool, tail []byte, err error) {
	if len(tok) == 0 {
		return
	}
	if tok[0] != '@' {
		err = s.wrap(fmt.Errorf("unknown format, cant parse: '%s'", tok))
		return
	}
	_, tok = s.advance(tok, 1)

	if len(tok) == 0 {
		err = s.wrap(fmt.Errorf("missing unit value"))
		return
	}
	// @@ means total value
	if tok[0] == '@' {
		_, tok = s.advance(tok, 1)
		perUnit = false
	} else {
		perUnit = true
	}
	comType.Prefix, tail, err = s.ParsePostPrefix(tok)
	if err != nil {
		return
	}
	value, comType.Decimal, tail, err = s.ParseDecimal(tail)
	if err != nil {
		return
	}
	comType.Postfix, comType.Code, tail = s.ParsePostfix(tail)
	return
}
