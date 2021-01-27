Streaming Fast on Ethereum
--------------------------

Stream Ethereum data like there's no tomorrow

https://streamingfast.io/


## Install

Get a release from [./releases](the releases pages)


## Usage

```bash
$ export STREAMINGFAST_API_KEY="server_......................"

# Stream UniswapV2 calls, for a single block and close
$ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" 11700000 11700001

# Stream the last thousand blocks and continue forever
$ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" -1000

# Stream from the last cursor, continue forever
$ sf --start-cursor "10928019832019283019283" "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']"
```

## Sample output

Here is a short peek into some output. Read [**a full sample output here**](./sample_output_block_11740433.json).

```jsonnet
{
  // Indicator that this is a new block, could be IRREVERSIBLE or UNDO
  "step": "STEP_NEW",

  // Use this to continue *exactly* where you left off, guaranteeing linearity of your streaming
  // processes.
  "cursor": "PWxQqpUKpA64sLwiUK9I7aWwLpcyB1toUQvhKRJLhY2goSHD1JryAGZ8YE-DmKukiRToGFOljdvOFix7-8ZWuIPrkr426CMxTy95woDt-73mefKhPFsfc-9hVuqJatLbUQ=="

  // Obtain block-level information, filtered to keep only the transactions
  // that matched your filtering criterias.
  "block": {
    "hash": "06fac697d0798bc82dd13c99c432b09664b1a6b5299dea9309911c5a42d39078",
    "number": "11740433",
    "header": {
      "parentHash": "d9e47617a4b60c6466c82f979b80046f765dc3ac341c0db113ba7428d6b192f8",
      "uncleHash": "1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
      "coinbase": "ea674fdde714fd979de3edf0f56aa9716b898ec8",
      "stateRoot": "9a35f4e8dd255e63d9fadc21796b7b9eb5e77deec8d3c553156ecd3bf8299ab9",
      "transactionsRoot": "c822a62db0ac0fee7b3ee5c5c774de6541050d0fe62a0ea8cd5389b0156e622a",
      "receiptRoot": "72a07569c369fc073b374a9f930e37fbf2aa52400f7d9b47c17375211a15b1c1",
      "logsBloom": "16ba440e4d1623d1e0ab05a28c019a82d0409011b86e59f04cf9580e12f755816b3c97242ea1122651acdbda300429d89b2964c21b42c8808759180bb07c3850551d2ceaa790987a5a71ca1c1149e1e97f5400362556f455151239489a473215dd4094562ef5c0042888c08813108fcacea8119167891496a6778851d71275ec8906c975df09a0309841316c4c806acbc8312f992172d0ea5122b16f255954e3279926464d1739f70492ffd8b599b0ea0a0100d5619ad89850b34c393c9662459cbb9d43292fc0000e00b60104de054cc31c8f0215cb10981240d876b34068a4901b6e0134e4d5ac7186364e0b97e3a5842d231e01cd32795ae8f85481920c21",
      "difficulty": "10489a91f1e9e1",
      "number": "11740433",
      "gasLimit": "12481378",
      "gasUsed": "12477353",
      "timestamp": "2021-01-27T21:58:38Z",
      "extraData": "65746865726d696e652d6575312d35",
      "mixHash": "744e2e38ed64d84bf393c8535478587eea7957cd12ec68a43b0deae14136c3c1",
      "nonce": "8789940137479893166",
      "hash": "06fac697d0798bc82dd13c99c432b09664b1a6b5299dea9309911c5a42d39078"
    },

    // Execution traces of matching transactions, along withj
    // LOTS of data and details.
    "transactionTraces": [
      {
        "hash": "3ef815c7b531ca2c9da824bc9daa322edbc1ae7ba99548f2498d7ecc4279b934",
        "to": "582b409cbf6c026a639f065641f910e9d2d7f482",
        "from": "ea674fdde714fd979de3edf0f56aa9716b898ec8",
        "nonce": "30654932",
        "gasPrice": "3b9aca00",       // Gas price, hex-encoded.
        "gasLimit": "50000",
        "value": "0164732a08272565",  // ETH value being transfered, hex-encoded.
        "v": "25", "r": "c355c........a515", "s": "4b04e........c314",
        "gasUsed": "21000",
        "receipt": {
          "cumulativeGasUsed": "21000",
          "logsBloom": "00000...."
        },
        "calls": [

          // This is where it gets interesting:

          {
            "index": 1,
            "callType": "CALL",
            "caller": "ac844b604d6c600fbe55c4383a6d87920b46a160",
            "address": "d9e1ce17f2641f24ae83637ab66a2cca9c378b9f",
            "value": "",              // This would be value == 0
            "gasLimit": "476928",
            "gasConsumed": "128974",

            // You get return data between each calls
            "returnData": "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000130feb5cd4e3ad00000000000000000000000000000000000000000000000000001bf019b3165faeed7",

            // You're also privy to the input that was given to each call in the callgraph.
            "input": "18cbafe5000000000000000000000000000000000000000000000130feb5cd4e3ad00000000000000000000000000000000000000000000000000001bec863b0fd5a700000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a160000000000000000000000000000000000000000000000000000000006011e28b00000000000000000000000000000000000000000000000000000000000000020000000000000000000000006b3595068778dd592e39a122f4f5a5cf09c90fe2000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
            "executedCode": true,

            //

            "balanceChanges": [
              {
                "address": "ac844b604d6c600fbe55c4383a6d87920b46a160",
                "oldValue": "1a85cdf3216e9f3c14",
                "newValue": "1a850c4f1f2f5abd34",
                "reason": "REASON_GAS_BUY"
              },
              {
                "address": "ac844b604d6c600fbe55c4383a6d87920b46a160",
                "oldValue": "1c440dea509555ac0b",
                "newValue": "1c449dbbd79f460b71",
                "reason": "REASON_GAS_REFUND"
              },

              // [... snip ...]

              {
                "address": "ea674fdde714fd979de3edf0f56aa9716b898ec8",
                "oldValue": "4fc1fd62ca11ec8278",
                "newValue": "4fc22f35454740a1f2",
                "reason": "REASON_REWARD_TRANSACTION_FEE"
              }
            ],

            // Self explanatory I guess, but pretty useful to avoid calling nodes all the time.

            "nonceChanges": [
              {
                "address": "ac844b604d6c600fbe55c4383a6d87920b46a160",
                "oldValue": "25310",
                "newValue": "25311"
              }
            ],
            "gasChanges": [
              {
                "oldValue": "500000",
                "newValue": "476928",
                // You've got all the reasons why gas is charged or refunded
                "reason": "REASON_INTRINSIC_GAS"
              },
              {
                // [... snip ...]
                "reason": "REASON_CALL_DATA_COPY"
              },
              {
                // [... snip ...]
                "reason": "REASON_REFUND_AFTER_EXECUTION"
              },

              // [... snip ...]

            ],

            // These allow you to rebuild a full graph of consumption and trace
            // gas consumption even between EVM calls
            "gasEvents": [
              {
                "id": "ID_BEFORE_CALL",
                "gas": "473893",
                "linkedCallIndex": "2"
              },
              {
                "id": "ID_AFTER_CALL",
                "gas": "471970",
                "linkedCallIndex": "2"
              },

              // [... snip ...]

            ]
          },

          // The following call is an internal transaction, a call initiated by the previous
          // one.

          {
            // 1-based index of the call within the transaction.
            "index": 2,

            // This is a 1-based index, where the value 0 would mean "no parent", or top-level
            "parentIndex": 1,

            // Can easily represent the call tree with depth
            "depth": 1,

            // This is a static call, mostly to get data out of the other contract.
            "callType": "STATIC",
            "caller": "d9e1ce17f2641f24ae83637ab66a2cca9c378b9f",
            "address": "795065dcc9f64b5614c407a6efdc400da6221fb0",
            "value": "",
            "gasLimit": "465794",
            "gasConsumed": "1217",

            // Also useful to debug or understand what's happening
            "returnData": "0000000000000000000000000000000000000000000f2484783460e8ff44c235000000000000000000000000000000000000000000001644567a439aadc9a3ea000000000000000000000000000000000000000000000000000000006011e1b4",
            "input": "0902f1ac",
            "executedCode": true
          },

          // Seems we made a third call here:

          {
            "index": 3,
            "parentIndex": 1,
            "depth": 1,
            "callType": "CALL",

            [... snip ...]

            // You can use this data to reverse the changes to the storage below
            // and understand which Externally Owned Address's balance is being
            // modified.. Avoid costly, and difficult to sync `getBalance()`
            // eth_calls, by using the State directly (see below)

            "keccakPreimages": {
              "0490a33f730091720d4c0d29bd2bf6a18ca8c44a423a64002ff24115dd8b8381": "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a1600000000000000000000000000000000000000000000000000000000000000001",

              // Take note of this one, and read below:

              "a60c07f2aed92cf0e2ca94448542cb8f5cc91bf932d411877ec1850bf66a155f": "000000000000000000000000795065dcc9f64b5614c407a6efdc400da6221fb00000000000000000000000000000000000000000000000000000000000000000",
              "b995e795ef3cc8a80fd42d092cd13c326fe2a42885cee30593e77e4b404db0e3": "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a1600000000000000000000000000000000000000000000000000000000000000000",
              "d7e11f80431dbe8f8fe4aba8c8c50b6b80718ea5764ac29e9b9b6e5b537bc944": "000000000000000000000000d9e1ce17f2641f24ae83637ab66a2cca9c378b9f0490a33f730091720d4c0d29bd2bf6a18ca8c44a423a64002ff24115dd8b8381"
            },

            // These are the *actual* values that are modified by the contract, with their balances.
            // What you usually only found on Etherscan can now be used by your algorithms
            // and apps, to speed up things and make things more consistent.

            "storageChanges": [
              {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",
                "key": "b995e795ef3cc8a80fd42d092cd13c326fe2a42885cee30593e77e4b404db0e3",
                "oldValue": "00000000000000000000000000000000000000000000062a350284837151aee9",
                "newValue": "0000000000000000000000000000000000000000000004f9364cb7353681aee9"
              },
              {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",

                // Noticed that the `keccakPreimages` above has data associated with the key
                // that follows (a60c07...)?
                // It happens to hold the address of the user's balance we're changing here:
                //   0x795065dcc9f64b5614c407a6efdc400da6221fb
                // Without the keccakPreimages, those state changes are admittedly pretty
                // opaque :)

                "key": "a60c07f2aed92cf0e2ca94448542cb8f5cc91bf932d411877ec1850bf66a155f",
                "oldValue": "0000000000000000000000000000000000000000000f2484783460e8ff44c235",
                "newValue": "0000000000000000000000000000000000000000000f25b576ea2e373a14c235"
              },
              {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",
                "key": "d7e11f80431dbe8f8fe4aba8c8c50b6b80718ea5764ac29e9b9b6e5b537bc944",
                "oldValue": "ffffffffffffffffffffffffffffffffffffffffffbe627c53c41b10f256ffff",
                "newValue": "ffffffffffffffffffffffffffffffffffffffffffbe614b550e4dc2b786ffff"
              }
            ],
            "logs": [
             {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",
                "topics": [

                  // Yes, that's an ERC-20 transfer

                  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",

                  // From address

                  "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a160",

                  // To address
                  "000000000000000000000000795065dcc9f64b5614c407a6efdc400da6221fb0"
                ],

                // And the amount is in here:
                "data": "000000000000000000000000000000000000000000000130feb5cd4e3ad00000",

                // This is the index of this log event within the whole block.

                "blockIndex": 5
              },
              {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",
                "topics": [
                  "8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
                  "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a160",
                  "000000000000000000000000d9e1ce17f2641f24ae83637ab66a2cca9c378b9f"
                ],
                "data": "ffffffffffffffffffffffffffffffffffffffffffbe614b550e4dc2b786ffff",
                "index": 1,
                "blockIndex": 6
              }
            ]
          }
        ]
      }
    ],

    // These are transaction-level balance changes (not caused by an EVM Call):

    "balanceChanges": [
      {
        "address": "ea674fdde714fd979de3edf0f56aa9716b898ec8",
        "oldValue": "4fcdc5db9057ea8285",
        "newValue": "4fe98748f7a6b28285",
        "reason": "REASON_REWARD_MINE_BLOCK"
      }
    ]
  }
}


```


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
