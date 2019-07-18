import unittest
from unittest.mock import Mock, call

from hexbytes import HexBytes
from web3.datastructures import AttributeDict

import fetch_and_publish_blocks as fpb


class Test_FetchAndPublish(unittest.TestCase):
    def test_get_blocks_with_web3(self):
        fpb.START_FROM_BLOCK_NO = 8123121

        # Mock w3.eth.getBlock() to control its behavior
        mock_getBlock = Mock()
        latest_call_counter = 0

        def getBlock_return_values(arg):
            blocks = {
                "latest": AttributeDict({"number": 8123123, "hash": HexBytes("0x123"), "parentHash": HexBytes("0x122")}),
                8123121: AttributeDict({"number": 8123121, "hash": HexBytes("0x121"), "parentHash": HexBytes("0x120")}),
                8123122: AttributeDict({"number": 8123122, "hash": HexBytes("0x122"), "parentHash": HexBytes("0x121")}),
                8123123: AttributeDict({"number": 8123123, "hash": HexBytes("0x123"), "parentHash": HexBytes("0x122")}),
            }
            if arg == "latest":
                # When "latest" is called the third time, raise an exception so we can test results
                nonlocal latest_call_counter
                latest_call_counter += 1
                if latest_call_counter >= 2:
                    raise Exception('stop loop for test assertion purposes')
                else:
                    return blocks["latest"]
            else:
                return blocks[arg]

        mock_getBlock.side_effect = getBlock_return_values
        fpb.w3.eth.getBlock = mock_getBlock

        # Run the main loop until the test-exception stops it
        try:
            fpb.main()
        except Exception as e:
            print(str(e))
            pass

        # Assert that w3.eth.getBlock() was called exactly as we expected
        fpb.w3.eth.getBlock.assert_has_calls([
            call("latest"), call(8123121), call(8123122), call(8123123), call("latest")
        ], any_order=False)
        self.assertEqual(fpb.w3.eth.getBlock.call_count, 5, "must have been called exactly 5 times")


if __name__ == '__main__':
    unittest.main()
