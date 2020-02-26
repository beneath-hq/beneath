import beneath
import os
import pytz
from datetime import datetime
from decimal import Decimal
from hexbytes import HexBytes
from structlog import get_logger
from tenacity import before_sleep_log, retry, stop_after_attempt, wait_fixed, wait_random
from time import sleep
from web3 import Web3
from web3.exceptions import BlockNotFound

PROJECT = "beneath-ethereum"
STABLE_STREAM = "blocks-stable"
UNSTABLE_STREAM = "blocks-unstable"
WEB3_PROVIDER_URL = os.getenv("WEB3_PROVIDER_URL", default=None)

LATEST_COUNT = 25
STABLE_AFTER = 12
POLL_SECONDS = 1

log = get_logger()
client = beneath.Client()
stable = client.stream(project_name=PROJECT, stream_name=STABLE_STREAM)
unstable = client.stream(project_name=PROJECT, stream_name=UNSTABLE_STREAM)
w3 = Web3(Web3.HTTPProvider(WEB3_PROVIDER_URL))

def main():
  latest = get_latest()
  stable_idx = len(latest) - 1

  while True:
    # get next block
    latest_number = -1 if len(latest) == 0 else latest[-1]["number"]
    next_number = latest_number + 1
    next_block = get_block(next_number)
    if not next_block:
      sleep(POLL_SECONDS)
      continue
    
    # reprocess previous block if parent hash doesn't match
    if (len(latest) > 0) and (next_block["parent_hash"] != latest[-1]["hash"]):
      latest.pop()
      log.info("fork", discard_number=latest_number)
      if stable_idx >= len(latest):
        stable_idx = len(latest) - 1
        log.info("fork_before_stable", next_block=next_block, latest=latest)
      continue

    # move latest forward (and keep latest trimmed)
    latest.append(next_block)
    if len(latest) >= LATEST_COUNT:
      latest = latest[1:]
      stable_idx -= 1

    # write unstable
    unstable.write([next_block])
    log.info("write_unstable", number=next_block["number"], hash=next_block["hash"].hex())

    # write stable if necessary
    if (len(latest) - STABLE_AFTER) > stable_idx:
      stable_idx += 1
      stable_block = latest[stable_idx]
      stable.write([stable_block])
      log.info("write_stable", number=stable_block["number"], hash=stable_block["hash"].hex())


def get_latest():
  # get most recently written value
  latest = stable.latest(max_rows=LATEST_COUNT, to_dataframe=False, warn_max=False)
  if len(latest) == 0:
    return latest
  latest = sorted(latest, key=lambda block: block["number"])
  if not validate_latest(latest):
    raise Exception("Inconsistent latest blocks from stable")
  return latest


def validate_latest(blocks):
  for b1, b2 in zip(blocks, blocks[1:]):
    if b2["parent_hash"] != b1["hash"]:
      return False
  return True


def log_retry(retry_state):
  log.info("get_block_retry", attempt=retry_state.attempt_number, outcome=retry_state.outcome)


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
    "timestamp": datetime.utcfromtimestamp(block["timestamp"]).replace(tzinfo=pytz.utc),
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


def safe_to_utf8(val):
  try:
    return val.decode('utf-8')
  except:
    return None


if __name__ == "__main__":
  main()
