[Englihs Version](./README.md)

Streaming Fast on Ethereum
--------------------------

获取以太坊极速数据流

https://streamingfast.io/

![StreamingFast Demo](https://streamingfast.io/streamingfast.gif)

**觉得 StreamingFast 很好用吗?点亮源码库的星标！**


## 上手指南

1. 在 https://streamingfast.io 上获取 API KEY
2. 下载 [最新版本](https://github.com/streamingfast/streamingfast-client/releases)
3. 连接数据流！


## 安装

从源码上开始构建：

    go get -v github.com/streamingfast/streamingfast-client/cmd/sf

或下载静态链接[二进制程序，支持 Windows、macOS or Linux](https://github.com/streamingfast/streamingfast-client/releases)。


## 使用

```bash
$ export STREAMINGFAST_API_KEY="server_......................"

# 查看所有对 UniswapV2 路由的调用，查询区间为一个区块
$ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" 11700000 11700001

# 查看所有对 UniswapV2 路由的调用，查询区间为前100个区块，并继续实时监听
$ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" -100

# 从上次的断电续连，利用最新的 cursor 定位，获取所有分叉的提示（UNDO, IRREVERSIBLE）并继续实时监听
$ sf --handle-forks --start-cursor "10928019832019283019283" "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']"

# 查看币安智能链（BSC）上一个指定区间的所有区块 
$ sf --bsc "true" 100000 100002

# 查看 Polygon 上一个指定区间的所有区块
$ sf --polygon "true" 100000 100002

# 查看 HECO 上一个指定区间的所有区块
$ sf --heco "true" 100000 100002

# 查看 Fantom Opera 主网上一个指定区间的所有区块
$ sf --fantom "true" -5
```

## 编程语言及访问

通过 [gRPC](https://grpc.io/) 提供对超过十几种编程语言的支持。

你需要两个 protobuf 定义文件：[bstream.proto](https://github.com/dfuse-io/proto/tree/develop/dfuse/bstream/v1) 以及 [codec.proto](https://github.com/dfuse-io/proto-ethereum/tree/develop/dfuse/ethereum/codec/v1)。

请参阅我们的[访问权验证文档](https://docs.dfuse.io/platform/dfuse-cloud/authentication/#obtaining-a-short-lived-jwt) 

可以拿此库中的 `main.go` 文件来做一些向导


## 输出示例

这是 StreamingFast 输出的简要节选，你也可以查看[**完整的输出示例**](./sample_output_block_11740433.json).

```jsonnet
{
  // 示意这是一个新区块，可能是 IRREVERSIBLE（不可逆）或 UNDO（回滚）
  "step": "STEP_NEW",

  // 使用此字段可以*精确地*从断点续传，确保无缝的直线性流式传输
  "cursor": "PWxQqpUKpA64sLwiUK9I7aWwLpcyB1toUQvhKRJLhY2goSHD1JryAGZ8YE-DmKukiRToGFOljdvOFix7-8ZWuIPrkr426CMxTy95woDt-73mefKhPFsfc-9hVuqJatLbUQ=="

  // 获取区块级信息，按照你所设定的条件对交易进行过滤
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

    // 过滤出的交易的执行痕迹，还有丰富的数据和细节信息
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

          // 到这就开始更有意思了

          {
            "index": 1,
            "callType": "CALL",
            "caller": "ac844b604d6c600fbe55c4383a6d87920b46a160",
            "address": "d9e1ce17f2641f24ae83637ab66a2cca9c378b9f",
            "value": "",              // This would be value == 0
            "gasLimit": "476928",
            "gasConsumed": "128974",

            // 你会在每个调用间都获取到反馈数据
            "returnData": "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000130feb5cd4e3ad00000000000000000000000000000000000000000000000000001bf019b3165faeed7",

            // 还能获取到调用表中每个调用的输入
            "input": "18cbafe5000000000000000000000000000000000000000000000130feb5cd4e3ad00000000000000000000000000000000000000000000000000001bec863b0fd5a700000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a160000000000000000000000000000000000000000000000000000000006011e28b00000000000000000000000000000000000000000000000000000000000000020000000000000000000000006b3595068778dd592e39a122f4f5a5cf09c90fe2000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
            "executedCode": true,

            // 可以看到所有内部 ETH 转账，以及转账前后的余额，这个功能只有 StreamingFast 才提供

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

            // 很明显，避免轮询节点还是非常有用的

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

            // 这个让你可以重构完整的 gas 消耗量和痕迹图表，甚至能看到两个 EVM 调用之间消耗的 gas
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

          // 下面你看到的这个调用是个内部交易，是由上一个调用发起的

          {
            // 对交易中调用基于1的索引
            "index": 2,

            // 在基于1的索引中，0 代表没有母操作，或者说是在最顶层的
            "parentIndex": 1,

            // 可以轻松但深度的理解调用间的顺序
            "depth": 1,

            // 这是一个静态调用，主要是为了从其他合约中获取数据
            "callType": "STATIC",
            "caller": "d9e1ce17f2641f24ae83637ab66a2cca9c378b9f",
            "address": "795065dcc9f64b5614c407a6efdc400da6221fb0",
            "value": "",
            "gasLimit": "465794",
            "gasConsumed": "1217",

            // 可以用来 debug 或了解正在发生的情况
            "returnData": "0000000000000000000000000000000000000000000f2484783460e8ff44c235000000000000000000000000000000000000000000001644567a439aadc9a3ea000000000000000000000000000000000000000000000000000000006011e1b4",
            "input": "0902f1ac",
            "executedCode": true
          },

          // 看来这还触发了第三个调用：

          {
            "index": 3,
            "parentIndex": 1,
            "depth": 1,
            "callType": "CALL",

            [... snip ...]

            "logs": [
             {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",
                "topics": [

                  // 这是一个 ERC-20 转账

                  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",

                  // 来自地址：

                  "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a160",

                  // 接收地址：
                  "000000000000000000000000795065dcc9f64b5614c407a6efdc400da6221fb0"
                ],

                // 转账数量:
                "data": "000000000000000000000000000000000000000000000130feb5cd4e3ad00000",

                // 这是对此日志事件在整个区块中的一个索引

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

            // 下面这个数据是非常实用的：直接获取 ERC-20 账号更新前后的余额，省去你之后去做好几百个 `getBalance()` 调用。

            "erc20BalanceChanges": [
              {
                "holderAddress": "164e948cb069f2008bda69d89b5bbdc0639f6783",
                "oldBalance": "0000000000000000000000000000000000000000028427eec9781a19c7b19ac0",
                "newBalance": "0000000000000000000000000000000000000000028427ee6162bfd660439ac0"
              },
              {
                "holderAddress": "703052a1ef835dd5842190e53896672b8f9249f1",
                "oldBalance": "00000000000000000000000000000000000000000000000068155a43676e0000",
                "newBalance": "000000000000000000000000000000000000000000000000d02ab486cedc0000"
              }
            ],
            "erc20TransferEvents": [
              {
                "from": "164e948cb069f2008bda69d89b5bbdc0639f6783",
                "to": "703052a1ef835dd5842190e53896672b8f9249f1",
                "amount": "00000000000000000000000000000000000000000000000068155a43676e0000"
              }
            ]

            // 你可以利用这个数据来撤消对下面的存储的更改
            // 并了解是哪个外部地址的余额在被更改，省去成本高、
            // 很难同步的 `getBalance()` eth_calls，直接利用状态信息

            "keccakPreimages": {
              "0490a33f730091720d4c0d29bd2bf6a18ca8c44a423a64002ff24115dd8b8381": "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a1600000000000000000000000000000000000000000000000000000000000000001",

              // 注意下这个数据，以及它下面的解释：

              "a60c07f2aed92cf0e2ca94448542cb8f5cc91bf932d411877ec1850bf66a155f": "000000000000000000000000795065dcc9f64b5614c407a6efdc400da6221fb00000000000000000000000000000000000000000000000000000000000000000",
              "b995e795ef3cc8a80fd42d092cd13c326fe2a42885cee30593e77e4b404db0e3": "000000000000000000000000ac844b604d6c600fbe55c4383a6d87920b46a1600000000000000000000000000000000000000000000000000000000000000000",
              "d7e11f80431dbe8f8fe4aba8c8c50b6b80718ea5764ac29e9b9b6e5b537bc944": "000000000000000000000000d9e1ce17f2641f24ae83637ab66a2cca9c378b9f0490a33f730091720d4c0d29bd2bf6a18ca8c44a423a64002ff24115dd8b8381"
            },

            // 这些是*真正*被合约更改了的值，和它们的余额
            // 通常只能在 Etherscan 找到的数据，现在你也能在你的算法或应用里用上了
            // 让你的程序更快、同步性更强

            "storageChanges": [
              {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",
                "key": "b995e795ef3cc8a80fd42d092cd13c326fe2a42885cee30593e77e4b404db0e3",
                "oldValue": "00000000000000000000000000000000000000000000062a350284837151aee9",
                "newValue": "0000000000000000000000000000000000000000000004f9364cb7353681aee9"
              },
              {
                "address": "6b3595068778dd592e39a122f4f5a5cf09c90fe2",

                // 你有没有注意到上面的 `keccakPreimages` 是有跟随 (a60c07...) 这个私钥的相关的数据的？
                // 它恰好有我们更改了余额的地址：
                //   0x795065dcc9f64b5614c407a6efdc400da6221fb
                // 如果你看不到 keccakPreimages 的数据，这些状态变化是很不透明的

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
            ]
          }
        ]
      }
    ],

    // 下面这是在交易层级的余额变化（非 EVM 调用所出发的变化）：
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


## 查询语言

这里的过滤查询所用到的语言是一个 _Common Expression
Language_ ，可以在这里看到它的定义:

https://github.com/google/cel-spec/blob/master/doc/langdef.md

查询发出后是会去对每个调用进行匹配的（对_内部交易_是单独匹配的），只要一个
交易中有一个与查询条件匹配的调用，这个交易就会被反馈给你

用于过滤的字段：

* **`from`**: _string_，调用的签署者或发起者
* **`to`**: _string_，目标合约/调用的地址
* **`nonce`**: _number_，被执行的操作的名字
* **`input`**: _string_，以"0x"为前缀的十六进制输入；如果输入是空的，则字符串也是空的
* **`gas_price_gwei`**: _number_，交易的 gas 价格，单位为 GWEI
* **`gas_limit`**: _number_，gas 上限，以计算单元为单位
* **`erc20_from`**: _string_，一个 ERC20 转账的 `from` 字段；如果非 ERC20 转账，则字符串为空
* **`erc20_to`**: _string_，转账的 `to` 字段；如果非 ERC20 转账，则字符串为空

**注意**: 十六进制字符的所有字符串比较均要求英文“小写”表示，请确保在发送查询之前对其进行规范

## 查询示例

```
to == '0x7a250d5630b4cf539739df2c5dacb4c659f2488d'
```




## 证书

[Apache 2.0](./LICENSE)
