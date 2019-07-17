from web3 import Web3
import time
import requests

PROVIDER_URL = "https://mainnet.infura.io/v3/4e5fbbd54aef484daabd76627d8f5dc9"
START_FROM_BLOCK_NO = 8168249
web3 = Web3(Web3.HTTPProvider(PROVIDER_URL))

def get_newer_block_no(current_block, newest_block_no):
    while current_block > newest_block_no:
        print("Getting newest block number...")
        block_no = web3.eth.getBlock('latest').number
        if current_block > block_no:
            print("No new block(s) found, waiting 5 sec...")
            time.sleep(5)
        else:
            return block_no

def main():
    current_block = START_FROM_BLOCK_NO
    newest_block = web3.eth.getBlock('latest')  # TODO: Try/Catch
    print("Newest block:", newest_block)
    newest_block_no = int(newest_block.number)

    while True:
        if current_block > newest_block_no:
            newest_block_no = get_newer_block_no(current_block, newest_block_no)

        block = web3.eth.getBlock(current_block)
        print(f"Block {block.number} hash is {block.hash.hex()}, with parent hash {block.parentHash.hex()}")

        current_block += 1

if __name__ == "__main__":
    main()
