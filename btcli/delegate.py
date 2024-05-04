import bittensor as bt
import click

from common import *


@click.command()
@delegate_options()
def delegate(ss58_address, amount):
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor()
    print(subtensor.delegate(
        wallet=wallet,
        delegate_ss58=ss58_address,
        amount=bt.Balance(amount),
    ))


@click.command()
@delegate_options()
def undelegate(ss58_address, amount):
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor()
    print(subtensor.undelegate(
        wallet=wallet,
        delegate_ss58=ss58_address,
        amount=bt.Balance(amount),
    ))


cli.add_command(delegate)
cli.add_command(undelegate)


if __name__ == '__main__':
    cli()
