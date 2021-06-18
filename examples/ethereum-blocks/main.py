import asyncio
from datetime import datetime
from decimal import Decimal
import logging
import os

from hexbytes import HexBytes
import pytz
from tenacity import retry, stop_after_attempt, wait_fixed, wait_random
from web3 import Web3
from web3.exceptions import BlockNotFound

import beneath

LATEST_COUNT = 24
STABLE_AFTER = 12
POLL_SECONDS = 1

WEB3_PROVIDER_URL = os.getenv("WEB3_PROVIDER_URL", default=None)
w3 = Web3(Web3.HTTPProvider(WEB3_PROVIDER_URL))

SCHEMA = """
  type Block @schema {
    " Block number "
    number: Int! @key

    " Block timestamp "
    timestamp: Timestamp!

    " Block hash "
    hash: Bytes32!

    " Hash of parent block "
    parent_hash: Bytes32!

    " Address of block miner "
    miner: Bytes20!

    " Size of block in bytes "
    size: Int!

    " Number of transactions in block "
    transactions: Int!

    " Block difficulty "
    difficulty: Numeric!

    " Total difficulty of the chain until this block "
    total_difficulty: Numeric!

    " Limit on the amount of gas that can be consumed in a single block at the time of this block "
    gas_limit: Int!

    " Total amount of gas consumed by transactions in this block "
    gas_used: Int!

    " Extra data embedded in the block by its miner "
    extra_data: Bytes!

    " Extra data parsed as a UTF-8 encoded string, if possible "
    extra_data_text: String

    " Proof-of-work for this block "
    nonce: Bytes!

    " Root value of the receipts trie at this block "
    receipts_root: Bytes32!

    " Root value of the state trie at this block "
    state_root: Bytes32!

    " Root value of the transactions trie at this block "
    transactions_root: Bytes32!

    " Bloom filter of logs emitted in this block "
    logs_bloom: Bytes256!

    " SHA3 hash of the uncles in the block "
    sha3_uncles: Bytes32!  
  }
"""


retry_logger = logging.getLogger("tenacity")


def log_retry(retry_state):
    retry_logger.info(
        "get_block_retry attempt=%d outcome=%s",
        retry_state.attempt_number,
        retry_state.outcome,
    )


def safe_to_utf8(val):
    try:
        return val.decode("utf-8")
    except UnicodeDecodeError:
        return None


@retry(
    before_sleep=log_retry,
    reraise=True,
    stop=stop_after_attempt(5),
    wait=wait_fixed(2) + wait_random(0, 2),
)
def get_block(num):
    # pylint: disable=no-member
    try:
        block = w3.eth.getBlock(num)
    except BlockNotFound:
        return None

    return {
        "number": block["number"],
        "timestamp": datetime.utcfromtimestamp(block["timestamp"]).replace(
            tzinfo=pytz.utc
        ),
        "hash": bytes(block["hash"]),
        "parent_hash": bytes(block["parentHash"]),
        "miner": bytes(HexBytes(block["miner"])),
        "size": block["size"],
        "transactions": len(block["transactions"]),
        "difficulty": Decimal(block["difficulty"]),
        "total_difficulty": Decimal(block["totalDifficulty"]),
        "gas_limit": block["gasLimit"],
        "gas_used": block["gasUsed"],
        "extra_data": bytes(block["extraData"]),
        "extra_data_text": safe_to_utf8(bytes(block["extraData"])),
        "nonce": bytes(block["nonce"]),
        "receipts_root": bytes(block["receiptsRoot"]),
        "state_root": bytes(block["stateRoot"]),
        "transactions_root": bytes(block["transactionsRoot"]),
        "logs_bloom": bytes(block["logsBloom"]),
        "sha3_uncles": bytes(block["sha3Uncles"]),
    }


async def generate_blocks(p: beneath.Pipeline):
    checkpoint = await p.checkpoints.get(
        "checkpoint",
        default={
            "next_stable": 0,
            "next_unstable": 0,
            "latest_hashes": [],
        },
    )

    cached_blocks = []

    while True:
        # get next block
        unstable_block = get_block(checkpoint["next_unstable"])
        if not unstable_block:
            await asyncio.sleep(POLL_SECONDS)
            continue

        # reprocess previous block if parent hash doesn't match
        if (len(checkpoint["latest_hashes"]) > 0) and (
            unstable_block["parent_hash"] != checkpoint["latest_hashes"][-1]
        ):
            checkpoint["latest_hashes"].pop()
            p.logger.info("fork next_number=%d", checkpoint["next_unstable"])
            checkpoint["next_unstable"] -= 1
            if len(cached_blocks) > 0:
                cached_blocks.pop()
            if checkpoint["next_unstable"] < checkpoint["next_stable"]:
                checkpoint["next_stable"] = checkpoint["next_unstable"]
                p.logger.info(
                    "fork_before_stable next_number=%d", checkpoint["next_unstable"]
                )
            continue

        # process unstable_block
        yield ("unstable", unstable_block)
        p.logger.info(
            "write_unstable number=%d hash=%s",
            unstable_block["number"],
            unstable_block["hash"].hex(),
        )
        checkpoint["next_unstable"] += 1
        cached_blocks.append(unstable_block)

        # track in latest hashes (and keep it trimmed)
        checkpoint["latest_hashes"].append(unstable_block["hash"])
        if len(checkpoint["latest_hashes"]) >= LATEST_COUNT:
            checkpoint["latest_hashes"] = checkpoint["latest_hashes"][1:]

        # get and write stable if necessary
        while (checkpoint["next_stable"] + STABLE_AFTER) < checkpoint["next_unstable"]:
            # get stable block from cached_blocks or using get_block
            stable_block = None
            if (len(cached_blocks) > 0) and cached_blocks[0]["number"] == checkpoint[
                "next_stable"
            ]:
                stable_block = cached_blocks[0]
                cached_blocks = cached_blocks[1:]
                p.logger.info(
                    "get_stable_cache_hit number=%d", checkpoint["next_stable"]
                )
            else:
                p.logger.info(
                    "get_stable_cache_miss number=%d", checkpoint["next_stable"]
                )
                stable_block = get_block(checkpoint["next_stable"])
                if stable_block is None:
                    raise Exception(
                        "get_block should not return a null value for a stable block"
                    )

            # write stable block
            yield ("stable", stable_block)
            p.logger.info(
                "write_stable number=%d hash=%s",
                stable_block["number"],
                stable_block["hash"].hex(),
            )
            checkpoint["next_stable"] += 1

        # write updated checkpoint
        await p.checkpoints.set("checkpoint", checkpoint)


async def filter_stable(record):
    (key, block) = record
    if key == "stable":
        return block


async def filter_unstable(record):
    (key, block) = record
    if key == "unstable":
        return block


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True)
    blocks = p.generate(generate_blocks)
    stable = p.apply(blocks, filter_stable)
    unstable = p.apply(blocks, filter_unstable)
    p.write_table(
        unstable,
        "blocks-unstable",
        schema=SCHEMA,
        description=(
            "Blocks loaded in real-time from the Ethereum mainnet. "
            "Blocks are loaded without delay, so forks will frequently occur in this table."
        ),
    )
    p.write_table(
        stable,
        "blocks-stable",
        schema=SCHEMA,
        description=(
            "Blocks loaded in real-time from the Ethereum mainnet. "
            "Blocks are loaded with a 12-block delay to minimize the chance of forks."
        ),
    )
    p.main()
