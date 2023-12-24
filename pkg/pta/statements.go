package pta

import (
	"strings"
)

type BalanceStatement struct {
	assets      map[string][]Lot
	liabilities map[string][]Lot
}

type IncomeStatement struct {
	revenue   map[string][]Lot
	expenses  map[string][]Lot
	netIncome []Lot
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
	// group the transaction lots by account
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
		revenue:  make(map[string][]Lot),
		expenses: make(map[string][]Lot),
	}
	// group the transaction lots by account
	for _, t := range transactions {
		for _, p := range t.Postings {
			if p.Commodity.Type == CURRENCY {
				if strings.Contains(p.Account, "income") || strings.Contains(p.Account, "revenue") {
					statement.revenue[p.Account] = append(statement.revenue[p.Account], p.Lot)
				} else if strings.Contains(p.Account, "expense") {
					statement.expenses[p.Account] = append(statement.expenses[p.Account], p.Lot)
				}
			}
		}
	}
	// aggregate lots to single value per asset type per account
	for acct, lots := range statement.revenue {
		lots = aggregateLotsPerCode(lots)
		// by convention, revenue is negative, but we need it to
		// as positive in the statement
		for i := range lots {
			lots[i].Amount = lots[i].Amount.Neg()
		}
		statement.revenue[acct] = aggregateLotsPerCode(lots)
	}
	for acct, lots := range statement.expenses {
		statement.expenses[acct] = aggregateLotsPerCode(lots)
	}
	// group revenue and expenses by commodity
	netIncomeByCode := make(map[string]Lot)
	for _, lots := range statement.revenue {
		for _, lot := range lots {
			netIncomeByCode[lot.Commodity.Code] = lot
		}
	}
	for _, lots := range statement.expenses {
		for _, explot := range lots {
			lot := netIncomeByCode[explot.Commodity.Code]
			lot.Amount = lot.Amount.Sub(explot.Amount)
			netIncomeByCode[lot.Commodity.Code] = lot
		}
	}
	// convert map to slice
	for _, lot := range netIncomeByCode {
		statement.netIncome = append(statement.netIncome, lot)
	}
	return statement
}
