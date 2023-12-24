package pta

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/shopspring/decimal"
)

func TestComputeIncomeStatement(t *testing.T) {
	testCases := []struct {
		name         string
		transactions []Transaction
		expected     IncomeStatement
	}{{
		name:         "empty",
		transactions: []Transaction{},
		expected: IncomeStatement{
			revenue:  make(map[string][]Lot),
			expenses: make(map[string][]Lot),
		},
	}, {
		name: "simple",
		transactions: []Transaction{
			{Postings: []Posting{{Account: "income:employer:salary", Lot: Lot{Amount: decimal.New(-1000, 0), Commodity: Commodity{Type: CURRENCY}}}}},
			{Postings: []Posting{{Account: "expenses:food:takeout", Lot: Lot{Amount: decimal.New(50, 0), Commodity: Commodity{Type: CURRENCY}}}}},
		},
		expected: IncomeStatement{
			revenue:   map[string][]Lot{"income:employer:salary": {{Amount: decimal.New(1000, 0), Commodity: Commodity{Type: CURRENCY}}}},
			expenses:  map[string][]Lot{"expenses:food:takeout": {{Amount: decimal.New(50, 0), Commodity: Commodity{Type: CURRENCY}}}},
			netIncome: []Lot{{Amount: decimal.New(950, 0), Commodity: Commodity{Type: CURRENCY}}},
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statement := ComputeIncomeStatement(tc.transactions)
			err := compareAccounts(statement.revenue, tc.expected.revenue)
			if err != nil {
				t.Errorf("revenue mismatch: %s", err)
			}
			err = compareAccounts(statement.expenses, tc.expected.expenses)
			if err != nil {
				t.Errorf("expenses mismatch: %s", err)
			}
			if len(statement.netIncome) != len(tc.expected.netIncome) {
				t.Errorf("net income length mismatch: got %d, want %d", len(statement.netIncome), len(tc.expected.netIncome))
			}
			for i, gotlot := range statement.netIncome {
				explot := tc.expected.netIncome[i]
				err := compareLots(gotlot, explot)
				if err != nil {
					t.Errorf("net income mismatch: %s", err)
				}
			}
		})
	}
}

func TestComputeBalanceStatement(t *testing.T) {

	startingBalance := BalanceStatement{
		assets: map[string][]Lot{
			"assets:savings": {
				{Amount: decimal.New(1000, 0)},
			},
			"assets:stocks": {
				{
					Amount:    decimal.New(100, 0),
					Commodity: Commodity{Code: "STOCK", Type: STOCK},
					UnitValue: Value{Decimal: decimal.New(10, 0)},
				},
			},
			"assets:house": {
				{
					Amount:    decimal.New(1, 0),
					Commodity: Commodity{Code: "123 street address", Type: OTHER},
					UnitValue: Value{Decimal: decimal.New(300_000, 0)},
				},
			},
		},
		liabilities: map[string][]Lot{
			"liabilities:mortgage": {
				{Amount: decimal.New(-200_000, 0)},
			},
		},
	}

	getPayed := Transaction{
		Postings: []Posting{
			{Account: "assets:checking", Lot: Lot{Amount: decimal.New(1000, 0)}},
			{Account: "income:emplyer", Lot: Lot{Amount: decimal.New(-1000, 0)}},
		},
	}

	payMortgage := Transaction{
		Postings: []Posting{
			{Account: "assets:checking", Lot: Lot{Amount: decimal.New(-900, 0)}},
			{Account: "liabilities:mortgage", Lot: Lot{Amount: decimal.New(900, 0)}},
		},
	}

	t.Run("starting balance", func(t *testing.T) {
		got := ComputeBalanceStatement(startingBalance, []Transaction{})
		expected := startingBalance
		err := compareAccounts(got.assets, expected.assets)
		if err != nil {
			t.Errorf("assets mismatch: %s", err)
			return
		}
		err = compareAccounts(got.liabilities, expected.liabilities)
		if err != nil {
			t.Errorf("liabilities mismatch: %s", err)
			return
		}
	})

	t.Run("a&l", func(t *testing.T) {
		got := ComputeBalanceStatement(startingBalance, []Transaction{getPayed, payMortgage})
		expected := startingBalance
		expected.assets["assets:checking"] = []Lot{{
			Amount: decimal.New(1000-900, 0),
		}}
		expected.liabilities["liabilities:mortgage"] = []Lot{{
			Amount: decimal.New(-200_000+900, 0),
		}}
		err := compareAccounts(got.assets, expected.assets)
		if err != nil {
			t.Errorf("assets mismatch: %s", err)
			return
		}
		err = compareAccounts(got.liabilities, expected.liabilities)
		if err != nil {
			t.Errorf("liabilities mismatch: %s", err)
			return
		}
	})
}

func TestAggregateLotsPerCode(t *testing.T) {
	testCases := []struct {
		name     string
		lots     []Lot
		expected []Lot
	}{{
		name:     "empty",
		lots:     []Lot{},
		expected: []Lot{},
	}, {
		name: "single",
		lots: []Lot{{
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}},
		expected: []Lot{{
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}},
	}, {
		name: "add currency",
		lots: []Lot{{
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}, {
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}},
		expected: []Lot{{
			Amount:    decimal.New(200, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}},
	}, {
		name: "add stock",
		lots: []Lot{{
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "STOCK", Type: STOCK},
			UnitValue: Value{Decimal: decimal.New(10, 0)},
		}, {
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "STOCK", Type: STOCK},
			UnitValue: Value{Decimal: decimal.New(20, 0)},
		}},
		expected: []Lot{{
			Amount:    decimal.New(200, 0),
			Commodity: Commodity{Code: "STOCK", Type: STOCK},
			UnitValue: Value{Decimal: decimal.New(15, 0)},
		}},
	}, {
		name: "mixed",
		lots: []Lot{{
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "STOCK", Type: STOCK},
			UnitValue: Value{Decimal: decimal.New(10, 0)},
		}, {
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}},
		expected: []Lot{{
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "STOCK", Type: STOCK},
			UnitValue: Value{Decimal: decimal.New(10, 0)},
		}, {
			Amount:    decimal.New(100, 0),
			Commodity: Commodity{Code: "USD", Type: CURRENCY},
		}},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := aggregateLotsPerCode(tc.lots)
			for i := range got {
				err := compareLots(got[i], tc.expected[i])
				if err != nil {
					t.Errorf("lot mismatch: %s", err)
					return
				}
			}
		})
	}
}

func compareAccounts(account, bccount map[string][]Lot) error {
	if len(account) != len(bccount) {
		return fmt.Errorf("account count mismatch")
	}
	for aname, alots := range account {
		blots, found := bccount[aname]
		if !found {
			return fmt.Errorf("account %v: not found", aname)
		}
		if len(alots) != len(blots) {
			return fmt.Errorf("account %v: lot count mismatch", aname)
		}
		for i, alot := range alots {
			blot := blots[i]
			err := compareLots(alot, blot)
			if err != nil {
				return fmt.Errorf("account %v: %v", aname, err)
			}
		}
	}
	return nil
}

func compareLots(got, exp Lot) error {
	if !got.Amount.Equal(exp.Amount) {
		return fmt.Errorf("Lot amount mismatch: got %v, want %v", got.Amount, exp.Amount)
	}
	if !reflect.DeepEqual(got.Commodity, exp.Commodity) {
		return fmt.Errorf("Lot commodity mismatch: got %v, want %v", got.Commodity, exp.Commodity)
	}
	if !got.UnitValue.Decimal.Equal(exp.UnitValue.Decimal) {
		return fmt.Errorf("Lot unit value mismatch: got %v, want %v", got.UnitValue, exp.UnitValue)
	}
	if !reflect.DeepEqual(got.UnitValue.Commodity, exp.UnitValue.Commodity) {
		return fmt.Errorf("Lot unit value commodity mismatch: got %v, want %v", got.UnitValue.Commodity, exp.UnitValue.Commodity)
	}
	return nil
}

func BenchmarkDecimalOps(b *testing.B) {
	// decimal: 311.7 ns/op	     120 B/op	       4 allocs/op
	{
		x := decimal.New(100, 0)
		y := decimal.New(50, 0)

		b.Run("decimal", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				x = x.Add(y)
				x = x.Sub(y)
			}
		})
	}
	// apd: 258.4 ns/op	       0 B/op	       0 allocs/op
	{
		c := apd.BaseContext.WithPrecision(1000)
		ed := apd.MakeErrDecimal(c)
		x := apd.New(100, 0)
		y := apd.New(50, 0)

		b.Run("apd", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				x = ed.Add(x, x, y)
				x = ed.Sub(x, x, y)
			}
		})
	}
}
