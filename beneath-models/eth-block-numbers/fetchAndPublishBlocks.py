from web3 import Web3
import time

PROVIDER_URL = "https://mainnet.infura.io/v3/4e5fbbd54aef484daabd76627d8f5dc9"
START_FROM_BLOCK_NO = 8167540

web3 = Web3(Web3.HTTPProvider(PROVIDER_URL))

current_block = START_FROM_BLOCK_NO
newest_block = int(web3.eth.getBlock('latest').number) # TODO: Try/Catch

while True:
    if current_block > newest_block:
        while current_block > newest_block:
          print("Getting newest block number...")
          newest_block = web3.eth.getBlock('latest').number
          if current_block > newest_block:
            print("No new block(s) found, waiting 5 sec...")
            time.sleep(5)

    block = web3.eth.getBlock(current_block)
    print(f"Block {block.number} hash is {block.hash.hex()}")

    current_block += 1
