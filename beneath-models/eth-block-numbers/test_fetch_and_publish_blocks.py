import importlib
import unittest
from unittest.mock import Mock, call

from hexbytes import HexBytes
from web3.datastructures import AttributeDict

import fetch_and_publish_blocks as fpb


class Test_FetchAndPublishBlocks(unittest.TestCase):

  def setUp(self):
    # Reload fetch_and_publish_blocks so no side effects from other tests are carried over
    importlib.reload(fpb)

  def test_get_blocks_with_web3(self):
    # Mock the current time, so we get predictable results
    current_time = 123123123
    fpb.current_milli_time = Mock(return_value=current_time)
    current_instance_id = "123TestInstanceId"

    # Mock the latest block synced, so we get predictable results
    fpb.get_start_block_no = Mock(return_value=8123121)

    # Mock W3.eth.getBlock() to control its behavior
    mock_get_block = Mock()
    latest_call_counter = 0

    def getBlock_return_values(*arg, **kwargs):
      blocks = {
          "latest":
              AttributeDict({
                  "number": 8123123,
                  "hash": HexBytes("3123"),
                  "parentHash": HexBytes("3122"),
                  "timestamp": 10003
              }),
          8123121:
              AttributeDict({
                  "number": 8123121,
                  "hash": HexBytes("3121"),
                  "parentHash": HexBytes("3120"),
                  "timestamp": 10001
              }),
          8123122:
              AttributeDict({
                  "number": 8123122,
                  "hash": HexBytes("3122"),
                  "parentHash": HexBytes("3121"),
                  "timestamp": 10002
              }),
          8123123:
              AttributeDict({
                  "number": 8123123,
                  "hash": HexBytes("3123"),
                  "parentHash": HexBytes("3122"),
                  "timestamp": 10003
              }),
      }
      if arg[0] == "latest":
        # When "latest" is called the third time, raise an exception so we can test results
        nonlocal latest_call_counter
        latest_call_counter += 1
        if latest_call_counter >= 2:
          raise KeyboardInterrupt("stop loop for test assertion purposes")
        return blocks["latest"]
      else:
        return blocks[arg[0]]

    mock_get_block.side_effect = getBlock_return_values
    fpb.W3.eth.getBlock = mock_get_block

    # Mock requests.get to control responses and inspect GET calls made to the gateway
    def request_get_handler(*args, **kwargs):
      response = Mock()
      response.raise_for_status = lambda: True
      response.status_code = 200
      data = {
          fpb.BENEATH_GET_LATEST_BLOCK_URL: {
              "number": 8123123,
              "hash": "0x3123",
              "parentHash": "0x3122",
              "timestamp": 10003
          },
          f"{fpb.BENEATH_STREAM_URL}/details": {
              "current_instance_id": current_instance_id
          }
      }
      response.json.return_value = data[args[0]]
      return response

    fpb.requests.get = Mock(side_effect=request_get_handler)

    # Mock requests.post to control behaviour and inspect calls made to it
    fpb.requests.post = Mock(return_value=AttributeDict({
        "status_code": 200,
        "raise_for_status": lambda: True
    }))

    # Run the main loop until the test-exception stops it
    try:
      fpb.main()
    except KeyboardInterrupt as ex:
      print(str(ex))

    # Assert that W3.eth.getBlock(...) was called exactly as we expected
    fpb.W3.eth.getBlock.assert_has_calls([
        call("latest"),
        call(8123121),
        call(8123122),
        call(8123123),
        call("latest")
    ],
                                         any_order=False)
    self.assertEqual(fpb.W3.eth.getBlock.call_count, 5,
                     "getBlock must have been called exactly 5 times")

    # Assert that the gateway was asked for the latest synced block
    fpb.get_start_block_no.assert_called_once()

    # Assert that blocks where POST'ed to the gateway as we expected
    expected_post_headers = {
        "Authorization": f"Bearer {fpb.config.BENEATH_PROJECT_KEY}",
        "content-type": "application/json"
    }
    fpb.requests.post.assert_has_calls([
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 8123121,
                 "hash": "0x3121",
                 "parentHash": "0x3120",
                 "timestamp": 10001
             }),
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 8123122,
                 "hash": "0x3122",
                 "parentHash": "0x3121",
                 "timestamp": 10002
             }),
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 8123123,
                 "hash": "0x3123",
                 "parentHash": "0x3122",
                 "timestamp": 10003
             })
    ],
                                       any_order=False)
    self.assertEqual(fpb.requests.post.call_count, 3,
                     "requests.post must have been called exactly 3 times")


if __name__ == "__main__":
  unittest.main()
