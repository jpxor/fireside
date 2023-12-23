package pta

import (
	"strings"

	"github.com/shopspring/decimal"
)

type BalanceStatement struct {
	assets      map[string][]Lot
	liabilities map[string][]Lot
}

type IncomeStatement struct {
	revenue   map[string]decimal.Decimal
	expenses  map[string]decimal.Decimal
	netIncome decimal.Decimal
}

func ComputeBalanceStatement(startingBalance BalanceStatement, transactions []Transaction) BalanceStatement {
	// deep copy the starting balance to create a starting point for
	// for the new balance statement
	statement := BalanceStatement{
		assets:      map[string][]Lot{},
		liabilities: map[string][]Lot{},
	}
	for acct, assetLots := range startingBalance.assets {
		statement.assets[acct] = append(statement.assets[acct], assetLots...)
	}
	for acct, liabLots := range startingBalance.liabilities {
		statement.liabilities[acct] = append(statement.liabilities[acct], liabLots...)
	}
	// add the transactions to the balance statement
	for _, t := range transactions {
		for _, p := range t.Postings {
			acct := p.Account
			if strings.Contains(acct, "asset") {
				statement.assets[acct] = append(statement.assets[acct], p.Lot)

			} else if strings.Contains(acct, "liability") || strings.Contains(acct, "liabilities") {
				statement.liabilities[acct] = append(statement.liabilities[acct], p.Lot)
			}
		}
	}
	// aggregate lots to single value per asset type per account
	for acct, lots := range statement.assets {
		statement.assets[acct] = aggregateLotsPerCode(lots)
	}
	for acct, lots := range statement.liabilities {
		statement.liabilities[acct] = aggregateLotsPerCode(lots)
	}
	return statement
}

func aggregateLotsPerCode(lots []Lot) []Lot {
	lotsByCode := make(map[string][]Lot)
	for _, lot := range lots {
		lotsByCode[lot.Commodity.Code] = append(lotsByCode[lot.Commodity.Code], lot)
	}
	aggregated := make([]Lot, 0, len(lotsByCode))
	for code, lots := range lotsByCode {
		reduced := Lot{
			Commodity: Commodity{
				Code: code,
				Type: lots[0].Commodity.Type,
			},
		}
		for _, lot := range lots {
			if reduced.Commodity.Type == CURRENCY {
				reduced.Amount = reduced.Amount.Add(lot.Amount)
			} else {
				totVal := reduced.Amount.Mul(reduced.UnitValue.Decimal)
				totVal = totVal.Add(lot.Amount.Mul(lot.UnitValue.Decimal))
				reduced.Amount = reduced.Amount.Add(lot.Amount)
				reduced.UnitValue.Decimal = totVal.Div(reduced.Amount)
			}
		}
		aggregated = append(aggregated, reduced)
	}
	return aggregated
}

func ComputeIncomeStatement(transactions []Transaction) IncomeStatement {
	statement := IncomeStatement{
		revenue:  make(map[string]decimal.Decimal),
		expenses: make(map[string]decimal.Decimal),
	}
	for _, t := range transactions {
		for _, p := range t.Postings {
			if p.Commodity.Type == CURRENCY {
				if strings.Contains(p.Account, "income") || strings.Contains(p.Account, "revenue") {
					// income & revenue are negative, as they are debits from external account
					// so we substract the amount to get positive value
					statement.revenue[p.Account] = statement.revenue[p.Account].Sub(p.Amount)
				} else if strings.Contains(p.Account, "expense") {
					statement.expenses[p.Account] = statement.expenses[p.Account].Add(p.Amount)
				}
			}
		}
	}
	for _, rev := range statement.revenue {
		statement.netIncome = statement.netIncome.Add(rev)
	}
	for _, exp := range statement.expenses {
		statement.netIncome = statement.netIncome.Sub(exp)
	}
	return statement
}
