import importlib
import unittest
from unittest.mock import Mock, call

from hexbytes import HexBytes
from web3.datastructures import AttributeDict

import fetch_and_publish_blocks as fpb


class TestFetchAndPublishBlocks(unittest.TestCase):

  def setUp(self):
    # Reload fetch_and_publish_blocks so no side effects from other tests are carried over
    importlib.reload(fpb)

  def test_fetch_and_publish_blocks(self):
    # Mock the current time, so we get predictable results
    current_time = 123123123
    fpb.current_milli_time = Mock(return_value=current_time)

    current_instance_id = "123TestInstanceId"

    # Mock get_block_from_web3() to control its behavior
    mock_get_block = Mock()
    latest_call_counter = 0

    def web3_return_values(*arg, **kwargs):
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

    mock_get_block.side_effect = web3_return_values
    fpb.get_block_from_web3 = mock_get_block

    # Mock get_stream_instance_id()
    fpb.get_stream_instance_id = Mock(return_value=current_instance_id)

    # Mock get_latest_block_from_gateway()
    latest_block_from_gateway = {
        "number": 8123121,
        "hash": "0x3121",
        "parentHash": "0x3122",
        "timestamp": 10003
    }
    fpb.get_latest_block_from_gateway = Mock(
        return_value=latest_block_from_gateway)

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

    # Assert that get_block_from_web3(...) was called exactly as we expected
    fpb.get_block_from_web3.assert_has_calls(
        [call("latest"),
         call(8123122),
         call(8123123),
         call("latest")],
        any_order=False)
    self.assertEqual(
        fpb.get_block_from_web3.call_count, 4,
        "get_block_from_web3 must have been called exactly 4 times")

    # Assert that the gateway was asked for the latest synced block
    fpb.get_latest_block_from_gateway.assert_called_once()

    # Assert that blocks were POST'ed to the gateway as we expected
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
    self.assertEqual(fpb.requests.post.call_count, 2,
                     "requests.post must have been called exactly 2 times")


if __name__ == "__main__":
  unittest.main()
