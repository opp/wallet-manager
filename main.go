package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gagliardetto/solana-go/rpc"
)

type Config struct {
	RPC      string
	Receiver struct {
		SOL   string
		Token string
	}
	Retries uint
}

var wg sync.WaitGroup

var options = []*survey.Question{
	{
		Name: "option",
		Prompt: &survey.Select{
			Message: "Choose an option:",
			Options: []string{"Get Balance", "Collect SOL", "Distribute SOL", "Check Tokens", "Transfer Tokens", "Close Empty Token Accounts", "Show Keys", "Exit"},
		},
	},
}

type OptionChosen struct {
	Option string
}

func main() {
	var pvtKeys []string

	if len(os.Args) == 1 {
		pvtKeys = parseCSV("wallets.csv")
	} else {
		pvtKeys = parseCSV(os.Args[1])
	}

	configuration, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(configuration, &config)
	if err != nil {
		panic(err)
	}

	var chosenOption OptionChosen
	err = survey.Ask(options, &chosenOption)
	if err != nil {
		fmt.Println(err.Error())
	}

	wg.Add(1)

	if chosenOption.Option == "Get Balance" {
		go getBalance(rpc.New(config.RPC), pvtKeys, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Collect SOL" {
		go collectSOL(rpc.New(config.RPC), pvtKeys, config.Receiver.SOL, config.Retries, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Distribute SOL" {
		go distributeSOL(rpc.New(config.RPC), pvtKeys, config.Retries, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Check Tokens" {
		go checkTokens(rpc.New(config.RPC), pvtKeys, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Transfer Tokens" {
		go transferTokens(rpc.New(config.RPC), pvtKeys, config.Receiver.Token, config.Retries, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Close Empty Token Accounts" {
		go closeEmptyTokens(rpc.New(config.RPC), pvtKeys, config.Retries, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Show Keys" {
		go showKeys(pvtKeys, &wg)
		wg.Wait()
		main()
	} else if chosenOption.Option == "Exit" {
		os.Exit(0)
	}
}
