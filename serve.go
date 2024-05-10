package main

import (
	"context"
	"time"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func ServeOnce(client *cosmosclient.Client) {
	ProcessBalance(client, GetBalance())

	ProcessStakeRequest(client, GetTransfers())
	ProcessUnstakeRequest(client)

	ProcessBridgePolygonToSigmoidRequest(client)
	ProcessBridgeSigmoidToPolygonRequest(client)
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

	ServeOnce(&client)

	for {
		select {
		case <-ticker.C:
			ServeOnce(&client)
		}
	}
}
