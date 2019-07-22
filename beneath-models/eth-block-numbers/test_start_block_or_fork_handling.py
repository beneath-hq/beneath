import importlib
import unittest
from unittest.mock import Mock, call

from hexbytes import HexBytes
from web3.datastructures import AttributeDict

import fetch_and_publish_blocks as fpb


class Test_ForkHandling(unittest.TestCase):

    def setUp(self):
        # Reload fetch_and_publish_blocks so no side effects from other tests are carried over
        importlib.reload(fpb)

    def test_start_block_or_fork_handling(self):
        # Mock the current time, so we get predictable results
        current_time = 123123123
        fpb.current_milli_time = Mock(return_value=current_time)
        current_instance_id = "123TestInstanceId"

        # Mock W3.eth.getBlock() to control its behavior
        latest_call_counter = 0

        def getBlock_return_values(arg):
            blocks = {
                "latest":
                    AttributeDict({
                        "number": 124,
                        "hash": HexBytes("0xb124"),
                        "parentHash": HexBytes("0xb123"),
                        "timestamp": 10004
                    }),
                124:
                    AttributeDict({
                        "number": 124,
                        "hash": HexBytes("0xb124"),
                        "parentHash": HexBytes("0xb123"),
                        "timestamp": 10004
                    }),
                123:
                    AttributeDict({
                        "number": 123,
                        "hash": HexBytes("0xb123"),
                        "parentHash": HexBytes("0xb122"),
                        "timestamp": 10003
                    }),
                122:
                    AttributeDict({
                        "number": 122,
                        "hash": HexBytes("0xb122"),
                        "parentHash": HexBytes("0xb121"),
                        "timestamp": 10002
                    }),
                121:
                    AttributeDict({
                        "number": 121,
                        "hash": HexBytes("0xb121"),
                        "parentHash": HexBytes("0xb120"),
                        "timestamp": 10001
                    }),
                120:
                    AttributeDict({
                        "number": 120,
                        "hash": HexBytes("0xa120"),
                        "parentHash": HexBytes("0xa119"),
                        "timestamp": 10001
                    }),
            }
            if arg == "latest":
                # When "latest" is called the third time, raise an exception so we can test results
                nonlocal latest_call_counter
                latest_call_counter += 1
                if latest_call_counter >= 2:
                    raise KeyboardInterrupt(
                        "stop loop for test assertion purposes")
                else:
                    return blocks["latest"]
            else:
                return blocks[arg]

        fpb.W3.eth.getBlock = Mock(side_effect=getBlock_return_values)

        # Mock requests.post to control responses and inspect POST calls made to the gateway
        fpb.requests.post = Mock(
            return_value=AttributeDict({"status_code": 200}))

        # Mock requests.get to control responses and inspect GET calls made to the gateway
        def request_get_handler(*args, **kwargs):
            response = Mock()
            response.status_code = 200
            data = {
                fpb.BENEATH_GET_LATEST_BLOCK_URL: {
                    "number": 123,
                    "hash": "0xa123",
                    "parentHash": "0xa122",
                    "timestamp": 1500000003
                },
                f"{fpb.BENEATH_STREAM_URL}?number=122": {
                    "number": 122,
                    "hash": "0xa122",
                    "parentHash": "0xa121",
                    "timestamp": 1500000002
                },
                f"{fpb.BENEATH_STREAM_URL}?number=121": {
                    "number": 121,
                    "hash": "0xa121",
                    "parentHash": "0xa120",
                    "timestamp": 1500000001
                },
                f"{fpb.BENEATH_STREAM_URL}?number=120": {
                    "number": 120,
                    "hash": "0xa120",
                    "parentHash": "0xa119",
                    "timestamp": 1500000000
                },
                f"{fpb.BENEATH_STREAM_URL}/details": {
                    "current_instance_id": current_instance_id
                }
            }
            response.json.return_value = data[args[0]]
            return response

        fpb.requests.get = Mock(side_effect=request_get_handler)

        # Run the main loop until the test-exception stops it
        try:
            fpb.main()
        except KeyboardInterrupt as ex:
            print(str(ex))

        # Assert that W3.eth.getBlock(...) was called exactly as we expected,
        # both to verify gateway block hashes and to get new blocks to POST.
        fpb.W3.eth.getBlock.assert_has_calls([
            call(123),
            call(122),
            call(121),
            call(120),
            call("latest"),
            call(121),
            call(122),
            call(123),
            call(124),
            call("latest")
        ],
                                             any_order=False)
        self.assertEqual(fpb.W3.eth.getBlock.call_count, 10,
                         "getBlock must have been called exactly 10 times")

        # Assert that the gateway was called to compare its blocks hashes with current block hashes from web3
        fpb.requests.get.assert_has_calls([
            call(f"{fpb.BENEATH_STREAM_URL}/details",
                 headers={"Bearer": fpb.config.BENEATH_PROJECT_KEY}),
            call(fpb.BENEATH_GET_LATEST_BLOCK_URL),
            call(f"{fpb.BENEATH_STREAM_URL}?number=122"),
            call(f"{fpb.BENEATH_STREAM_URL}?number=121"),
            call(f"{fpb.BENEATH_STREAM_URL}?number=120")
        ])
        self.assertEqual(fpb.requests.get.call_count, 5,
                         "requests.get must have been called exactly 5 times")

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
                     "number": 121,
                     "hash": "0xb121",
                     "parentHash": "0xb120",
                     "timestamp": 10001
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
                     "timestamp": 10002
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
                     "timestamp": 10003
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
                     "timestamp": 10004
                 })
        ])
        self.assertEqual(fpb.requests.post.call_count, 4,
                         "requests.post must have been called exactly 4 times")


if __name__ == "__main__":
    unittest.main()
