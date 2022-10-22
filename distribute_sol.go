package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

type Sender struct {
	Wallet int
}

type Receivers struct {
	Wallets []int
}

func distributeSOL(client *rpc.Client, pvtKeys []string, retries uint, wgrp *sync.WaitGroup) {
	// var retries uint = 3
	var amountToSend uint64
	var amountToSendUserInput float64

	fmt.Print("Amount each wallet should have: ")
	fmt.Scanln(&amountToSendUserInput)

	amountToSend = uint64(amountToSendUserInput * 1000000000)
	if amountToSend <= 0 {
		fmt.Println("Amount should be >= 0. Amount received:", amountToSend)
		main()
	}

	var pubKeys []string
	for _, key := range pvtKeys {
		pvtKey, err := solana.PrivateKeyFromBase58(key)
		if err != nil {
			panic(err)
		}
		pubKeys = append(pubKeys, pvtKey.PublicKey().String())
	}

	var chooseSender = []*survey.Question{
		{
			Name: "wallet",
			Prompt: &survey.Select{
				Message: "Choose the sending wallet:",
				Options: pubKeys,
			},
		},
	}

	var chosenSender Sender
	err := survey.Ask(chooseSender, &chosenSender)
	if err != nil {
		fmt.Println(err.Error())
	}

	var sendingWalletPub solana.PublicKey = solana.MustPublicKeyFromBase58(pubKeys[chosenSender.Wallet])
	var sendingWalletPvt solana.PrivateKey = solana.MustPrivateKeyFromBase58(pvtKeys[chosenSender.Wallet])

	bal, err := client.GetBalance(
		context.TODO(),
		sendingWalletPub,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	if bal.Value == 0 {
		fmt.Printf("Sender %s has %d SOL.\n\n", sendingWalletPub.String(), bal.Value)
		main()
	}

	copy(pubKeys[chosenSender.Wallet:], pubKeys[chosenSender.Wallet+1:])
	pubKeys[len(pubKeys)-1] = ""
	pubKeys = pubKeys[:len(pubKeys)-1]

	// copy(pvtKeys[chosenSender.Wallet:], pvtKeys[chosenSender.Wallet+1:])
	// pvtKeys[len(pvtKeys)-1] = ""
	// pvtKeys = pvtKeys[:len(pvtKeys)-1]

	var chooseReceivers = []*survey.Question{
		{
			Name: "wallets",
			Prompt: &survey.MultiSelect{
				Message: "Choose receiving wallets:",
				Options: pubKeys,
			},
		},
	}

	var chosenReceivers Receivers
	err = survey.Ask(chooseReceivers, &chosenReceivers)
	if err != nil {
		fmt.Println(err.Error())
	}

	var tx *solana.TransactionBuilder = solana.NewTransactionBuilder()

	var count uint8 = 0

	for i := 0; i < len(chosenReceivers.Wallets); i++ {
		count++

		var receivingWallet solana.PublicKey = solana.MustPublicKeyFromBase58(pubKeys[chosenReceivers.Wallets[i]])

		bal, err := client.GetBalance(
			context.TODO(),
			receivingWallet,
			rpc.CommitmentFinalized,
		)
		if err != nil {
			panic(err)
		}

		var finLamports uint64 = amountToSend - bal.Value
		if finLamports == 0 {
			continue
		}

		recentBlockhash, err := client.GetRecentBlockhash(
			context.TODO(),
			rpc.CommitmentFinalized,
		)
		if err != nil {
			panic(err)
		}

		var txInstruc *system.Transfer = system.NewTransferInstruction(
			finLamports,
			sendingWalletPub,
			receivingWallet,
		)

		tx.SetRecentBlockHash(recentBlockhash.Value.Blockhash)
		tx.AddInstruction(txInstruc.Build())

		if count == 20 {
			txToSend, err := tx.Build()
			if err != nil {
				panic(err)
			}

			_, err = txToSend.Sign(
				func(pubkey solana.PublicKey) *solana.PrivateKey {
					if sendingWalletPvt.PublicKey().Equals(pubkey) {
						return &sendingWalletPvt
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

			fmt.Printf("Filled %d wallets up to %f - %s\n", count, amountToSendUserInput, sig.String())

			count = 0
			tx = solana.NewTransactionBuilder()
		}
	}

	if count > 0 {
		txToSend, err := tx.Build()
		if err != nil {
			panic(err)
		}

		_, err = txToSend.Sign(
			func(pubkey solana.PublicKey) *solana.PrivateKey {
				if sendingWalletPvt.PublicKey().Equals(pubkey) {
					return &sendingWalletPvt
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

		fmt.Printf("Filled %d wallets up to %f - %s\n", count, amountToSendUserInput, sig.String())
	}
	fmt.Println()
	wgrp.Done()
}
