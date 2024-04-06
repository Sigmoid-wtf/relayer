package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	return transfers
}

func ProcessTransfers(client *cosmosclient.client, addr string, transfers []Tranfer) {
	ctx := context.Background()

	queryClient := types.NewQueryClient(client.Context())

	queryResp, err := queryClient.GetLastProcessed(ctx, &types.QueryGetLastProcessedRequest{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(transfers)
	fmt.Println(queryResp)
}

func Serve() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	addressPrefix := "cosmos"

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix))
	if err != nil {
		log.Fatal(err)
	}

	account, err := client.Account("bob")
	if err != nil {
		log.Fatal(err)
	}

	addr, err := account.Address(addressPrefix)
	if err != nil {
		log.Fatal(err)
	}

	ProcessTransfers(client, addr, GetTransfers())

	go func() {
		for {
			select {
			case <-ticker.C:
				ProcessTransfers(client, addr, GetTransfers())
			}
		}
	}()
}
