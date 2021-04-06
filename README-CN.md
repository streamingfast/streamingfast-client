[English Version](./README.md)

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
```

## 编程语言及访问

通过 [gRPC](https://grpc.io/) 提供对超过十几种编程语言的支持。

你需要两个 protobuf 定义文件：[bstream.proto](https://github.com/dfuse-io/proto/tree/develop/dfuse/bstream/v1) 以及 [codec.proto](https://github.com/dfuse-io/proto-ethereum/tree/develop/dfuse/ethereum/codec/v1)。

请参阅我们的[访问权验证文档](https://docs.dfuse.io/platform/dfuse-cloud/authentication/#obtaining-a-short-lived-jwt) 

可以拿此库中的 `main.go` 文件来做一些向导

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
