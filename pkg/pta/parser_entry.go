package pta

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Parse a Plain Text Accounting file (the journal)
//
// A procedural take on the combinator style parser:
// composes parser functions to extract data from
// a structured file. Optimized for speed and
// minimizing allocations
//
// A TTD approach was taken to ensure correctness,
// covering corner cases, and preventing regressions
//
// The plain text file IS the database, so parsing
// speed and correctness is critical
//

func ParseJournal(filepath string) (Journal, []Transaction, error) {

	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("failed to open '%s' (%s)\r\n", filepath, err)
		return Journal{}, []Transaction{}, err
	}
	defer file.Close()

	s := Scanner{
		filename: filepath,
		Scanner:  bufio.NewScanner(file),
	}

	errs := ParseErrors{}
	transactions := []Transaction{}

	journal := Journal{
		Filepath:        s.filename,
		Decimal:         DefaultNumberFormat.Decimal,
		DefaultCurrency: DefaultCurrency,
		Alias:           make(map[string]string),
		Includes:        make([]Journal, 0),
	}

	for s.Scan() {
		line := s.Bytes()

		// trim comments, skip empty lines
		line, empty, _ := tidy(line)
		if empty {
			continue
		}

		// check for transaction (common case)
		tx, err := s.ParseTransaction(line)
		if err == nil {
			transactions = append(transactions, tx)
			continue
		} else if err != ErrNoMatch {
			errs.add(err)
			continue
		}

		// check for directives (rare case)
		txs, err := s.ParseDirective(&journal, line)
		if err == nil {
			if len(txs) > 0 {
				transactions = append(transactions, txs...)
			}
			continue
		} else if err != nil {
			errs.add(err)
			continue
		}

		errs.add(s.wrap(fmt.Errorf("skiped line: '%s'", line)))
	}

	err = s.Err()
	if err != nil {
		errs.add(err)
	}

	for i := 0; i < len(transactions); i++ {
		err := balanceTransaction(&transactions[i])
		if err != nil {
			errs.add(err)
		}
	}

	return journal, transactions, errs.get()
}

func (s *Scanner) ParseTransaction(line []byte) (tx Transaction, err error) {
	var tail []byte
	var date time.Time

	date, tail, err = s.ParseDate(line)
	if err != nil {
		return
	}

	tx = Transaction{
		Date:     date,
		Postings: make([]Posting, 0, 2),
	}

	tx.Pending, tail, err = s.ParseTxPending(tail)
	if err != nil {
		return
	}

	tx.Code, tail, err = s.ParseTxCode(tail)
	if err != nil {
		return
	}

	tx.Description, tx.Tags, err = s.ParseTxDesc(tail)
	if err != nil {
		return
	}

	// tx postings are indented on the following lines
	for s.Scan() {
		var empty bool
		var hadComment bool

		// check for end of tx postings
		line, empty, hadComment = tidy(s.Bytes())
		if empty {
			if hadComment {
				continue
			}
			break
		}

		tail, err = s.ParseIndent(line)
		if err != nil {
			return
		}

		var post Posting
		post.Account, tail, err = s.ParseAcctName(tail)
		if err != nil {
			return
		}

		post.Lot, tail, err = s.ParseLot(tail)
		if err != nil {
			return
		}

		if len(tail) > 0 {
			s.wrap(fmt.Errorf("unexpected tokens after transaction posting: '%s'", tail))
		}

		tx.Postings = append(tx.Postings, post)
	}

	return
}

// the file format allows omitting a single posting amount.
// this amount needs to be inferred. Omitted amounts will
// appear as decimal.Zero
func balanceTransaction(tx *Transaction) error {

	// 1. sum all amounts per commodity type
	// 2. identify posting with missing amount
	balances := make(map[string]decimal.Decimal)
	var inferredPost *Posting = nil

	missingCount := 0
	for i := 0; i < len(tx.Postings); i++ {
		post := &tx.Postings[i]

		if post.Amount.Equal(decimal.Zero) {
			missingCount++
			inferredPost = post
		} else {
			balances[post.Commodity.Code] = balances[post.Commodity.Code].Add(post.Amount)
		}
	}

	if missingCount > 1 {
		// TODO: I need to add file details: filename + row (+col)?
		return fmt.Errorf("missing posting amount, cannot infer more than one")
	}

	// 3. check balances (should all equal zero)
	// 4. infer the one missing amount if needed
	for code, balance := range balances {
		if !balance.Equal(decimal.Zero) {
			if missingCount > 0 {
				inferredPost.Commodity = commodityFromCode(code)
				inferredPost.Amount = balance.Neg()
				missingCount--
			} else {
				// TODO: I need to add file details: filename + row (+col)?
				return fmt.Errorf("transaction is not balanced")
			}
		}
	}

	return nil
}

func ParsePath(cwd, incfile string) string {
	if filepath.IsAbs(incfile) {
		return incfile
	}
	return filepath.Join(cwd, incfile)
}

func (s *Scanner) ParseDirective(j *Journal, line []byte) (txs []Transaction, err error) {

	// include path should be relative to the current journal path
	if bytes.HasPrefix(line, []byte("include")) {
		cwd := filepath.Dir(j.Filepath)
		incfile := ParsePath(cwd, strings.TrimSpace(string(line[len("include"):])))
		subj, txs, err := ParseJournal(incfile)
		if err != nil {
			return nil, err
		}
		j.Includes = append(j.Includes, subj)
		return txs, nil
	}

	return nil, ErrNoMatch
}
