import logging
import time

import requests
from tenacity import before_sleep_log, retry, stop_after_attempt, wait_random
from web3 import Web3

import config

BENEATH_STREAM_URL = f"{config.BENEATH_BASE_URL}/projects/{config.BENEATH_PROJECT}/streams/{config.BENEATH_PROJECT_STREAM}"
BENEATH_GET_LATEST_BLOCK_URL = "http://not.working.yet/"  # f"{config.BENEATH_BASE_URL}/projects/{config.BENEATH_PROJECT}/streams/{config.BENEATH_PROJECT_STREAM}?get-latest-record"

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)

W3 = Web3(Web3.HTTPProvider(config.WEB3_PROVIDER_URL))


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR),
       reraise=True)
def get_stream_instance_id():
  response = requests.get(f"{BENEATH_STREAM_URL}/details",
                          headers={"Bearer": config.BENEATH_PROJECT_KEY})
  response.raise_for_status()
  return response.json()["current_instance_id"]


def get_beneath_instance_url(instance_id):
  return f"{config.BENEATH_BASE_URL}/streams/instances/{instance_id}"


def current_milli_time():
  return int(time.time() * 1000)


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR),
       reraise=True)
def get_latest_block_from_gateway():
  ''' Get the most recent block that was sent to the gateway '''
  LOG.info("get_latest_block_from_gateway from %s",
           BENEATH_GET_LATEST_BLOCK_URL)
  response = requests.get(BENEATH_GET_LATEST_BLOCK_URL)
  response.raise_for_status()
  return response.json()


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR),
       reraise=True)
def get_block_from_gateway(block_number):
  response = requests.get(f"{BENEATH_STREAM_URL}?number={block_number}")
  response.raise_for_status()
  return response.json()


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR),
       reraise=True)
def get_block_from_web3(block_no):
  return W3.eth.getBlock(block_no)


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR))
def post_block_to_gateway(instance_id, block_number, block_hash,
                          block_parent_hash, block_timestamp):
  headers = {
      "Authorization": f"Bearer {config.BENEATH_PROJECT_KEY}",
      "content-type": "application/json"
  }

  post_json = {
      "@meta": {
          "sequence_number": current_milli_time()
      },
      "number": block_number,
      "hash": block_hash,
      "parentHash": block_parent_hash,
      "timestamp": block_timestamp
  }

  response = requests.post(get_beneath_instance_url(instance_id),
                           json=post_json,
                           headers=headers)
  response.raise_for_status()

  LOG.info("Block posted successfully to gateway")


def main():
  instance_id = get_stream_instance_id()
  LOG.info("Got gateway instance ID: %s", instance_id)

  # Get the newest block from the gateway, so we know where to start syncing
  latest_gateway_block = get_latest_block_from_gateway()
  block_no = latest_gateway_block['number'] + 1
  last_block_hash = latest_gateway_block['hash']
  LOG.info("Starting sync from block %s", block_no)

  newest_block = get_block_from_web3('latest')
  target_block_no = int(newest_block.number)
  LOG.info("Current newest block from Web3 is %s", target_block_no)

  while True:
    # If next_block_no > target_block_no, then we have reached
    # our target block number and need to ask for a new target.
    while block_no > target_block_no:
      LOG.info("Target block reached. Asking for latest block number...")
      latest_block_no = get_block_from_web3('latest').number
      if block_no > latest_block_no:
        LOG.info("No new block(s) found, waiting 5 sec...")
        time.sleep(5)
      else:
        LOG.info("New block number found: %s", latest_block_no)
        target_block_no = latest_block_no

    # Get next block from Web3
    block = get_block_from_web3(block_no)
    LOG.info("Got block %s with hash %s and parent hash %s", block.number,
             block.hash.hex(), block.parentHash.hex())

    # Compare the new block's parent hash with the last blocks hash.
    # If they don't match, we're on a fork that have been abandoned in a
    # chain reorg! Log warning and handle it, by walking backwards through
    # the blocks from Web3 and compare them with gateway block hashes
    # until they match.
    # That's the root of the fork and we can continue syncing from there.
    if block.parentHash.hex() != last_block_hash:
      LOG.warning(
          "Block %s with hash %s expected to have parent hash %s, but it is %s",
          block.number, block.hash.hex(), last_block_hash,
          block.parentHash.hex())
      LOG.info("Finding the last common hash")
      block_no -= 1
      gateway_block = get_block_from_gateway(block_no)
      block = get_block_from_web3(block_no)
      while gateway_block["hash"] != block.hash.hex():
        LOG.info("At block %s Web3 hash is %s and gateway hash is %s", block_no,
                 block.hash.hex(), gateway_block["hash"])
        if block_no <= 0:
          raise Exception(
              "Could not find any block hash in gateway matching Web3 block hashes! Aborting."
          )
        block_no -= 1
        gateway_block = get_block_from_gateway(block_no)
        block = get_block_from_web3(block_no)
      LOG.info("Common hash found at block %s. Continuing sync.", block_no)

    # Send the block to the gateway
    post_block_to_gateway(instance_id, block.number, block.hash.hex(),
                          block.parentHash.hex(), block.timestamp)

    # Continue to the next block number
    last_block_hash = block.hash.hex()
    block_no += 1


if __name__ == "__main__":
  main()
