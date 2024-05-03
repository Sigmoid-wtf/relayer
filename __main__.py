import argparse 
import bittensor as bt
import click
import json


WALLET_NAME = 'sigma'


@click.group()
def cli():
    pass


@click.command()
@click.argument('ss58_address')
@click.option('--mnemonic', type=str)
def init_wallet(ss58_address, mnemonic):
    wallet = bt.wallet(name=WALLET_NAME)
    wallet.regenerate_coldkeypub(ss58_address=ss58_address, overwrite=True)
    wallet.regenerate_coldkey(mnemonic=mnemonic, use_password=False, overwrite=True)
    print(wallet)


@click.command()
@click.option('--dest', type=str, required=True, help='Destination ss58_address')
@click.option('--amount', type=float, required=True, help='TAO amount to delegate')
def delegate(dest, amount):
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor()
    subtensor.delegate(
        wallet=wallet,
        delegate_ss58=dest,
        amount=amount,
    )


@click.command()
@click.option('--src', type=str, required=True, help='Source ss58_address')
@click.option('--amount', type=float, required=True, help='TAO amount to undelegate')
def undelegate(src, amount):
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor()
    subtensor.undelegate(
        wallet=wallet,
        delegate_ss58=src,
        amount=amount,
    )


@click.command()
def transfer_list():
    wallet = bt.wallet(name=WALLET_NAME)
    print(json.dumps(bt.commands.wallets.get_wallet_transfers(wallet.coldkeypub.ss58_address), indent=4))


@click.command()
def my_delegates():
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor()
    all_delegates = subtensor.get_delegated(
        coldkey_ss58=wallet.coldkeypub.ss58_address,
    )

    delegates = {}
    for delegate in all_delegates:
        for coldkey_addr, staked in delegate[0].nominators:
            if coldkey_addr == wallet.coldkeypub.ss58_address and staked.tao > 0:
                delegates[delegate[0].hotkey_ss58] = str(staked)
    print(json.dumps(delegates, indent=4))


@click.command()
def balance():
    wallet = bt.wallet(name=WALLET_NAME)
    subtensor = bt.subtensor(log_verbose=False)
    print(json.dumps({
        'free': str(subtensor.get_balance(wallet.coldkeypub.ss58_address).tao),
        'staked': str(subtensor.get_total_stake_for_coldkey(wallet.coldkeypub.ss58_address).tao),
    }))


cli.add_command(init_wallet)
cli.add_command(delegate)
cli.add_command(undelegate)
cli.add_command(transfer_list)
cli.add_command(my_delegates)
cli.add_command(balance)


if __name__ == '__main__':
    bt.turn_console_on()
    cli()
