package app

import (
	"fireside/pkg/pta"
	"fmt"
)

func Run() {

	fmt.Println("Fireside")
	_, txs, err := pta.ParseJournal("pkg/pta/test/bench1.journal")

	if err != nil {
		fmt.Println(err)
	}

	for _, tx := range txs {
		fmt.Println(pta.WriteTransaction(tx))
	}
}
