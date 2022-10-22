package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

func collectSOL(client *rpc.Client, pvtKeys []string, receiver string, retries uint, wgrp *sync.WaitGroup) {
	var receivingWallet solana.PublicKey = solana.MustPublicKeyFromBase58(receiver)

	var userAnswer bool = false
	yesNoPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Is this receiving wallet correct? %s", receivingWallet.String()),
	}
	survey.AskOne(yesNoPrompt, &userAnswer)
	if !userAnswer {
		os.Exit(0)
	}

	var fees uint16 = 5000
	// var retries uint = 3

	for _, key := range pvtKeys {
		pvtKey, err := solana.PrivateKeyFromBase58(key)
		if err != nil {
			panic(err)
		}

		if pvtKey.PublicKey().String() == receiver {
			continue
		}

		bal, err := client.GetBalance(
			context.TODO(),
			pvtKey.PublicKey(),
			rpc.CommitmentFinalized,
		)
		if err != nil {
			panic(err)
		}

		if bal.Value == 0 {
			fmt.Printf("%s - 0 SOL\n", pvtKey.PublicKey().Short(4))
			continue
		}

		recentBlockhash, err := client.GetRecentBlockhash(
			context.TODO(),
			rpc.CommitmentFinalized,
		)
		if err != nil {
			panic(err)
		}

		var lamportsAmount uint64 = bal.Value - uint64(fees)

		tx, err := solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					lamportsAmount,
					pvtKey.PublicKey(),
					receivingWallet,
				).Build(),
			},
			recentBlockhash.Value.Blockhash,
			solana.TransactionPayer(pvtKey.PublicKey()),
		)
		if err != nil {
			panic(err)
		}

		_, err = tx.Sign(
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
			tx,
			rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentFinalized,
				MaxRetries:          &retries,
			},
		)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s - %s\n", pvtKey.PublicKey().Short(4), sig.String())
	}
	fmt.Println()
	wgrp.Done()
}
