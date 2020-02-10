import gzip
import json

import web3


FUNC_LEN = 8
FUNC_HASHES = dict(
    vote=web3.Web3.keccak(text="vote(address[])").hex()[2:2 + FUNC_LEN],
    lift=web3.Web3.keccak(text="lift(address)").hex()[2:2 + FUNC_LEN],
)

STR_WORD_SIZE = 64
STR_ADDRESS_SIZE = 40



with gzip.open("../data/mkr-governance-txs.json.gz") as f:
    data = json.load(f)


def is_call(tx, func_name):
    prefix = "0x" + FUNC_HASHES[func_name]
    return tx["input"].startswith(prefix)


def get_addresses(tx):
    args = tx["input"][len(VOTE_HASH) + 2:]
    args_count = int(args[STR_WORD_SIZE:STR_WORD_SIZE * 2], 16)
    if not args_count:
        return ()
    args_start = STR_WORD_SIZE * 2
    args_end = args_start + STR_WORD_SIZE * args_count
    addr_start = STR_WORD_SIZE - STR_ADDRESS_SIZE
    addresses = [args[i:i + STR_WORD_SIZE][addr_start:]
                 for i in range(args_start, args_end, STR_WORD_SIZE)]
    return tuple(addresses)



votes = [tx for tx in data["result"] if is_vote(tx)]

address_sets = set(get_addresses(tx) for tx in votes)

