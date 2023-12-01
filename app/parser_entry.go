package app

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
		Decimal:         defaultDecimal(),
		DefaultCurrency: defaultCurrency(),
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
		tx, match, err := s.ParseTransaction(line)
		if match {
			transactions = append(transactions, tx)
			continue

		} else if err != nil {
			errs.add(err)
			continue
		}

		// check for directives (rare case)
		txs, match, err := s.ParseDirective(&journal, line)
		if match {
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

func (s *Scanner) ParseTransaction(line []byte) (tx Transaction, ok bool, err error) {
	var tail []byte
	var date time.Time

	date, tail, err = s.ParseDate(line)
	if date == NotDate || err != nil {
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

		var neg bool
		neg, tail, err = s.ParsePostNeg(tail)
		if err != nil {
			return
		}

		post.Amount, post.Commodity, tail, err = s.ParseCommodity(tail)
		if err != nil {
			return
		}

		if neg {
			post.Amount = post.Amount.Neg()
		}

		if len(tail) > 0 {
			var perUnit bool
			post.Commodity.BookValue, post.Commodity.ValueType, perUnit,
				tail, err = s.ParsePrice(tail)

			if err != nil {
				return
			}

			if !perUnit && !(post.Commodity.BookValue == decimal.Zero) {
				post.Commodity.BookValue = post.Commodity.BookValue.Div(post.Amount)
			}
		}

		if len(tail) > 0 {
			s.wrap(fmt.Errorf("unexpected tokens after transaction posting: '%s'", tail))
		}

		tx.Postings = append(tx.Postings, post)
	}

	ok = true
	return
}

// the file format allows omitting a single posting amount.
// this amount needs to be inferred. Omitted amounts will
// appear as decimal.Zero
//
// TODO:
// need to deal with cases where commodity types may not
// match, ie. when purchasing stocks
func balanceTransaction(tx *Transaction) error {

	// 1. sum all amounts per commodity type
	// 2. identify posting with missing amount
	balances := make(map[Commodity]decimal.Decimal)
	var inferredPost *Posting = nil

	missingCount := 0
	for i := 0; i < len(tx.Postings); i++ {
		post := &tx.Postings[i]

		if post.Amount.Equal(decimal.Zero) {
			missingCount++
			inferredPost = post
		} else {
			balances[post.Commodity] = balances[post.Commodity].Add(post.Amount)
		}
	}

	if missingCount > 1 {
		// TODO: I need to add file details: filename + row (+col)?
		return fmt.Errorf("missing posting amount, cannot infer more than one")
	}

	// 3. check balances (should all equal zero)
	// 4. infer the one missing amount if needed
	for com, balance := range balances {
		if !balance.Equal(decimal.Zero) {
			if missingCount > 0 {
				inferredPost.Commodity = com
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

func (s *Scanner) ParseDirective(j *Journal, line []byte) (txs []Transaction, ok bool, err error) {
	return
}
