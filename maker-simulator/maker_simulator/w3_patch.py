from typing import Callable
from web3.types import RPCEndpoint
from web3.geth import GethMiner
from web3.method import (
    Method,
    default_root_munger,
)


miner_setMaxBlockNumber = RPCEndpoint("miner_setMaxBlockNumber")


setMaxBlockNumber: Method[Callable[[int], bool]] = Method(
    miner_setMaxBlockNumber,
    mungers=[default_root_munger],
)

GethMiner.setMaxBlockNumber = setMaxBlockNumber
