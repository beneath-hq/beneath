import time

import requests
from web3 import Web3

WEB3_PROVIDER_URL = "https://mainnet.infura.io/v3/4e5fbbd54aef484daabd76627d8f5dc9"
BENEATH_POST_BLOCK_URL = "https://beneath.network/projects/ethereum/streams/block-numbers"
START_FROM_BLOCK_NO = 8173297
w3 = Web3(Web3.HTTPProvider(WEB3_PROVIDER_URL))


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
        "blockParentHash": block_parent_hash
    }

    response = requests.post(
        BENEATH_POST_BLOCK_URL,
        json=post_json,
        headers=headers
    )

    if (response.status_code > 200):
        print("Error posting block to gateway!", response)
    else:
        print("Block posted successfully to gateway")


def main():
    current_block = START_FROM_BLOCK_NO
    newest_block = w3.eth.getBlock('latest')  # TODO: Try/Catch
    print("Newest block:", newest_block)
    newest_block_no = int(newest_block.number)

    while True:
        if current_block > newest_block_no:
            newest_block_no = get_newer_block_no(current_block, newest_block_no)

        block = w3.eth.getBlock(current_block)
        print(f"Block {block.number} hash is {block.hash.hex()}, with parent hash {block.parentHash.hex()}")

        post_block_to_gateway(block.number, block.hash.hex(), block.parentHash.hex())

        current_block += 1


if __name__ == "__main__":
    main()
