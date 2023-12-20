package pta

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	CURRENCY = CommodityType("currency|fungible")
	STOCK    = CommodityType("stock|fungible")
	OTHER    = CommodityType("nonfungible")
)

var DefaultNumberFormat = CommodityFormat{
	Thousandths: ",",
	Decimal:     ".",
}

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
	Account string
	Lot
}

type Lot struct {
	Date      time.Time
	Amount    decimal.Decimal
	UnitValue Value
	Commodity
}

type Value struct {
	decimal.Decimal
	Commodity
}

type Commodity struct {
	Type CommodityType
	Code string
}

type CommodityType string

type CommodityFormat struct {
	Prefix      string
	Thousandths string
	Decimal     string
	Postfix     string
}
