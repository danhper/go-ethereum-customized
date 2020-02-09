from . import transactions_fetcher
from .etherscan_events_fetcher import fetch_events
from . import utils


def run_sync_transactions(_args):
    transactions_fetcher.replicate_transactions()


def run_fetch_events(args):
    with utils.open_file(args.output, "wt") as fout:
        fetch_events(args.address, fout, max_to_fetch=args.to_fetch)
