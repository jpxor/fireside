package pta

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCommodityStringPadded(t *testing.T) {
	amount := func(s string) decimal.Decimal {
		amount, _ := decimal.NewFromString(s)
		return amount
	}
	usd := Commodity{
		Code: "USD",
		Type: CURRENCY,
	}
	sek := Commodity{
		Code: "SEK",
		Type: CURRENCY,
	}
	aaa := Commodity{
		Code: "AAA",
		Type: STOCK,
	}

	cases := []struct {
		width    int
		comm     Commodity
		amount   decimal.Decimal
		expected string
	}{
		{width: 0, comm: usd, amount: amount("1234"), expected: "$1,234.00"},
		{width: 10, comm: usd, amount: amount("1234"), expected: "  $1,234.00"},
		{width: 0, comm: usd, amount: amount("-1234"), expected: "- $1,234.00"},
		{width: 14, comm: usd, amount: amount("-1234"), expected: "- $ 1,234.00"},
		{width: 0, comm: sek, amount: amount("1234"), expected: "1.234,00 kr SEK"},
		{width: 0, comm: aaa, amount: amount("1234"), expected: "1,234.00 AAA"},
	}

	for _, c := range cases {
		got := commodityStringPadded(c.width, c.comm, c.amount)
		if got != c.expected {
			t.Errorf("commodityStringPadded(%d, %v, %v) = %q, want %q",
				c.width, c.comm, c.amount, got, c.expected)
		}
	}
}
