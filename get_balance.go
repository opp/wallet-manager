package main

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func getBalance(client *rpc.Client, pvtKeys []string, wgrp *sync.WaitGroup) {
	for idx, key := range pvtKeys {
		pvtKey, err := solana.PrivateKeyFromBase58(key)
		if err != nil {
			panic(err)
		}

		bal, err := client.GetBalance(
			context.TODO(),
			pvtKey.PublicKey(),
			rpc.CommitmentFinalized,
		)
		if err != nil {
			panic(err)
		}

		var lamportsOnAccount = new(big.Float).SetUint64(uint64(bal.Value))
		var solAmount = new(big.Float).Quo(lamportsOnAccount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))
		fmt.Printf("%d \t %s \t %v SOL\n", idx, pvtKey.PublicKey().String(), solAmount)
	}
	fmt.Println()
	wgrp.Done()
}
