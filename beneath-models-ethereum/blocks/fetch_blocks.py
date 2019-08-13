import time
from datetime import datetime
from web3 import Web3
from beneath import Client
from hexbytes import HexBytes
from decimal import Decimal

WEB3_PROVIDER_URL = "https://mainnet.infura.io/v3/c7cb3618405a48d5bb6392ccd6c2ad1c"

w3 = Web3(Web3.HTTPProvider(WEB3_PROVIDER_URL))
client = Client()
stream = client.stream(project="ethereum", stream="blocks")

def run():
  current = get_block("latest")["number"] - 13
  while True:
    next = get_block("latest")["number"] - 12
    if next > current:
      for num in range(current + 1, next + 1):
        block = get_block(num)
        save_block(block)
      current = next
    time.sleep(5)


def get_block(num):
  return w3.eth.getBlock(num)


def save_block(block):
  save = {
    "number": block["number"],
    "timestamp": datetime.fromtimestamp(block["timestamp"]),
    "hash": bytes(block["hash"]),
    "parentHash": bytes(block["parentHash"]),
    "miner": bytes(HexBytes(block["miner"])),
    "size": block["size"],
    "transactions": len(block["transactions"]),
    "difficulty": Decimal(block["difficulty"]),
    "extraData": bytes(block["extraData"]),
    "gasLimit": block["gasLimit"],
    "gasUsed": block["gasUsed"],
  }
  stream.write_records(stream.current_instance_id, [save])
  print("Saved block {}".format(save["number"]))


if __name__ == '__main__':
  run()
