import time

import requests
from web3 import Web3

WEB3_PROVIDER_URL = "https://mainnet.infura.io/v3/4e5fbbd54aef484daabd76627d8f5dc9"
BENEATH_POST_BLOCK_URL = "https://beneath.network/projects/ethereum/streams/ethereum-blocks"
BENEATH_GET_LATEST_BLOCK_URL = "https://beneath.network/projects/ethereum/streams/ethereum-blocks?get-latest-record"
DEFAULT_START_BLOCK_NO = 0
w3 = Web3(Web3.HTTPProvider(WEB3_PROVIDER_URL))


def current_milli_time():
    return int(time.time() * 1000)


def get_block_from_gateway(block_number):
    return requests.get(
        f"https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number={block_number}"
    ).json()


def get_start_block():
    # First get the most recent block sent to the gateway
    response = requests.get(BENEATH_GET_LATEST_BLOCK_URL)
    if response.status_code <= 200:
        gateway_block = response.json()
    else:
        # If gateway fails to respond, return hard-coded block 0
        return {
            "blockNumber":
                0,
            "blockHash":
                "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
            "blockParentHash":
                "0x0000000000000000000000000000000000000000000000000000000000000000",
            "syncTimestamp":
                1563444444
        }

    # Compare gateway block hash with same blocknumbers hash from web3
    web3_block = w3.eth.getBlock(gateway_block['blockNumber'])
    while gateway_block['blockHash'] != web3_block.hash.hex():
        # If hashes don't match, a fork probably happened. Check previous block until hash matches, so we are behind the fork and can go forward again.
        print(
            f"Warning! Gateway block {gateway_block['blockNumber']} with hash {gateway_block['blockHash']} does not match Web3 block {web3_block.number} with hash {web3_block.hash.hex()}"
        )
        print("Trying the previous block...")
        if gateway_block['blockNumber'] > 0:
            gateway_block = get_block_from_gateway(
                gateway_block['blockNumber'] - 1)
            web3_block = w3.eth.getBlock(gateway_block['blockNumber'])
        else:
            raise Exception(
                "Could not find any block hash in gateway matching Web3 block hashes!"
            )
    return gateway_block['blockNumber'] + 1


def get_newer_block_no(current_block, newest_block_no):
    while current_block > newest_block_no:
        print("Getting newest block number...")
        block_no = w3.eth.getBlock('latest').number
        if current_block > block_no:
            print("No new block(s) found, waiting 5 sec...")
            time.sleep(5)
        else:
            return block_no


def post_block_to_gateway(block_number, block_hash, block_parent_hash):
    headers = {'content-type': 'application/json'}

    post_json = {
        "blockNumber": block_number,
        "blockHash": block_hash,
        "blockParentHash": block_parent_hash,
        "syncTimestamp": current_milli_time()
    }

    response = requests.post(BENEATH_POST_BLOCK_URL,
                             json=post_json,
                             headers=headers)

    if (response.status_code > 200):
        print("Error posting block to gateway!", response)
    else:
        print("Block posted successfully to gateway")


def main():
    current_block = get_start_block()
    newest_block = w3.eth.getBlock('latest')  # TODO: Try/Catch
    print("Newest block:", newest_block)
    newest_block_no = int(newest_block.number)

    while True:
        if current_block > newest_block_no:
            newest_block_no = get_newer_block_no(current_block, newest_block_no)

        block = w3.eth.getBlock(current_block)
        print(
            f"Block {block.number} hash is {block.hash.hex()}, with parent hash {block.parentHash.hex()}"
        )

        post_block_to_gateway(block.number, block.hash.hex(),
                              block.parentHash.hex())

        current_block += 1


if __name__ == "__main__":
    main()
