import click
import json

from web3 import Web3
from web3.exceptions import MismatchedABI
from web3.middleware import geth_poa_middleware
from web3._utils.events import get_event_data


CONTRACT_ADDRESS = '0x7f61C18F800b22ff2eEe21C779B674d57Cd99e6F'
CONTRACT_ABI = json.load(open('./btcli/abi/abi.json', 'r'))
PRIVATE_KEY = 'PRIVATE_KEY'
RPC_URL = 'https://polygon-mainnet.g.alchemy.com/v2/rE2mrPxHwMpTsGzejzyUv4X-Nd9FxkNK'


@click.group()
def cli():
    pass


@click.command()
@click.option('--address', type=str, required=True, help='erc20_address')
@click.option('--amount', type=int, required=True, help='RAO amount to delegate')
def bridge(address, amount):
    w3 = Web3(Web3.HTTPProvider(RPC_URL))
    w3.middleware_onion.inject(geth_poa_middleware, layer=0)
    token_contract = w3.eth.contract(address=CONTRACT_ADDRESS, abi=CONTRACT_ABI)

    try:
        account = w3.eth.account.from_key(PRIVATE_KEY)
        nonce = w3.eth.get_transaction_count(account.address)
        txn = token_contract.functions.mint(address, amount).build_transaction({
            'chainId': 137,  # polygon
            'nonce': nonce,
            'from': account.address
        })
        signed_txn = w3.eth.account.sign_transaction(txn, PRIVATE_KEY)
        tx_hash = w3.eth.send_raw_transaction(signed_txn.rawTransaction)
        w3.eth.wait_for_transaction_receipt(tx_hash)
        print(txn)
    except Exception as e:
        print(repr(e))


@click.command()
@click.option('--block', type=int, required=True, help='Block number')
def event_list(block):
    w3 = Web3(Web3.HTTPProvider(RPC_URL))
    w3.middleware_onion.inject(geth_poa_middleware, layer=0)

    latest_block = w3.eth.block_number
    token_contract = w3.eth.contract(address=CONTRACT_ADDRESS, abi=CONTRACT_ABI)
        
    events = w3.eth.get_logs({'fromBlock':block, 'toBlock': latest_block, 'address':CONTRACT_ADDRESS})

    event_template = token_contract.events.BridgeRequested
    result = {'events': [], 'latest': latest_block}
    for event in events:
        try:
            event_data = get_event_data(event_template.w3.codec, event_template._get_event_abi(), event)
            result['events'].append({
                'args': dict(event_data.args),
                'event': event_data.event,
                'block': event_data.blockNumber,
            })
        except MismatchedABI:
            pass
    print(json.dumps(result))


cli.add_command(bridge)
cli.add_command(event_list)


if __name__ == '__main__':
    cli()
