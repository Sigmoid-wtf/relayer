package main

import (
	"chain/x/sigmoid/types"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

type Balance struct {
	Free   uint64 `json:"free"`
	Staked uint64 `json:"staked"`
}

func GetBalance() Balance {
	var balance Balance

	err := json.Unmarshal(RunPython3Command([]string{"btcli/balance.py", "balance"}), &balance)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Get balance: ", balance)

	return balance
}

func ProcessBalance(client *cosmosclient.Client, balance Balance) {
	ctx := context.Background()

	account, err := client.Account("bob")
	if err != nil {
		panic(err.Error())
	}

	addr, err := account.Address("cosmos")
	if err != nil {
		panic(err.Error())
	}

	msg := &types.MsgSetRaoCurrentStakedBalance{
		Creator:                 addr,
		RaoCurrentStakedBalance: balance.Staked,
	}

	txResp, err := client.BroadcastTx(ctx, account, msg)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(txResp)
}
