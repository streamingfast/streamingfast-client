[查看中文版](./README-CN.md)

Streaming Fast on Ethereum
--------------------------

Stream Ethereum data like there's no tomorrow

https://streamingfast.io/

![StreamingFast Demo](https://streamingfast.io/streamingfast.gif)

**Impressed? Star the repo!**


## Getting started

1. Get an API key from https://streamingfast.io
1. Download a release from [the releases](https://github.com/streamingfast/streamingfast-client/releases)
1. Get streaming!


## Install

Build from source:

    go get -v github.com/streamingfast/streamingfast-client/cmd/sf

or download a statically linked [binary for Windows, macOS or Linux](https://github.com/streamingfast/streamingfast-client/releases).


## Usage

```bash
$ export STREAMINGFAST_API_KEY="server_......................"

# Watch all calls to the UniswapV2 Router, for a single block and close
$ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" 11700000 11700001

# Watch all calls to the UniswapV2 Router, include the last 100 blocks, and stream forever
$ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" -100

# Continue where you left off, start from the last known cursor, get all fork notifications (UNDO, IRREVERSIBLE), stream forever
$ sf --handle-forks --start-cursor "10928019832019283019283" "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']"

```

## Programmatic access

Access is done through [gRPC](https://grpc.io/), supporting more than a dozen popular languages.

You will need those two protobuf definition files: [bstream.proto](https://github.com/dfuse-io/proto/tree/develop/dfuse/bstream/v1) and [codec.proto](https://github.com/dfuse-io/proto-ethereum/tree/develop/dfuse/ethereum/codec/v1).

Refer to the [Authentication](https://docs.dfuse.io/platform/dfuse-cloud/authentication/#obtaining-a-short-lived-jwt) section of our docs for details.

Take inspiration from the `main.go` file in this repository.

## Query language

The language used as the search query is a _Common Expression
Language_ expression, as defined here:

https://github.com/google/cel-spec/blob/master/doc/langdef.md

Queries match on individual CALLs (so called _internal transactions_
are matched individually), and any transaction with at least one
matching Call will be returned.

Fields available for filtering:

* **`from`**: _string_, signer or originator of the Call
* **`to`**: _string_, target contract or address of the Call
* **`nonce`**: _number_, the name of the action being executed
* **`input`**: _string_, "0x"-prefixed hex of the input; the string will be empty if input is empty.
* **`gas_price_gwei`**: _number_, gas price for the transaction, in GWEI.
* **`gas_limit`**: _number_, gas limit, in units of computation.
* **`erc20_from`**: _string_, the `from` field of an ERC20 Transfer; string empty when not an ERC20 Transfer.
* **`erc20_to`**: _string_, the `to` field of an ERC20 Transfer; string empty when not an ERC20 Transfer.

**NOTE**: all string comparisons of hex characters are in **lower case**, so make sure to normalize your query before sending it.



## Sample queries

```
to == '0x7a250d5630b4cf539739df2c5dacb4c659f2488d'
```




## License

[Apache 2.0](./LICENSE)
