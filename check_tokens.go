package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func checkTokens(client *rpc.Client, pvtKeys []string, wgrp *sync.WaitGroup) {
	for idx, key := range pvtKeys {
		pvtKey, err := solana.PrivateKeyFromBase58(key)
		if err != nil {
			panic(err)
		}

		getTokenAccounts, err := client.GetTokenAccountsByOwner(
			context.TODO(),
			pvtKey.PublicKey(),
			&rpc.GetTokenAccountsConfig{
				ProgramId: solana.TokenProgramID.ToPointer(),
			},
			&rpc.GetTokenAccountsOpts{
				Encoding: solana.EncodingBase64Zstd,
			},
		)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%d \t %s \t %d tokens\n", idx, pvtKey.PublicKey().String(), len(getTokenAccounts.Value))
	}

	fmt.Println()
	wg.Done()
}
