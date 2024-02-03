package app

import (
	"fireside/pkg/pta"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"time"
)

func AppendPlaintext(uid, selectedFile, txStr string) error {
	if selectedFile == "" {
		return fmt.Errorf("no journal file selected")
	}
	absFilepath := path.Clean(
		filepath.Join(root, uid, selectedFile),
	)
	journal, _, err := pta.ParseJournal(absFilepath)
	if err != nil {
		return err
	}
	txs, err := journal.ParseTransactionStrings(txStr)
	if err != nil {
		return err
	}
	return journal.AppendTxs(txs)
}

func RecentTransactions(uid, selectedFile string, since time.Time) ([]pta.Transaction, error) {
	if selectedFile == "" {
		return nil, fmt.Errorf("no journal file selected")
	}
	absFilepath := path.Clean(
		filepath.Join(root, uid, selectedFile),
	)
	_, txs, err := pta.ParseJournal(absFilepath)
	if err != nil {
		return nil, err
	}
	filtered := filterTxSince(txs, since)
	slices.Reverse[[]pta.Transaction](filtered)
	return filtered, nil
}

func filterTxSince(txs []pta.Transaction, filterDate time.Time) []pta.Transaction {
	var filtered []pta.Transaction
	for _, tx := range txs {
		if tx.Date.After(filterDate) {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

func TxStringify(txs []pta.Transaction) (ret []string) {
	for _, tx := range txs {
		ret = append(ret, pta.WriteTransaction(tx))
	}
	return
}
