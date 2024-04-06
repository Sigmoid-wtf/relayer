import click
from functools import wraps
import json


WALLET_NAME = 'sigma'


@click.group()
def cli():
    pass


def delegate_options():

    def decorator_wrapper(func):

        @click.option('--ss58-address', type=str, required=True, help='ss58_address')
        @click.option('--amount', type=float, required=True, help='TAO amount to delegate')
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return decorator_wrapper
