## Foundry

**Foundry is a blazing fast, portable and modular toolkit for Ethereum application development written in Rust.**

Foundry consists of:

- **Forge**: Ethereum testing framework (like Truffle, Hardhat and DappTools).
- **Cast**: Swiss army knife for interacting with EVM smart contracts, sending transactions and getting chain data.
- **Anvil**: Local Ethereum node, akin to Ganache, Hardhat Network.
- **Chisel**: Fast, utilitarian, and verbose solidity REPL.

## Documentation

https://book.getfoundry.sh/

## Usage

### Build

```shell
$ forge build
```

### Test

```shell
$ forge test
```

### Format

```shell
$ forge fmt
```

### Gas Snapshots

```shell
$ forge snapshot
```

### Anvil

```shell
$ anvil
```

### Deploy

```shell
$ forge script script/Counter.s.sol:CounterScript --rpc-url <your_rpc_url> --private-key <your_private_key>
```

### Cast

```shell
$ cast <subcommand>
```

### Help

```shell
$ forge --help
$ anvil --help
$ cast --help
```

---

## Go 绑定代码使用指南

本项目包含使用 `abigen` 生成的 Go 绑定代码，用于与 Counter 智能合约进行交互。

### 1. 生成绑定代码

如果尚未生成绑定代码，可以使用以下命令：

```bash
# 从 Solidity 源文件直接生成（推荐）
abigen --sol=src/Counter.sol --pkg=counter --out=bindings/Counter.go

# 或者从 ABI 和 Bin 文件生成
jq '.abi' out/Counter.sol/Counter.json > contract/Counter.abi
jq -r '.bytecode.object' out/Counter.sol/Counter.json > contract/Counter.bin
abigen --abi=contract/Counter.abi --bin=contract/Counter.bin --pkg=counter --out=bindings/Counter.go
```

### 2. 设置环境变量

在运行程序之前，需要设置以下环境变量：

```bash
# Linux/WSL/macOS
export INFURA_API_KEY="your_infura_api_key"
export PRIVATE_KEY="your_private_key_without_0x_prefix"

# Windows (CMD)
set INFURA_API_KEY=your_infura_api_key
set PRIVATE_KEY=your_private_key_without_0x_prefix

# Windows (PowerShell)
$env:INFURA_API_KEY="your_infura_api_key"
$env:PRIVATE_KEY="your_private_key_without_0x_prefix"
```

**获取 API Key 和私钥：**
- **Infura API Key**: 注册 [Infura](https://infura.io/) 账户并创建项目
- **Private Key**: 从 MetaMask 或其他钱包导出私钥（不要包含 `0x` 前缀）

### 3. 运行程序

```bash
# 进入 task2 目录
cd task2

# 运行 use_bindings.go
go run use_bindings.go
```

### 4. 程序功能

该程序演示了完整的智能合约交互流程：

1. **环境准备**
   - 连接到 Sepolia 测试网络
   - 验证账户余额
   - 创建交易授权对象

2. **部署合约**
   - 使用 `DeployCounter` 函数部署新的 Counter 合约
   - 等待交易确认
   - 输出合约地址和交易哈希

3. **读取数据**
   - 调用 `Number()` 方法读取当前计数值
   - 无需发送交易，只需读取链上状态

4. **写入数据**
   - 调用 `Increment()` 方法增加计数
   - 调用 `SetNumber(100)` 设置特定值
   - 等待每笔交易确认

5. **验证结果**
   - 验证最终计数值是否符合预期
   - 显示所有交易记录
   - 提供区块浏览器链接

### 5. 代码结构说明

```
task2/
├── bindings/
│   └── Counter.go          # abigen 生成的绑定代码
├── src/
│   └── Counter.sol         # Solidity 合约源码
├── out/
│   └── Counter.sol/
│       └── Counter.json    # 编译输出（包含 ABI 和 Bytecode）
├── contract/
│   ├── Counter.abi         # 提取的 ABI 文件
│   └── Counter.bin         # 提取的 Bytecode 文件
└── use_bindings.go         # 使用绑定代码的示例程序
```

### 6. 主要 API 说明

#### 部署合约
```go
address, tx, instance, err := counter.DeployCounter(auth, client)
```

#### 创建合约实例
```go
instance, err := counter.NewCounter(contractAddress, client)
```

#### 读取数据（只读调用）
```go
count, err := instance.Number(&bind.CallOpts{})
```

#### 写入数据（发送交易）
```go
// 增加计数
tx, err := instance.Increment(auth)

// 设置数值
tx, err := instance.SetNumber(auth, big.NewInt(100))
```

#### 等待交易确认
```go
receipt, err := bind.WaitMined(ctx, client, tx)
```

### 7. 注意事项

1. **Gas 费用**: 确保账户有足够的 Sepolia ETH 支付 Gas 费用
2. **网络超时**: 所有网络请求都设置了 15 秒超时
3. **交易确认**: 使用 `bind.WaitMined` 等待交易被打包到区块
4. **错误处理**: 所有操作都有完善的错误处理
5. **私钥安全**: 永远不要将私钥硬编码在代码中，始终使用环境变量

### 8. 常见问题

**Q: 提示 "insufficient funds"？**
A: 你的账户没有足够的 Sepolia ETH。可以从 [Sepolia Faucet](https://sepoliafaucet.com/) 获取测试币。

**Q: 交易长时间未确认？**
A: Sepolia 测试网有时会拥堵，可以尝试提高 Gas Price 或稍后重试。

**Q: 如何查看已部署的合约？**
A: 程序会输出合约地址，可以在 [Sepolia Etherscan](https://sepolia.etherscan.io/) 上查看。

**Q: 如何与已部署的合约交互？**
A: 修改代码，使用 `counter.NewCounter(existingAddress, client)` 创建已有合约的实例。
