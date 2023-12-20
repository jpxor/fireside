package pta

import (
	"fmt"
	"strings"
)

var DefaultCurrency = Commodity{
	Type: CURRENCY,
	Code: "USD",
}

var currencyFormats = map[string]CommodityFormat{
	"ARS": {"$", ".", ",", ""},
	"AUD": {"$", ".", ",", ""},
	"BRL": {"R$", ".", ",", ""},
	"CAD": {"$", ",", ".", ""},
	"CLP": {"$", ",", ".", ""},
	"CNY": {"¥", ",", ".", ""},
	"COP": {"$", ",", ".", ""},
	"CZK": {"", ".", ",", " Kč"},
	"DKK": {"kr.", ".", ",", ""},
	"EUR": {"€", ".", ",", ""},
	"HKD": {"HK$", ",", ".", ""},
	"HUF": {"", ".", ",", "Ft"},
	"INR": {"₹", ",", ".", ""},
	"ILS": {"₪", ".", ",", ""},
	"JPY": {"¥", ",", ".", ""},
	"KRW": {"₩", ",", ".", ""},
	"MYR": {"RM", ",", ".", ""},
	"MXN": {"$", ",", ".", ""},
	"MAD": {"", ",", ".", ".د.م."},
	"NZD": {"$", ",", ".", ""},
	"NOK": {"kr", ",", ".", ""},
	"PHP": {"₱", ",", ".", ""},
	"PLN": {"", ".", ",", "zł"},
	"RUB": {"", ".", ",", "p."},
	"SAR": {"", ",", ".", "﷼"},
	"SGD": {"$", ",", ".", ""},
	"ZAR": {"R", ",", ".", ""},
	"SEK": {"", ".", ",", " kr"},
	"CHF": {"fr.", ".", ",", ""},
	"TWD": {"元", ",", ".", ""},
	"THB": {"", ",", ".", "฿"},
	"TRY": {"", ",", ".", "₺"},
	"GBP": {"£", ",", ".", ""},
	"USD": {"$", ",", ".", ""},
	"VND": {"", ".", ",", "₫"},
}

func isCurrencyCode(code string) bool {
	_, ok := currencyFormats[code]
	return ok
}

func commodityFromCode(code string) (com Commodity) {
	com.Code = code
	if isCurrencyCode(code) {
		com.Type = CURRENCY
	} else {
		com.Type = STOCK
	}
	return
}

func findMatchingCurrency(f CommodityFormat) (Commodity, error) {
	if f.Prefix == "" && f.Postfix == "" {
		return DefaultCurrency, nil
	}
	defaultFmt := currencyFormats[DefaultCurrency.Code]
	if strings.TrimSpace(defaultFmt.Prefix) == f.Prefix &&
		strings.TrimSpace(defaultFmt.Postfix) == f.Postfix {
		return DefaultCurrency, nil
	}
	if f.Prefix == "$" {
		return DefaultCurrency, fmt.Errorf("ambiguous currency format %+v", f)
	}
	for code, fmt := range currencyFormats {
		if strings.TrimSpace(fmt.Prefix) == f.Prefix &&
			strings.TrimSpace(fmt.Postfix) == f.Postfix {
			return Commodity{
				Type: CURRENCY,
				Code: code,
			}, nil
		}
	}
	return DefaultCurrency, fmt.Errorf("no currency matching format %+v", f)
}
