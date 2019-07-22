import logging as log
import time

import requests
from requests.exceptions import RequestException
from web3 import Web3

import config

BENEATH_STREAM_URL = f"{config.BENEATH_BASE_URL}/projects/{config.BENEATH_PROJECT}/streams/{config.BENEATH_PROJECT_STREAM}"
BENEATH_GET_LATEST_BLOCK_URL = "http://not.working.yet/"  # f"{config.BENEATH_BASE_URL}/projects/{config.BENEATH_PROJECT}/streams/{config.BENEATH_PROJECT_STREAM}?get-latest-record"

log.basicConfig(level=log.INFO)
W3 = Web3(Web3.HTTPProvider(config.WEB3_PROVIDER_URL))


def get_beneath_instance_url(instance_id):
    return f"{config.BENEATH_BASE_URL}/streams/instances/{instance_id}"


def current_milli_time():
    return int(time.time() * 1000)


def get_block_from_gateway(block_number):
    return requests.get(f"{BENEATH_STREAM_URL}?number={block_number}").json()


def get_start_block():
    # First get the most recent block sent to the gateway
    try:
        response = requests.get(BENEATH_GET_LATEST_BLOCK_URL)
    except RequestException as ex:
        # If gateway cannot be reached, start from block 0
        log.error("Gateway could not be reached on %s\n%s",
                  BENEATH_GET_LATEST_BLOCK_URL, ex)
        return 0

    if response.status_code <= 200:
        gateway_block = response.json()
    else:
        # If gateway respond with an error, start from block 0
        return 0

    # Compare gateway block hash with same blocknumbers hash from web3
    web3_block = W3.eth.getBlock(gateway_block['number'])
    while gateway_block['hash'] != web3_block.hash.hex():
        # If hashes don't match, a fork probably happened. Check previous block until hash matches, so we are behind the fork and can go forward again.
        log.warning(
            "Warning! Gateway block %s with hash %s does not match Web3 block %s with hash %s",
            gateway_block['number'], gateway_block['hash'], web3_block.number,
            web3_block.hash.hex())
        log.info("Trying the previous block...")
        if gateway_block['number'] > 0:
            gateway_block = get_block_from_gateway(gateway_block['number'] - 1)
            web3_block = W3.eth.getBlock(gateway_block['number'])
        else:
            raise Exception(
                "Could not find any block hash in gateway matching Web3 block hashes!"
            )
    return gateway_block['number'] + 1


def get_newer_block_no(current_block, newest_block_no):
    while current_block > newest_block_no:
        log.info("Sync finished. Checking for newer block number...")
        block_no = W3.eth.getBlock('latest').number
        if current_block > block_no:
            log.info("No new block(s) found, waiting 5 sec...")
            time.sleep(5)
        else:
            log.info("New block number found: %s", block_no)
            return block_no


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

    if response.status_code > 200:
        log.error("Error posting block to gateway!\n%s", response)
    else:
        log.info("Block posted successfully to gateway")


def main():
    instance_id = requests.get(f"{BENEATH_STREAM_URL}/details",
                               headers={
                                   "Bearer": config.BENEATH_PROJECT_KEY
                               }).json()["current_instance_id"]
    log.info("Got gateway instance ID: %s", instance_id)

    current_block = get_start_block()
    log.info("Starting sync from block %s", current_block)

    newest_block = W3.eth.getBlock('latest')  # TODO: Try/Catch
    newest_block_no = int(newest_block.number)
    log.info("Current newest block from Web3 is %s", newest_block_no)

    while True:
        if current_block > newest_block_no:
            newest_block_no = get_newer_block_no(current_block, newest_block_no)

        block = W3.eth.getBlock(current_block)

        log.info("Block %s hash is %s, with parent hash %s", block.number,
                 block.hash.hex(), block.parentHash.hex())

        # Keep previous block and compare its hash to the new block's parent hash.
        # If they are not the same, we have a chain reorg! Log warning and handle it,
        # by walking backwards and compare with gateway block hashes until they are the same.

        post_block_to_gateway(instance_id, block.number, block.hash.hex(),
                              block.parentHash.hex(), block.timestamp)

        current_block += 1


if __name__ == "__main__":
    main()
