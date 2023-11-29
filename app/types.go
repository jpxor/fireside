package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Journal struct {
	Filepath        string
	Alias           map[string]string
	Decimal         string
	DefaultCurrency Commodity
	Includes        []Journal
	ParseErrs       ParseErrors
}

type Transaction struct {
	Date        time.Time
	Description string
	Code        string
	Tags        []string
	Postings    []Posting
	Pending     bool
}

type Posting struct {
	Account   string
	Amount    decimal.Decimal
	Commodity Commodity
}

type Commodity struct {
	Type      CommodityType
	ValueType CommodityType
	BookValue decimal.Decimal
}

type CommodityType struct {
	Prefix  string
	Decimal string
	Postfix string
	Code    string
}

func defaultDecimal() string {
	return "."
}

func defaultCurrency() Commodity {
	return Commodity{
		Type: CommodityType{
			Prefix:  "$",
			Decimal: defaultDecimal(),
		},
	}
}

func (c Commodity) format(val decimal.Decimal) string {
	var neg string
	var code string
	if val.LessThan(decimal.Zero) {
		neg = "- "
		val = val.Neg()
	}
	if c.Type.Code != "" {
		code = " " + c.Type.Code
	}
	str := fmt.Sprintf("%s%s%s%s%s", neg, c.Type.Prefix, val, c.Type.Postfix, code)
	if c.Type.Decimal != "." {
		str = strings.Replace(str, ".", c.Type.Decimal, 1)
	}
	return str
}

func (p Posting) AmountStr() string {
	return p.Commodity.format(p.Amount)
}
