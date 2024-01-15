package app

import (
	"fireside/pkg/pta"
	"fmt"
	"path"
	"path/filepath"
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
