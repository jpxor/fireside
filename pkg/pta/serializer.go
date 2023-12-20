package pta

import (
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
		if acctWidth < len(post.Account) {
			acctWidth = len(post.Account)
		}
	}

	amountWidth := 0
	for _, post := range tx.Postings {
		nd := len(post.AmountStr(0))
		if amountWidth < nd {
			amountWidth = nd
		}
	}

	for _, post := range tx.Postings {
		sb.WriteString("\r\n\t")
		sb.WriteString(post.Account)
		sb.WriteString(strings.Repeat(" ", 2+acctWidth-len(post.Account)))
		sb.WriteString(post.AmountStr(amountWidth))

		if !post.Lot.UnitValue.Decimal.Equal(decimal.Zero) {
			sb.WriteString(" @ ")
			sb.WriteString(post.ValueStr())
		}
	}
	sb.WriteString("\r\n\r\n")
	return sb.String()
}

func (p Posting) ValueStr() string {
	return commodityStringPadded(0, p.Lot.UnitValue.Commodity, p.Lot.UnitValue.Decimal)
}

func (p Posting) AmountStr(width int) string {
	return commodityStringPadded(width, p.Commodity, p.Amount)
}

func commodityStringPadded(width int, c Commodity, v decimal.Decimal) string {

	stringsReverse := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	var format CommodityFormat
	if c.Type == CURRENCY {
		format = currencyFormats[c.Code]
	} else {
		format = DefaultNumberFormat
	}

	var neg string = "  "
	if width == 0 {
		neg = ""
	}
	if v.LessThan(decimal.Zero) {
		neg = "- "
		v = v.Abs()
	}

	valstr := v.StringFixed(2)
	if format.Decimal != "." {
		valstr = strings.Replace(valstr, ".", format.Decimal, 1)
	}

	if len(valstr) > len("000.00") {
		revstr := stringsReverse(valstr[:len(valstr)-3])
		var sb strings.Builder
		for i, r := range revstr {
			if i > 0 && i%3 == 0 {
				sb.WriteString(format.Thousandths)
			}
			sb.WriteRune(r)
		}
		valstr = stringsReverse(sb.String())
	}

	var pad int
	if width > 0 {
		pad = width - (len(neg) + len(format.Prefix) + len(valstr) + len(format.Postfix))
		if pad < 0 {
			pad = 0
		}
	}

	sb := strings.Builder{}
	sb.WriteString(neg)
	sb.WriteString(format.Prefix)
	sb.WriteString(strings.Repeat(" ", pad))
	sb.WriteString(valstr)
	sb.WriteString(format.Postfix)

	if c.Code != DefaultCurrency.Code {
		sb.WriteString(" ")
		sb.WriteString(c.Code)
	}

	return sb.String()
}
