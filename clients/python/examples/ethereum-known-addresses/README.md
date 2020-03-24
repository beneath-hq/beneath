# Known Addresses on the Ethereum Mainnet

- This stream contains a manually curated and frequently updated list of known external and smart contract addresses on the Ethereum mainnet.
- Any address can only figure once in the dataset with one assigned category and description. That makes it easy to segment activity by known addresses.
- Where multiple labels apply to an address, we select the most appropriate. 
- To suggest the addition of a new known address (or source of known addresses!), please [file an issue](https://gitlab.com/_beneath/beneath-models/issues) with evidence of the identity (e.g. a link or screenshot).

The stream was deployed as a root stream at [beneath.dev/projects/beneath-ethereum/streams/known-addresses](https://beneath.dev/projects/beneath-ethereum/streams/known-addresses). It was deployed with:

    beneath root-stream create -f known-addresses.graphql --project beneath-ethereum --manual
