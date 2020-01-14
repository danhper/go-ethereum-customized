from . import transactions_fetcher


def sync_transactions(_args):
    transactions_fetcher.replicate_transactions()
