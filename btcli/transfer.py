import bittensor as bt
import click
import json

from common import *


@click.command()
def transfer_list():
    wallet = bt.wallet(name=WALLET_NAME)
    print(json.dumps(bt.commands.wallets.get_wallet_transfers(wallet.coldkeypub.ss58_address), indent=4))


cli.add_command(transfer_list)


if __name__ == '__main__':
    bt.turn_console_on()
    cli()
