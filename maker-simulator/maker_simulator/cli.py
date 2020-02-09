import argparse
import logging

from . import commands
from . import settings


parser = argparse.ArgumentParser(prog="feed-syncer")
subparsers = parser.add_subparsers(dest="command")

_sync_transactions_parser = subparsers.add_parser("sync-transactions")


fetch_events_parser = subparsers.add_parser("fetch-events")
fetch_events_parser.add_argument(
    "-a", "--address", default=settings.MKR_CONTRACT_ADDRESS,
    help="address of the contract from which to fetch the events")
fetch_events_parser.add_argument(
    "-n", "--to-fetch", type=int, default=100_000,
    help="number of events to fetch")
fetch_events_parser.add_argument("-o", "--output", required=True,
                                 help="output file")


def run():
    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO)

    if not args.command:
        parser.error("no command given")

    func = getattr(commands, "run_" + args.command.replace("-", "_"))
    func(args)
