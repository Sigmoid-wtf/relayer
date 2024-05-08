package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"

	"chain/x/sigmoid/types"
)

type Balance struct {
	Free   uint64 `json:"free"`
	Staked uint64 `json:"staked"`
}

type Tranfer struct {
	Id          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	ExtrinsicId uint   `json:"extrinsicId"`
	BlockNumber string `json:"blockNumber"`
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

func GetTransfers() []Tranfer {
	var transfers []Tranfer

	err := json.Unmarshal(RunPython3Command([]string{"btcli/transfer.py", "transfer-list"}), &transfers)
	if err != nil {
		panic(err.Error())
	}
	slices.Reverse(transfers)

	fmt.Println("Transfer transactions count: ", len(transfers))

	return transfers
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

func ProcessTransfers(client *cosmosclient.Client, transfers []Tranfer) {
	ctx := context.Background()

	account, err := client.Account("alice")
	if err != nil {
		panic(err.Error())
	}

	addr, err := account.Address("cosmos")
	if err != nil {
		panic(err.Error())
	}

	queryClient := types.NewQueryClient(client.Context())
	getLastProcessedResp, err := queryClient.GetLastProcessed(ctx, &types.QueryGetLastProcessedRequest{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(getLastProcessedResp)

	processed, err := strconv.ParseInt(getLastProcessedResp.TransactionId, 10, 64)
	if err != nil {
		panic(err.Error())
	}

	for processed++; processed < int64(len(transfers)); processed++ {
		processTransaction := func() {
			msg := &types.MsgProcessTransaction{
				Creator:       addr,
				TransactionId: strconv.FormatInt(int64(processed), 10),
			}

			txResp, err := client.BroadcastTx(ctx, account, msg)
			if err != nil {
				panic(err.Error())
			}

			fmt.Println(txResp)
		}

		amount, err := strconv.ParseUint(transfers[processed].Amount, 10, 64)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Amount from bittensor tx", amount)

		getAmountResp, err := queryClient.GetAmount(ctx, &types.QueryGetAmountRequest{SenderAddress: transfers[processed].From})
		if err != nil {
			fmt.Println(err.Error())
			processTransaction()
			continue
		}
		fmt.Println("Amount from sigmoid tx", amount)

		if getAmountResp.Amount == amount {
			fmt.Println("Delegate ", amount, " RAO")
			delegate := RunPython3Command([]string{
				"btcli/delegate.py", "delegate",
				"--ss58-address", "5F4tQyWrhfGVcNhoqeiNsR6KjD4wMZ2kfhLj4oHYuyHbZAc3",
				"--amount", strconv.FormatUint(amount, 10),
			})
			fmt.Println(string(delegate))

			msg := &types.MsgApproveRequest{
				Creator:       addr,
				SenderAddress: transfers[processed].From,
				TransactionId: strconv.FormatInt(int64(processed), 10),
			}

			txResp, err := client.BroadcastTx(ctx, account, msg)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println(txResp)
		} else {
			processTransaction()
		}
	}
}

func ProcessUnstakeRequest(client *cosmosclient.Client) {
	ctx := context.Background()

	account, err := client.Account("bob")
	if err != nil {
		panic(err.Error())
	}

	addr, err := account.Address("cosmos")
	if err != nil {
		panic(err.Error())
	}

	queryClient := types.NewQueryClient(client.Context())
	getPendingUnstakeResponse, err := queryClient.GetPendingUnstakeRequest(ctx, &types.QueryGetPendingUnstakeRequestRequest{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(getPendingUnstakeResponse)

	if *getPendingUnstakeResponse.Request == (types.MsgCreateUnstakeRequest{}) {
		fmt.Println("No requests to unstake")
		return
	}

	getSigtaoRateDResponse, err := queryClient.GetSigtaoRateD(ctx, &types.QueryGetSigtaoRateDRequest{})
	if err != nil {
		panic(err.Error())
	}

	undelegateRao := getPendingUnstakeResponse.Request.Amount * getSigtaoRateDResponse.SigtaoRateD / 1000000000
	fmt.Println("Undelegate ", undelegateRao, " RAO")
	delegate := RunPython3Command([]string{
		"btcli/delegate.py", "undelegate",
		"--ss58-address", "5F4tQyWrhfGVcNhoqeiNsR6KjD4wMZ2kfhLj4oHYuyHbZAc3",
		"--amount", strconv.FormatUint(undelegateRao, 10),
	})
	fmt.Println(string(delegate))

	transfer := RunPython3Command([]string{
		"btcli/transfer.py", "transfer",
		"--ss58-address", getPendingUnstakeResponse.Request.UnstakeAddress,
		"--amount", strconv.FormatUint(undelegateRao, 10),
	})
	fmt.Println(string(transfer))

	msg := &types.MsgApproveUnstakeRequest{
		Creator:        addr,
		UnstakeAddress: getPendingUnstakeResponse.Request.UnstakeAddress,
	}

	txResp, err := client.BroadcastTx(ctx, account, msg)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(txResp)
}

func ProcessBridgeRequest(client *cosmosclient.Client) {
	ctx := context.Background()

	account, err := client.Account("bob")
	if err != nil {
		panic(err.Error())
	}

	addr, err := account.Address("cosmos")
	if err != nil {
		panic(err.Error())
	}

	queryClient := types.NewQueryClient(client.Context())
	getPendingBridgeResponse, err := queryClient.GetPendingBridgeRequest(ctx, &types.QueryGetPendingBridgeRequestRequest{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(getPendingBridgeResponse)

	if *getPendingBridgeResponse.Request == (types.MsgCreateBridgeRequest{}) {
		fmt.Println("No requests to bridge")
		return
	}

	fmt.Println("Bridge from ", getPendingBridgeResponse.Request.Creator, " to ", getPendingBridgeResponse.Request.Erc20Address, " ", getPendingBridgeResponse.Request.Amount, " RAO")
	delegate := RunPython3Command([]string{
		"btcli/bridge.py", "bridge",
		"--address", getPendingBridgeResponse.Request.Erc20Address,
		"--amount", strconv.FormatUint(getPendingBridgeResponse.Request.Amount, 10),
	})
	fmt.Println(string(delegate))

	msg := &types.MsgApproveBridgeRequest{
		Creator: addr,
		Address: getPendingBridgeResponse.Request.Creator,
	}

	txResp, err := client.BroadcastTx(ctx, account, msg)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(txResp)
}

func Serve() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	addressPrefix := "cosmos"

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix))
	if err != nil {
		panic(err.Error())
	}

	ProcessTransfers(&client, GetTransfers())
	ProcessUnstakeRequest(&client)
	ProcessBalance(&client, GetBalance())
	ProcessBridgeRequest(&client)

	for {
		select {
		case <-ticker.C:
			ProcessTransfers(&client, GetTransfers())
			ProcessBridgeRequest(&client)
			ProcessUnstakeRequest(&client)
			ProcessBalance(&client, GetBalance())
		}
	}
}
