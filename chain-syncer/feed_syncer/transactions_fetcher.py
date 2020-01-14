import logging
import time

import web3

from . import settings
from . import w3_patch # this is used for side effets pylint: disable=unused-import


def replicate_transactions():
    mainnet_client = web3.Web3(web3.HTTPProvider(settings.MAINNET_URL))
    fork_client = web3.Web3(web3.HTTPProvider(settings.FORK_URL))

    main_block_number = mainnet_client.eth.blockNumber
    fork_block_number = fork_client.eth.blockNumber

    def replicate_transaction(tx_hash):
        raw_transaction = mainnet_client.eth.getRawTransaction(tx_hash)
        try:
            fork_client.eth.sendRawTransaction(raw_transaction)
        except ValueError as ex:
            logging.error("failed to replicate %s: %s", tx_hash.hex(), ex)

    logging.info("catching up from block %s to block %s", fork_block_number, main_block_number)
    for block_number in range(fork_block_number, main_block_number + 1):
        block = mainnet_client.eth.getBlock(block_number)
        for tx_hash in block["transactions"]:
            replicate_transaction(tx_hash)
        fork_client.geth.miner.setMaxBlockNumber(block_number)
        fork_client.geth.miner.start(1)
        while fork_client.eth.blockNumber < block_number:
            logging.info("waiting for blocked to be mined...")
            time.sleep(1)

        logging.info("progress: %s/%s", block_number, main_block_number)
