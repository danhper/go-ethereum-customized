import argparse
import logging

from . import commands


parser = argparse.ArgumentParser(prog="feed-syncer")
subparsers = parser.add_subparsers(dest="command")

_sync_transactions_parser = subparsers.add_parser("sync-transactions")


def run():
    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO)

    if not args.command:
        parser.error("no command given")

    func = getattr(commands, args.command.replace("-", "_"))
    func(args)
