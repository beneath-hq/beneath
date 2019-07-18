import unittest
from unittest.mock import Mock, call

from hexbytes import HexBytes
from web3.datastructures import AttributeDict

import fetch_and_publish_blocks as fpb


class Test_FetchAndPublishBlocks(unittest.TestCase):

    def test_get_blocks_with_web3(self):
        # Mock the current time, so we get predictable results
        CURRENT_TIME = 123123123
        fpb.current_milli_time = Mock(return_value=CURRENT_TIME)

        # Mock the latest block synced, so we get predictable results
        fpb.get_start_block = Mock(
            return_value={
                "blockNumber": 8123121,
                "blockHash": "0x3121",
                "blockParentHash": "0x3120"
            })

        # Mock w3.eth.getBlock() to control its behavior
        mock_getBlock = Mock()
        latest_call_counter = 0

        def getBlock_return_values(arg):
            blocks = {
                "latest":
                    AttributeDict({
                        "number": 8123123,
                        "hash": HexBytes("3123"),
                        "parentHash": HexBytes("3122")
                    }),
                8123121:
                    AttributeDict({
                        "number": 8123121,
                        "hash": HexBytes("3121"),
                        "parentHash": HexBytes("3120")
                    }),
                8123122:
                    AttributeDict({
                        "number": 8123122,
                        "hash": HexBytes("3122"),
                        "parentHash": HexBytes("3121")
                    }),
                8123123:
                    AttributeDict({
                        "number": 8123123,
                        "hash": HexBytes("3123"),
                        "parentHash": HexBytes("3122")
                    }),
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

        # Mock requests.post to control behaviour and inspect calls made to it
        fpb.requests.post = Mock(
            return_value=AttributeDict({"status_code": 200}))

        # Run the main loop until the test-exception stops it
        try:
            fpb.main()
        except Exception as e:
            print(str(e))
            pass

        # Assert that w3.eth.getBlock(...) was called exactly as we expected
        fpb.w3.eth.getBlock.assert_has_calls([
            call("latest"),
            call(8123121),
            call(8123122),
            call(8123123),
            call("latest")
        ],
                                             any_order=False)
        self.assertEqual(fpb.w3.eth.getBlock.call_count, 5,
                         "getBlock must have been called exactly 5 times")

        # Assert that the gateway was asked for the latest synced block
        fpb.get_start_block.assert_called_once()

        # Assert that blocks where POST'ed to the gateway as we expected
        fpb.requests.post.assert_has_calls([
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 8123121,
                     'blockHash': '0x3121',
                     'blockParentHash': '0x3120',
                     "syncTimestamp": CURRENT_TIME
                 }),
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 8123122,
                     'blockHash': '0x3122',
                     'blockParentHash': '0x3121',
                     "syncTimestamp": CURRENT_TIME
                 }),
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 8123123,
                     'blockHash': '0x3123',
                     'blockParentHash': '0x3122',
                     "syncTimestamp": CURRENT_TIME
                 })
        ],
                                           any_order=False)
        self.assertEqual(fpb.requests.post.call_count, 3,
                         "requests.post must have been called exactly 3 times")

    def test_start_block_or_fork_handling(self):
        # Mock the current time, so we get predictable results
        CURRENT_TIME = 123123123
        fpb.current_milli_time = Mock(return_value=CURRENT_TIME)

        # Mock w3.eth.getBlock() to control its behavior
        latest_call_counter = 0

        def getBlock_return_values(arg):
            blocks = {
                "latest":
                    AttributeDict({
                        "number": 124,
                        "hash": HexBytes("0xb124"),
                        "parentHash": HexBytes("0xb123")
                    }),
                124:
                    AttributeDict({
                        "number": 124,
                        "hash": HexBytes("0xb124"),
                        "parentHash": HexBytes("0xb123")
                    }),
                123:
                    AttributeDict({
                        "number": 123,
                        "hash": HexBytes("0xb123"),
                        "parentHash": HexBytes("0xb122")
                    }),
                122:
                    AttributeDict({
                        "number": 122,
                        "hash": HexBytes("0xb122"),
                        "parentHash": HexBytes("0xb121")
                    }),
                121:
                    AttributeDict({
                        "number": 121,
                        "hash": HexBytes("0xb121"),
                        "parentHash": HexBytes("0xb120")
                    }),
                120:
                    AttributeDict({
                        "number": 120,
                        "hash": HexBytes("0xa120"),
                        "parentHash": HexBytes("0xa119")
                    }),
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

        fpb.w3.eth.getBlock = Mock(side_effect=getBlock_return_values)

        # Mock requests.post to control responses and inspect POST calls made to the gateway
        fpb.requests.post = Mock(
            return_value=AttributeDict({"status_code": 200}))

        # Mock requests.get to control responses and inspect GET calls made to the gateway
        def request_get_handler(args):
            response = Mock()
            response.status_code = 200
            data = {
                fpb.BENEATH_GET_LATEST_BLOCK_URL: {
                    "blockNumber": 123,
                    "blockHash": "0xa123",
                    "blockParentHash": "0xa122",
                    "syncTimestamp": 1500000003
                },
                "https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number=122":
                    {
                        "blockNumber": 122,
                        "blockHash": "0xa122",
                        "blockParentHash": "0xa121",
                        "syncTimestamp": 1500000002
                    },
                "https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number=121":
                    {
                        "blockNumber": 121,
                        "blockHash": "0xa121",
                        "blockParentHash": "0xa120",
                        "syncTimestamp": 1500000001
                    },
                "https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number=120":
                    {
                        "blockNumber": 120,
                        "blockHash": "0xa120",
                        "blockParentHash": "0xa119",
                        "syncTimestamp": 1500000000
                    }
            }
            response.json.return_value = data[args]
            return response

        fpb.requests.get = Mock(side_effect=request_get_handler)

        # Run the main loop until the test-exception stops it
        try:
            fpb.main()
        except Exception as e:
            print(str(e))
            pass

        # Assert that w3.eth.getBlock(...) was called exactly as we expected,
        # both to verify gateway block hashes and to get new blocks to POST.
        fpb.w3.eth.getBlock.assert_has_calls([
            call(123),
            call(122),
            call(121),
            call(120),
            call('latest'),
            call(121),
            call(122),
            call(123),
            call(124),
            call('latest')
        ],
                                             any_order=False)
        self.assertEqual(fpb.w3.eth.getBlock.call_count, 10,
                         "getBlock must have been called exactly 10 times")

        # Assert that the gateway was called to compare its blocks hashes with current block hashes from web3
        fpb.requests.get.assert_has_calls([
            call(
                'https://beneath.network/projects/ethereum/streams/ethereum-blocks?get-latest-record'
            ),
            call(
                'https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number=122'
            ),
            call(
                'https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number=121'
            ),
            call(
                'https://beneath.network/projects/ethereum/streams/ethereum-blocks?=block-number=120'
            )
        ])
        self.assertEqual(fpb.requests.get.call_count, 4,
                         "requests.get must have been called exactly 4 times")

        # Assert that the new blocks gets POST'ed to the gateway
        fpb.requests.post.assert_has_calls([
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 121,
                     'blockHash': '0xb121',
                     'blockParentHash': '0xb120',
                     'syncTimestamp': 123123123
                 }),
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 122,
                     'blockHash': '0xb122',
                     'blockParentHash': '0xb121',
                     'syncTimestamp': 123123123
                 }),
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 123,
                     'blockHash': '0xb123',
                     'blockParentHash': '0xb122',
                     'syncTimestamp': 123123123
                 }),
            call(fpb.BENEATH_POST_BLOCK_URL,
                 headers={'content-type': 'application/json'},
                 json={
                     'blockNumber': 124,
                     'blockHash': '0xb124',
                     'blockParentHash': '0xb123',
                     'syncTimestamp': 123123123
                 })
        ])
        self.assertEqual(fpb.requests.post.call_count, 4,
                         "requests.post must have been called exactly 4 times")


if __name__ == '__main__':
    unittest.main()
