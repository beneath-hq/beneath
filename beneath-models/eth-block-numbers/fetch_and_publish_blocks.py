import logging as log
import time

import requests
from web3 import Web3

WEB3_PROVIDER_URL = "https://mainnet.infura.io/v3/4e5fbbd54aef484daabd76627d8f5dc9"
BENEATH_BASE_URL = "http://localhost:5000"
BENEATH_PROJECT = "ethereum"
BENEATH_PROJECT_KEY = "Ipv9X3mW3fwu+qwtfnJyalUdDOvxTtXG00V+NFEjvBg="
BENEATH_PROJECT_STREAM = "block-numbers"

beneath_stream_url = f"{BENEATH_BASE_URL}/projects/{BENEATH_PROJECT}/streams/{BENEATH_PROJECT_STREAM}"
beneath_get_latest_block_url = "http://not.working.yet/"  # f"{BENEATH_BASE_URL}/projects/{BENEATH_PROJECT}/streams/{BENEATH_PROJECT_STREAM}?get-latest-record"

log.basicConfig(level=log.INFO)
w3 = Web3(Web3.HTTPProvider(WEB3_PROVIDER_URL))


def get_beneath_instance_url(instance_id):
    return f"{BENEATH_BASE_URL}/streams/instances/{instance_id}"


def current_milli_time():
    return int(time.time() * 1000)


def get_block_from_gateway(block_number):
    return requests.get(f"{beneath_stream_url}?number={block_number}").json()


def get_start_block():
    # First get the most recent block sent to the gateway
    try:
        response = requests.get(beneath_get_latest_block_url)
    except Exception as e:
        # If gateway cannot be reached, start from block 0
        return 0

    if response.status_code <= 200:
        gateway_block = response.json()
    else:
        # If gateway respond with an error, start from block 0
        return 0

    # Compare gateway block hash with same blocknumbers hash from web3
    web3_block = w3.eth.getBlock(gateway_block['number'])
    while gateway_block['hash'] != web3_block.hash.hex():
        # If hashes don't match, a fork probably happened. Check previous block until hash matches, so we are behind the fork and can go forward again.
        log.warning(
            f"Warning! Gateway block {gateway_block['number']} with hash {gateway_block['hash']} does not match Web3 block {web3_block.number} with hash {web3_block.hash.hex()}"
        )
        log.info("Trying the previous block...")
        if gateway_block['number'] > 0:
            gateway_block = get_block_from_gateway(gateway_block['number'] - 1)
            web3_block = w3.eth.getBlock(gateway_block['number'])
        else:
            raise Exception(
                "Could not find any block hash in gateway matching Web3 block hashes!"
            )
    return gateway_block['number'] + 1


def get_newer_block_no(current_block, newest_block_no):
    while current_block > newest_block_no:
        log.info("Sync finished. Checking for newer block number...")
        block_no = w3.eth.getBlock('latest').number
        if current_block > block_no:
            log.info("No new block(s) found, waiting 5 sec...")
            time.sleep(5)
        else:
            log.info(f"New block number found: {block_no}")
            return block_no


def post_block_to_gateway(instance_id, block_number, block_hash,
                          block_parent_hash, block_timestamp):
    headers = {
        "Authorization": f"Bearer {BENEATH_PROJECT_KEY}",
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

    if (response.status_code > 200):
        log.error(f"Error posting block to gateway!\n{response}")
    else:
        log.info("Block posted successfully to gateway")


def main():
    instance_id = requests.get(f"{beneath_stream_url}/details",
                               headers={
                                   "Bearer": BENEATH_PROJECT_KEY
                               }).json()["current_instance_id"]
    log.info(f"Got gateway instance ID: {instance_id}")

    current_block = get_start_block()
    log.info(f"Starting sync from block {current_block}")

    newest_block = w3.eth.getBlock('latest')  # TODO: Try/Catch
    newest_block_no = int(newest_block.number)
    log.info(f"Current newest block from Web3 is {newest_block_no}")

    while True:
        if current_block > newest_block_no:
            newest_block_no = get_newer_block_no(current_block, newest_block_no)

        block = w3.eth.getBlock(current_block)

        log.info(
            f"Block {block.number} hash is {block.hash.hex()}, with parent hash {block.parentHash.hex()}"
        )

        # Keep previous block and compare its hash to the new block's parent hash.
        # If they are not the same, we have a chain reorg! Log warning and handle it,
        # by walking backwards and compare with gateway block hashes until they are the same.

        post_block_to_gateway(instance_id, block.number, block.hash.hex(),
                              block.parentHash.hex(), block.timestamp)

        current_block += 1


if __name__ == "__main__":
    main()
