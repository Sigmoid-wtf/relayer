import click
import json

from time import sleep
from web3 import Web3
from web3.middleware import geth_poa_middleware


@click.group()
def cli():
    pass


@click.command()
@click.option('--address', type=str, required=True, help='erc20_address')
@click.option('--amount', type=int, required=True, help='RAO amount to delegate')
def bridge(address, amount):
    RPC_URL = "https://polygon-mainnet.g.alchemy.com/v2/rE2mrPxHwMpTsGzejzyUv4X-Nd9FxkNK"
    CONTRACT_ADDRESS = "0x7f61C18F800b22ff2eEe21C779B674d57Cd99e6F"
    CONTRACT_ABI = json.load(open("./btcli/abi/abi.json", "r"))
    print(CONTRACT_ABI)
    PRIVATE_KEY = "PRIVATE_KEY"

    w3 = Web3(Web3.HTTPProvider(RPC_URL))
    w3.middleware_onion.inject(geth_poa_middleware, layer=0)
    account = w3.eth.account.from_key(PRIVATE_KEY)
    token_contract = w3.eth.contract(address=CONTRACT_ADDRESS, abi=CONTRACT_ABI)
    for _ in range(3):
        try:
            nonce = w3.eth.get_transaction_count(account.address)
            txn = token_contract.functions.mint(address, amount).build_transaction({
                'chainId': 137,  # polygon
                'nonce': nonce,
                'from': account.address
            })
            print(txn)
            signed_txn = w3.eth.account.sign_transaction(txn, PRIVATE_KEY)
            print(signed_txn)
            tx_hash = w3.eth.send_raw_transaction(signed_txn.rawTransaction)
            print(tx_hash)
            w3.eth.wait_for_transaction_receipt(tx_hash)
            break
        except Exception as e:
            print(repr(e))
            sleep(2)


cli.add_command(bridge)


if __name__ == '__main__':
    cli()
