# Known Addresses on the Ethereum Mainnet

- This table contains a manually curated and frequently updated list of known external and smart contract addresses on the Ethereum mainnet.
- Any address can only figure once in the dataset with one assigned category and description. That makes it easy to segment activity by known addresses.
- Where multiple labels apply to an address, we select the most appropriate.
- To suggest the addition of a new known address (or source of known addresses!), please [file an issue](https://gitlab.com/_beneath/beneath-models/issues) with evidence of the identity (e.g. a link or screenshot).

The table was deployed as a root table at [beneath.dev/beneath/ethereum/table:known-addresses](https://beneath.dev/beneath/ethereum/table:known-addresses). It was deployed with:

    beneath root-table create -f known-addresses.graphql --project beneath-ethereum --manual
