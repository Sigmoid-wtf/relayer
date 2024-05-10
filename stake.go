package main

import (
	"chain/x/sigmoid/types"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

type Tranfer struct {
	Id          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	ExtrinsicId uint   `json:"extrinsicId"`
	BlockNumber string `json:"blockNumber"`
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

func ProcessStakeRequest(client *cosmosclient.Client, transfers []Tranfer) {
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
