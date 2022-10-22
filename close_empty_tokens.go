package main

import (
	"context"
	"fmt"
	"sync"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func closeEmptyTokens(client *rpc.Client, pvtKeys []string, retries uint, wgrp *sync.WaitGroup) {
	// var retries uint = 0

	for _, key := range pvtKeys {
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

		for idx, rawAccount := range getTokenAccounts.Value {
			var tokenAccount token.Account
			var data []byte = rawAccount.Account.Data.GetBinary()
			var dec *bin.Decoder = bin.NewBinDecoder(data)
			err := dec.Decode(&tokenAccount)
			if err != nil {
				panic(err)
			}

			if tokenAccount.Amount == 0 {
				getMintAccount, err := client.GetTokenAccountsByOwner(
					context.TODO(),
					tokenAccount.Owner,
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

				var rawMintAccount []*rpc.TokenAccount = getMintAccount.Value
				var mintAccount solana.PublicKey = rawMintAccount[idx].Pubkey

				closeAccountParams := token.NewCloseAccountInstruction(
					mintAccount,
					tokenAccount.Owner,
					tokenAccount.Owner,
					[]solana.PublicKey{},
				)

				validateCloseAccountParams, err := closeAccountParams.ValidateAndBuild()
				if err != nil {
					panic(err)
				}

				var tx *solana.TransactionBuilder = solana.NewTransactionBuilder()

				recentBlockhash, err := client.GetRecentBlockhash(
					context.TODO(),
					rpc.CommitmentFinalized,
				)
				if err != nil {
					panic(err)
				}

				tx.AddInstruction(validateCloseAccountParams)
				tx.SetRecentBlockHash(recentBlockhash.Value.Blockhash)

				txToSend, err := tx.Build()
				if err != nil {
					panic(err)
				}

				_, err = txToSend.Sign(
					func(pubkey solana.PublicKey) *solana.PrivateKey {
						if pvtKey.PublicKey().Equals(pubkey) {
							return &pvtKey
						}
						return nil
					},
				)
				if err != nil {
					panic(fmt.Errorf("unable to sign tx %s", err.Error()))
				}

				sig, err := client.SendTransactionWithOpts(
					context.TODO(),
					txToSend,
					rpc.TransactionOpts{
						SkipPreflight:       true,
						PreflightCommitment: rpc.CommitmentFinalized,
						MaxRetries:          &retries,
					},
				)
				if err != nil {
					panic(err)
				}

				fmt.Printf("%s - %s\n", pvtKey.PublicKey(), sig.String())
			}
		}
		// fmt.Printf("%d \t %s \t %d tokens\n", idx, pvtKey.PublicKey().String(), len(getTokenAccounts.Value))
	}
	fmt.Println()
	wgrp.Done()
}
