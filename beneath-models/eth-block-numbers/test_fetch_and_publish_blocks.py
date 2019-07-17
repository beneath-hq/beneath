import unittest
from unittest.mock import call, Mock
import fetch_and_publish_blocks as fpb
from web3.datastructures import (AttributeDict)
from hexbytes import HexBytes

class TestFetchAndPublish(unittest.TestCase):

    def test_get_blocks_with_web3(self):
        fpb.START_FROM_BLOCK_NO = 8123121

        # Mock web3.eth.getBlock() to control its behavior
        mock_getBlock = Mock()
        latest_call_counter = 0

        def getBlock_return_values(arg):
            if arg == "latest":
                # 
                nonlocal latest_call_counter
                latest_call_counter += 1
                if latest_call_counter >= 2:
                   raise Exception('stop loop for test assertion purposes')
                else:
                    return AttributeDict({"number": 8123123, "hash": HexBytes("0x123"), "parentHash": HexBytes("0x122")})
            elif arg == 8123121:
                return AttributeDict({"number": 8123121, "hash": HexBytes("0x121"), "parentHash": HexBytes("0x120")})
            elif arg == 8123122:
                return AttributeDict({"number": 8123122, "hash": HexBytes("0x122"), "parentHash": HexBytes("0x121")})
            elif arg == 8123123:
                return AttributeDict({"number": 8123123, "hash": HexBytes("0x123"), "parentHash": HexBytes("0x122")})

        mock_getBlock.side_effect = getBlock_return_values
        fpb.web3.eth.getBlock = mock_getBlock
        
        # Run the main loop until the test-exception stops it
        try:
          fpb.main()
        except Exception as e:
          print(str(e))
          pass
        
        # Assert that web3.eth.getBlock() was called exactly as we expected
        fpb.web3.eth.getBlock.assert_has_calls([
            call("latest"), call(8123121), call(8123122), call(8123123), call("latest")
        ], any_order=False)
        self.assertEqual(fpb.web3.eth.getBlock.call_count, 5, "must have been called exactly 5 times")

if __name__ == '__main__':
    unittest.main()
