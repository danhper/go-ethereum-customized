import argparse
import pickle
import logging
import gzip
import json
from functools import lru_cache
import datetime as dt
from collections import defaultdict
import copy
from os import path

import matplotlib.pyplot as plt
import seaborn as sns

import web3



# contract: https://etherscan.io/address/0x9ef05f7f6deb616fd37ac3c959a2ddd25a54e4f5

STATE_CACHE = "./data/states.pickle"


def get_func_signature_hash(signature):
    return web3.Web3.keccak(text=signature).hex()[2:2 + FUNC_LEN]


FUNC_LEN = 8
FUNC_HASHES = dict(
    vote=get_func_signature_hash("vote(address[])"),
    vote_slate=get_func_signature_hash("vote(bytes32)"),
    etch=get_func_signature_hash("etch(address[])"),
    lift=get_func_signature_hash("lift(address)"),
    lock=get_func_signature_hash("lock(uint256)"),
    free=get_func_signature_hash("free(uint256)"),
)

TRACK_VOTE_THRESHOLD = 50_000

STR_WORD_SIZE = 64
STR_ADDRESS_SIZE = 40

DATA_DIR = path.realpath("data")


def load_data(data_dir=DATA_DIR):
    with gzip.open(path.join(data_dir, "mkr-governance-txs.json.gz")) as f:
        return json.load(f)


def is_call(tx, func_name):
    prefix = "0x" + FUNC_HASHES[func_name]
    return tx["input"].startswith(prefix)


def get_vote_addresses(tx):
    args = tx["input"][FUNC_LEN + 2:]
    args_count = int(args[STR_WORD_SIZE:STR_WORD_SIZE * 2], 16)
    if not args_count:
        return ()
    args_start = STR_WORD_SIZE * 2
    args_end = args_start + STR_WORD_SIZE * args_count
    addr_start = STR_WORD_SIZE - STR_ADDRESS_SIZE
    addresses = [args[i:i + STR_WORD_SIZE][addr_start:]
                 for i in range(args_start, args_end, STR_WORD_SIZE)]
    return tuple(addresses)


def get_lift_address(tx):
    return tx["input"][FUNC_LEN + 2:][STR_WORD_SIZE - STR_ADDRESS_SIZE:]


def get_slate(tx):
    return "0x" + tx["input"][FUNC_LEN + 2:]


def get_locked_amount(tx):
    return int(tx["input"][FUNC_LEN + 2:], 16) / 1e18


def get_time(tx):
    return dt.datetime.fromtimestamp(int(tx["timeStamp"]))


def group_by_func(txs):
    grouped = defaultdict(list)
    for tx in txs:
        for func_name in FUNC_HASHES:
            if is_call(tx, func_name):
                grouped[func_name].append(tx)
    return grouped


def compute_locked_amount_evolution(transactions):
    current_locked = 0
    locked_amounts = []
    timestamps = []
    for tx in transactions:
        is_lock = is_call(tx, "lock")
        if is_lock or is_call(tx, "free"):
            sign = 1 if is_lock else -1
            locked_amount = get_locked_amount(tx)
            if locked_amount >= 30_000:
                print(tx["hash"], locked_amount, sign, get_time(tx).isoformat())
            current_locked += locked_amount * sign
            locked_amounts.append(current_locked)
            timestamps.append(get_time(tx))
    return locked_amounts, timestamps


@lru_cache(maxsize=None)
def hash_addresses(addresses):
    return web3.Web3.soliditySha3(
        ["address[]"],
        [[web3.Web3.toChecksumAddress(a) for a in addresses]]
    ).hex()


class State:
    def __init__(self):
        self.slates = defaultdict(list)
        self.addresses_slates = defaultdict(list)
        self.votes = defaultdict(int)
        self.deposits = defaultdict(int)
        self.approvals = defaultdict(int)
        self.fetched_approvals = {}
        self.timestamp = None
        self.block_number = None
        self.hat = None

    def add_weight(self, weight, slate):
        yays = self.slates[slate]
        for yay in yays:
            self.approvals[yay] += weight

    def sub_weight(self, weight, slate):
        yays = self.slates[slate]
        for yay in yays:
            self.approvals[yay] -= weight

    def lock(self, tx):
        sender = tx["from"]
        wad = get_locked_amount(tx)
        self.deposits[sender] += wad
        self.add_weight(wad, self.votes[sender])

    def free(self, tx):
        sender = tx["from"]
        wad = get_locked_amount(tx)
        self.deposits[sender] -= wad
        self.sub_weight(wad, self.votes[sender])

    def _etch(self, addresses):
        slate = hash_addresses(addresses)
        self.slates[slate] = addresses
        for address in addresses:
            self.addresses_slates[address].append(slate)
        return slate

    def etch(self, tx):
        addresses = get_vote_addresses(tx)
        slate = self._etch(addresses)
        return slate

    def vote(self, tx):
        addresses = get_vote_addresses(tx)
        slate = self._etch(addresses)
        self._vote_slate(tx, slate)
        return slate

    def vote_slate(self, tx):
        slate = get_slate(tx)
        self._vote_slate(tx, slate)

    def _vote_slate(self, tx, slate):
        if slate not in self.slates:
            logging.warning("slate not found tx=%s slate=%s", tx["hash"], slate)
        sender = tx["from"]
        weight = self.deposits[sender]
        self.sub_weight(weight, self.votes[sender])
        self.votes[sender] = slate
        self.add_weight(weight, slate)

    def lift(self, tx):
        lift_address = get_lift_address(tx)
        self.hat = lift_address

    def process_tx(self, tx):
        self.timestamp = get_time(tx)
        self.block_number = int(tx["blockNumber"])
        for func_name in FUNC_HASHES:
            if is_call(tx, func_name):
                getattr(self, func_name)(tx)
                break


def refetch_approvals(states):
    w3 = web3.Web3(web3.HTTPProvider("http://satoshi.doc.ic.ac.uk:8545"))
    with open("./contracts/mkr-governance.json") as f:
        abi = json.load(f)
    contract_address = w3.toChecksumAddress("0x9ef05f7f6deb616fd37ac3c959a2ddd25a54e4f5")
    contract = w3.eth.contract(address=contract_address, abi=abi)
    for i, state in enumerate(states):
        if i % 100 == 0:
            logging.info("progress: %s/%s", i, len(states))
        block = state.block_number
        for approval in state.approvals:
            address = w3.toChecksumAddress(approval)
            state.fetched_approvals[approval] = contract.functions.approvals(address).call(block_identifier=block) / 1e18


def compute_votes_evolution(transactions):
    state = State()
    states = []
    for tx in transactions:
        state = copy.deepcopy(state)
        state.process_tx(tx)
        states.append(state)
    refetch_approvals(states)
    with open(STATE_CACHE, "wb") as f:
        pickle.dump(states, f)
    return states


def _plot_locked(timestamps, locked_amounts, output=None):
    fig, ax = plt.subplots()
    ax.plot(timestamps, locked_amounts)
    ax.set_xlabel("Date")
    ax.set_ylabel("MKR amount")
    fig.autofmt_xdate(rotation=45)
    plt.tight_layout()

    if output is None:
        plt.show()
    else:
        plt.savefig(output)


def plot_votes_evolution(transactions, output=None):
    if path.exists(STATE_CACHE):
        with open(STATE_CACHE, "rb") as f:
            states = pickle.load(f)
    else:
        states = compute_votes_evolution(transactions)

    # sanity checking
    # locked_amounts = [sum(state.approvals.values()) for state in states]
    # _plot_locked(timestamps, locked_amounts)

    hats = set(state.hat for state in states if state.hat)

    addresses_to_track = list(
        set(address
            for state in states
            for address, amount in state.fetched_approvals.items()
            if amount >= TRACK_VOTE_THRESHOLD or address in hats))

    lines = list(zip(*[
        [state.fetched_approvals.get(address, 0) for address in addresses_to_track]
        for state in states
    ]))
    timestamps = [state.timestamp for state in states]

    fig, ax = plt.subplots()
    for address, line in zip(addresses_to_track, lines):
        ax.plot(timestamps, line, label="0x" + address[:10])
    ax.set_xlabel("Date")
    ax.set_ylabel("MKR amount")
    fig.autofmt_xdate(rotation=45)
    # ax.legend(ncol=3, bbox_to_anchor=(1.0, -0.5))
    plt.tight_layout()

    if output is None:
        plt.show()
    else:
        plt.savefig(output)


def plot_locked(transactions, output=None):
    locked_amounts, timestamps = compute_locked_amount_evolution(transactions)
    _plot_locked(timestamps, locked_amounts, output=output)


parser = argparse.ArgumentParser(prog="analyze-mkr-txs")
parser.add_argument("-d", "--data-dir", default=DATA_DIR, help="data directory")
subparsers = parser.add_subparsers(dest="command")

plot_locked_parser = subparsers.add_parser("plot-locked")
plot_locked_parser.add_argument("-o", "--output", help="output file")

plot_votes_parser = subparsers.add_parser("plot-votes")
plot_votes_parser.add_argument("-o", "--output", help="output file")



def main():
    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO)

    sns.set(style="darkgrid")

    transactions = sorted(load_data(args.data_dir)["result"], key=lambda x: (int(x["blockNumber"]), int(x["transactionIndex"])))
    successful_transactions = [tx for tx in transactions if tx["isError"] == "0"]

    if not args.command:
        parser.error("no command given")

    if args.command == "plot-locked":
        plot_locked(successful_transactions, args.output)
    elif args.command == "plot-votes":
        plot_votes_evolution(successful_transactions, args.output)


if __name__ == "__main__":
    main()
