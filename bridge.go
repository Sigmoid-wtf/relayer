package main

import (
	"chain/x/sigmoid/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

type EventArgs struct {
	Erc20Address   string `json:"sender"`
	SigmoidAddress string `json:"sigmoidAddress"`
	Amount         uint64 `json:"amount"`
}

type Event struct {
	Args        EventArgs `json:"args"`
	EventName   string    `json:"event"`
	BlockNumber uint64    `json:"block"`
}

type EventList struct {
	Events            []Event `json:"events"`
	LatestBlockNumber uint64  `json:"latest"`
}

func ProcessBridgePolygonToSigmoidRequest(client *cosmosclient.Client) {
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
	getLatestProcessedEthBlockResponse, err := queryClient.GetLatestProcessedEthBlock(ctx, &types.QueryGetLatestProcessedEthBlockRequest{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(getLatestProcessedEthBlockResponse)

	var events EventList
	event_list := RunPython3Command([]string{
		"btcli/bridge.py", "event-list",
		"--block", getLatestProcessedEthBlockResponse.BlockNumber,
	})
	fmt.Println(string(event_list))

	err = json.Unmarshal(event_list, &events)
	if err != nil {
		panic(err.Error())
	}

	for _, event := range events.Events {
		msg := &types.MsgIncomeBridgeRequest{
			Creator: addr,
			Address: event.Args.SigmoidAddress,
			Amount:  event.Args.Amount,
		}

		txResp, err := client.BroadcastTx(ctx, account, msg)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(txResp)
	}

	msg := &types.MsgSetLatestProcessedEthBlock{
		Creator:     addr,
		BlockNumber: strconv.FormatUint(events.LatestBlockNumber, 10),
	}

	txResp, err := client.BroadcastTx(ctx, account, msg)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(txResp)
}

func ProcessBridgeSigmoidToPolygonRequest(client *cosmosclient.Client) {
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
