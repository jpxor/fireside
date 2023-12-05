package app

import (
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
