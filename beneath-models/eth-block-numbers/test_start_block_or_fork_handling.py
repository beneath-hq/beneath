import importlib
import unittest
from unittest.mock import Mock, call

from hexbytes import HexBytes
from web3.datastructures import AttributeDict

import fetch_and_publish_blocks as fpb


class TestForkHandling(unittest.TestCase):

  def setUp(self):
    # Reload fetch_and_publish_blocks so no side effects from other tests are carried over
    importlib.reload(fpb)

  def test_fork_handling(self):
    # Mock the current time, so we get predictable results
    current_time = 123123123
    fpb.current_milli_time = Mock(return_value=current_time)
    current_instance_id = "123TestInstanceId"

    # Mock get_block_from_web3() to control its behavior
    latest_call_counter = 0

    def web3_return_values(arg):
      blocks = {
          "latest":
              AttributeDict({
                  "number": 124,
                  "hash": HexBytes("0xb124"),
                  "parentHash": HexBytes("0xb123"),
                  "timestamp": 1500000004
              }),
          124:
              AttributeDict({
                  "number": 124,
                  "hash": HexBytes("0xb124"),
                  "parentHash": HexBytes("0xb123"),
                  "timestamp": 1500000004
              }),
          123:
              AttributeDict({
                  "number": 123,
                  "hash": HexBytes("0xb123"),
                  "parentHash": HexBytes("0xb122"),
                  "timestamp": 1500000003
              }),
          122:
              AttributeDict({
                  "number": 122,
                  "hash": HexBytes("0xb122"),
                  "parentHash": HexBytes("0xb121"),
                  "timestamp": 1500000002
              }),
          121:
              AttributeDict({
                  "number": 121,
                  "hash": HexBytes("0xb121"),
                  "parentHash": HexBytes("0xa120"),
                  "timestamp": 1500000001
              }),
          120:
              AttributeDict({
                  "number": 120,
                  "hash": HexBytes("0xa120"),
                  "parentHash": HexBytes("0xa119"),
                  "timestamp": 1500000000
              }),
      }
      if arg == "latest":
        # When "latest" is called the third time, raise an exception so we can test results
        nonlocal latest_call_counter
        latest_call_counter += 1
        if latest_call_counter >= 2:
          raise KeyboardInterrupt("stop loop for test assertion purposes")
        else:
          return blocks["latest"]
      else:
        return blocks[arg]

    fpb.get_block_from_web3 = Mock(side_effect=web3_return_values)

    # Mock get_stream_instance_id()
    fpb.get_stream_instance_id = Mock(return_value=current_instance_id)

    # Mock get_latest_block_from_gateway()
    latest_block_from_gateway = {
        "number": 123,
        "hash": "0xa123",
        "parentHash": "0xa122",
        "timestamp": 1500000003
    }
    fpb.get_latest_block_from_gateway = Mock(
        return_value=latest_block_from_gateway)

    # Mock get_block_from_gateway()
    def block_from_gateway(block_no):
      blocks = {
          123: {
              "number": 123,
              "hash": "0xa123",
              "parentHash": "0xa122",
              "timestamp": 1500000003
          },
          122: {
              "number": 122,
              "hash": "0xa122",
              "parentHash": "0xa121",
              "timestamp": 1500000002
          },
          121: {
              "number": 121,
              "hash": "0xa121",
              "parentHash": "0xa120",
              "timestamp": 1500000001
          },
          120: {
              "number": 120,
              "hash": "0xa120",
              "parentHash": "0xa119",
              "timestamp": 1500000000
          }
      }
      return blocks[block_no]

    fpb.get_block_from_gateway = Mock(side_effect=block_from_gateway)

    # Mock requests.post to control responses and inspect POST calls made to the gateway
    fpb.requests.post = Mock(return_value=AttributeDict({
        "status_code": 200,
        "raise_for_status": lambda: True
    }))

    # Run the main loop until the test-exception stops it
    try:
      fpb.main()
    except KeyboardInterrupt as ex:
      print(str(ex))

    # Assert that get_block_from_web3(...) was called exactly as we expected,
    # both to verify gateway block hashes and to get new blocks to POST.
    fpb.get_block_from_web3.assert_has_calls([
        call("latest"),
        call(124),
        call(123),
        call(122),
        call(121),
        call(120),
        call(121),
        call(122),
        call(123),
        call(124),
        call("latest")
    ],
                                             any_order=False)
    self.assertEqual(
        fpb.get_block_from_web3.call_count, 11,
        "get_block_from_web3 must have been called exactly 11 times")

    # Assert that we asked for the stream instance ID
    fpb.get_stream_instance_id.assert_called_once()

    # Assert that the gateway was asked for its latest block
    fpb.get_latest_block_from_gateway.assert_called_once()

    # Assert that the gateway was called to compare its blocks hashes with current block hashes from web3
    fpb.get_block_from_gateway.assert_has_calls(
        [call(123), call(122), call(121),
         call(120)], any_order=False)
    self.assertEqual(
        fpb.get_block_from_gateway.call_count, 4,
        "get_block_from_gateway(...) must have been called exactly 4 times")

    # Assert that the new blocks gets POST"ed to the gateway
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
                 "number": 120,
                 "hash": "0xa120",
                 "parentHash": "0xa119",
                 "timestamp": 1500000000
             }),
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 121,
                 "hash": "0xb121",
                 "parentHash": "0xa120",
                 "timestamp": 1500000001
             }),
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 122,
                 "hash": "0xb122",
                 "parentHash": "0xb121",
                 "timestamp": 1500000002
             }),
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 123,
                 "hash": "0xb123",
                 "parentHash": "0xb122",
                 "timestamp": 1500000003
             }),
        call(fpb.get_beneath_instance_url(current_instance_id),
             headers=expected_post_headers,
             json={
                 "@meta": {
                     "sequence_number": current_time
                 },
                 "number": 124,
                 "hash": "0xb124",
                 "parentHash": "0xb123",
                 "timestamp": 1500000004
             })
    ],
                                       any_order=False)
    self.assertEqual(fpb.requests.post.call_count, 5,
                     "requests.post must have been called exactly 5 times")


if __name__ == "__main__":
  unittest.main()
