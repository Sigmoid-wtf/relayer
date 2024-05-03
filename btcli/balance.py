import bittensor as bt
import click

from common import *


@click.command()
def balance():
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor(log_verbose=False)
    print(json.dumps({
        'free': subtensor.get_balance(wallet.coldkeypub.ss58_address).rao,
        'staked': subtensor.get_total_stake_for_coldkey(wallet.coldkeypub.ss58_address).rao,
    }))


cli.add_command(balance)


if __name__ == '__main__':
    cli()
