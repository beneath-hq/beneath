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
       before_sleep=before_sleep_log(LOG, logging.ERROR))
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
       before_sleep=before_sleep_log(LOG, logging.ERROR))
def get_latest_block_from_gateway():
    ''' Get the most recent block that was sent to the gateway '''
    LOG.info("get_latest_block_from_gateway from %s",
             BENEATH_GET_LATEST_BLOCK_URL)
    response = requests.get(BENEATH_GET_LATEST_BLOCK_URL)
    response.raise_for_status()
    return response.json()


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR))
def get_block_from_gateway(block_number):
    response = requests.get(f"{BENEATH_STREAM_URL}?number={block_number}")
    response.raise_for_status()
    return response.json()


@retry(wait=wait_random(min=5, max=10),
       stop=stop_after_attempt(5),
       before_sleep=before_sleep_log(LOG, logging.ERROR))
def get_block_from_web3(block):
    return W3.eth.getBlock(block)


def get_start_block_no():
    gateway_block = get_latest_block_from_gateway()
    # Compare gateway block hash with same blocknumbers hash from web3
    web3_block = get_block_from_web3(gateway_block['number'])
    while gateway_block['hash'] != web3_block.hash.hex():
        # If hashes don't match, a fork probably happened.
        # Check previous block until hash matches, so we are behind the fork and can go forward again.
        LOG.warning(
            "Warning! Gateway block %s with hash %s does not match Web3 block %s with hash %s",
            gateway_block['number'], gateway_block['hash'], web3_block.number,
            web3_block.hash.hex())
        LOG.info("Trying the previous block...")
        if gateway_block['number'] > 0:
            gateway_block = get_block_from_gateway(gateway_block['number'] - 1)
            web3_block = get_block_from_web3(gateway_block['number'])
        else:
            raise Exception(
                "Could not find any block hash in gateway matching Web3 block hashes!"
            )
    return gateway_block['number'] + 1


def get_newer_block_no(current_block_no, target_block_no):
    while current_block_no > target_block_no:
        LOG.info("Sync finished. Checking for newer block number...")
        block_no = get_block_from_web3('latest').number
        if current_block_no > block_no:
            LOG.info("No new block(s) found, waiting 5 sec...")
            time.sleep(5)
        else:
            LOG.info("New block number found: %s", block_no)
            return block_no


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

    current_block_no = get_start_block_no()
    LOG.info("Starting sync from block %s", current_block_no)

    newest_block = get_block_from_web3('latest')
    target_block_no = int(newest_block.number)
    LOG.info("Current newest block from Web3 is %s", target_block_no)

    while True:
        if current_block_no > target_block_no:
            target_block_no = get_newer_block_no(current_block_no, target_block_no)

        block = get_block_from_web3(current_block_no)

        LOG.info("Block %s hash is %s, with parent hash %s", block.number,
                 block.hash.hex(), block.parentHash.hex())

        # Keep previous block and compare its hash to the new block's parent hash.
        # If they are not the same, we have a chain reorg! Log warning and handle it,
        # by walking backwards and compare with gateway block hashes until they are the same.

        post_block_to_gateway(instance_id, block.number, block.hash.hex(),
                              block.parentHash.hex(), block.timestamp)

        current_block_no += 1


if __name__ == "__main__":
    main()
