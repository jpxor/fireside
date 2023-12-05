package pta

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/shopspring/decimal"
)

func (j *Journal) AppendTxs(txs []Transaction) error {
	f, err := os.OpenFile(j.Filepath, os.O_APPEND|os.O_WRONLY, fs.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, tx := range txs {
		_, err = f.WriteString(WriteTransaction(tx))
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteTransaction(tx Transaction) string {
	sb := strings.Builder{}
	sb.Grow(14 + len(tx.Code) + len(tx.Description))

	sb.WriteString(tx.Date.Format("2006/01/02"))
	if tx.Pending {
		sb.WriteString(" !")
	}
	if tx.Code != "" {
		sb.WriteString(" (")
		sb.WriteString(tx.Code)
		sb.WriteString(")")
	}
	if tx.Description != "" {
		if !tx.Pending && tx.Code == "" {
			sb.WriteString(" ")
		}
		sb.WriteString(" ")
		sb.WriteString(tx.Description)
	}

	acctWidth := 0
	for _, post := range tx.Postings {
		if acctWidth < len(post.Account)+2 {
			acctWidth = len(post.Account) + 2
		}
	}

	amountWidth := 0
	for _, post := range tx.Postings {
		nd := len(post.Amount.StringFixed(2))
		if amountWidth < nd {
			amountWidth = nd
		}
	}

	for _, post := range tx.Postings {
		sb.WriteString("\r\n\t")
		sb.WriteString(post.Account)
		sb.WriteString(strings.Repeat(" ", acctWidth-len(post.Account)))
		sb.WriteString(post.AmountStr(amountWidth))

		if !post.Commodity.BookValue.Equal(decimal.Zero) {
			sb.WriteString(" @ ")
			sb.WriteString(post.ValueStr())
		}
	}
	sb.WriteString("\r\n\r\n")
	return sb.String()
}

func (p Posting) ValueStr() string {
	sb := strings.Builder{}
	sb.WriteString(p.Commodity.ValueType.Prefix)
	sb.WriteString(p.Commodity.BookValue.String())
	sb.WriteString(p.Commodity.ValueType.Postfix)
	return sb.String()
}

func (p Posting) AmountStr(width int) string {
	return p.Commodity.Type.format(p.Amount, width)
}

func (c CommodityType) format(val decimal.Decimal, maxWidth int) string {
	var code string = ""
	var neg string = "  "
	var prefix string = " "

	if val.LessThan(decimal.Zero) {
		neg = "- "
		val = val.Abs()
	}
	if c.Code != "" {
		code = " " + c.Code
	}
	if c.Prefix != "" {
		prefix = c.Prefix
	}
	width := len(val.StringFixed(2))
	if maxWidth == 0 {
		maxWidth = width
	}
	str := fmt.Sprintf("%s%s%s%s%s%s", neg, prefix, strings.Repeat(" ", maxWidth-width), val.StringFixedBank(2), c.Postfix, code)
	if c.Decimal != "." && c.Decimal != "" {
		str = strings.Replace(str, ".", c.Decimal, 1)
	}

	return str
}
