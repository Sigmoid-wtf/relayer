package relayer

import (
	"encoding/binary"
	"fmt"
	"math/big"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func TransferWithGSRPC(destPublicKey string, rao_amount uint64, mnemonic string) {
	api, err := gsrpc.NewSubstrateAPI("wss://entrypoint-finney.opentensor.ai:443")
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}

	balance := new(big.Int).SetUint64(rao_amount)
	dest, err := types.NewMultiAddressFromHexAccountID(destPublicKey)
	if err != nil {
		panic(err)
	}

	call, err := types.NewCall(meta, "Balances.transfer", dest, types.NewUCompact(balance))
	if err != nil {
		panic(err)
	}

	keyPair, err := signature.KeyringPairFromSecret(mnemonic, 42)
	if err != nil {
		panic(err)
	}

	ext := types.NewExtrinsic(call)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		panic(err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		panic(err)
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", keyPair.PublicKey)
	if err != nil {
		panic(err)
	}

	storage, err := api.RPC.State.GetStorageRawLatest(key)
	if err != nil {
		panic(err)
	}

	nonce := binary.LittleEndian.Uint32((*storage)[0:4])
	signatureOptions := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(keyPair, signatureOptions)
	if err != nil {
		panic(err)
	}

	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		status := <-sub.Chan()
		fmt.Printf("Transaction status: %#v\n", status)

		if status.IsInBlock {
			fmt.Printf("Completed at block hash: %#x\n", status.AsInBlock)
			return
		}
	}
}
