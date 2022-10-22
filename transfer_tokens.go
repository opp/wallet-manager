package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func transferTokens(client *rpc.Client, pvtKeys []string, receiver string, retries uint, wgrp *sync.WaitGroup) {
	var receivingWallet solana.PublicKey = solana.MustPublicKeyFromBase58(receiver)

	var userAnswer bool = false
	yesNoPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Is this receiving wallet correct? %s", receivingWallet.String()),
	}
	survey.AskOne(yesNoPrompt, &userAnswer)
	if !userAnswer {
		os.Exit(0)
	}

	// var retries uint = 3

	for _, key := range pvtKeys {
		pvtKey, err := solana.PrivateKeyFromBase58(key)
		if err != nil {
			panic(err)
		}

		if pvtKey.PublicKey() == receivingWallet {
			continue
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

			if tokenAccount.Amount > 0 {
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

				var tx *solana.TransactionBuilder = solana.NewTransactionBuilder()

				var rawMintAccount []*rpc.TokenAccount = getMintAccount.Value
				var mintAccount solana.PublicKey = rawMintAccount[idx].Pubkey

				var createAssociatedTokenAddress *associatedtokenaccount.Create = associatedtokenaccount.NewCreateInstruction(
					pvtKey.PublicKey(),
					receivingWallet,
					tokenAccount.Mint,
				)
				tx.AddInstruction(createAssociatedTokenAddress.Build())

				associatedTokenAccount, _, _ := solana.FindAssociatedTokenAddress(createAssociatedTokenAddress.Wallet, createAssociatedTokenAddress.Mint)

				var splTransfer *token.Transfer = token.NewTransferInstruction(
					tokenAccount.Amount,
					mintAccount,
					associatedTokenAccount,
					pvtKey.PublicKey(),
					[]solana.PublicKey{},
				)
				tx.AddInstruction(splTransfer.Build())

				var closeAccountParams *token.CloseAccount = token.NewCloseAccountInstruction(
					mintAccount,
					tokenAccount.Owner,
					tokenAccount.Owner,
					[]solana.PublicKey{},
				)

				validateCloseAccountParams, err := closeAccountParams.ValidateAndBuild()
				if err != nil {
					panic(err)
				}
				tx.AddInstruction(validateCloseAccountParams)

				recentBlockhash, err := client.GetRecentBlockhash(
					context.TODO(),
					rpc.CommitmentFinalized,
				)
				if err != nil {
					panic(err)
				}
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
	}

	fmt.Println()
	wgrp.Done()
}
