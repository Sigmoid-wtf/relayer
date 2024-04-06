package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ignite/cli/ignite/pkg/cosmosclient"

	"chain/x/sigmoidtest/types"
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

	fmt.Println(transfers)

	return transfers
}

func ProcessTransfers(client *cosmosclient.Client, transfers []Tranfer) {
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
	getLastProcessedResp, err := queryClient.GetLastProcessed(ctx, &types.QueryGetLastProcessedRequest{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(getLastProcessedResp)

	processed, err := strconv.ParseUint(getLastProcessedResp.TransactionId, 10, 64)
	if err != nil {
		panic(err.Error())
	}

	for processed++; processed < uint64(len(transfers)); processed++ {
		amount, err := strconv.ParseUint(transfers[processed].Amount, 10, 64)
		if err != nil {
			panic(err.Error())
		}

		getAmountResp, err := queryClient.GetAmount(ctx, &types.QueryGetAmountRequest{SenderAddress: transfers[processed].From})
		if err != nil {
			panic(err.Error())
		}

		if getAmountResp.Amount == amount {
			msg := &types.MsgApproveRequest{
				Creator:       addr,
				SenderAddress: transfers[processed].From,
				TransactionId: string(processed),
			}

			txResp, err := client.BroadcastTx(ctx, account, msg)
			if err != nil {
				panic(err.Error())
			}

			fmt.Println(txResp)
		} else {
			msg := &types.MsgProcessTransaction{
				Creator:       addr,
				TransactionId: string(processed),
			}

			txResp, err := client.BroadcastTx(ctx, account, msg)
			if err != nil {
				panic(err.Error())
			}

			fmt.Println(txResp)
		}
	}
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

	ProcessTransfers(client, GetTransfers())

	go func() {
		for {
			select {
			case <-ticker.C:
				ProcessTransfers(client, GetTransfers())
			}
		}
	}()
}
