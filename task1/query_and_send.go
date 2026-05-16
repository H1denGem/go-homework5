package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 连接到 Sepolia 测试网络
	apiKey := os.Getenv("INFURA_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 INFURA_API_KEY 环境变量")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, "https://sepolia.infura.io/v3/"+apiKey)
	if err != nil {
		log.Fatalf("无法连接到以太坊节点: %v", err)
	}
	defer client.Close()

	fmt.Println("成功连接到 Sepolia 测试网络")

	// // 查询区块信息
	// blockNumber := uint64(5000000) // 可以修改为您想查询的区块号
	// queryBlock(client, blockNumber)

	// // 查询最新区块
	// queryLatestBlock(client)

	// 测试钱包地址
	to := common.HexToAddress("0x8D140087931f66257DE21faB78bD1c2B52d4e6d6")

	// 构建交易并发送
	sendTestTransaction(client, to)
}

// queryBlock 查询指定区块号的区块信息
func queryBlock(client *ethclient.Client, blockNumber uint64) {

	fmt.Printf("\n=== 查询区块 #%d ===\n", blockNumber)

	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	header, err := client.HeaderByNumber(reqCtx, big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Fatalf("查询区块失败: %v", err)
	}

	fmt.Printf("区块哈希: %s\n", header.Hash().Hex())
	fmt.Printf("父哈希: %s\n", header.ParentHash.Hex())
	fmt.Printf("区块号: %d\n", header.Number.Uint64())
	fmt.Printf("时间戳: %d\n", header.Time)
	fmt.Printf("难度: %s\n", header.Difficulty.String())
	fmt.Printf("Gas 限制: %d\n", header.GasLimit)
	fmt.Printf("Gas 已使用: %d\n", header.GasUsed)
	fmt.Printf("Base Fee: %s wei\n", header.BaseFee.String())
	fmt.Printf("交易数量: %d\n", len(header.TxHash))

	// 获取完整区块信息（包含交易）
	block, err := client.BlockByNumber(reqCtx, big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Fatalf("获取完整区块失败: %v", err)
	}
	fmt.Printf("实际交易数量: %d\n", len(block.Transactions()))
}

// queryLatestBlock 查询最新区块信息
func queryLatestBlock(client *ethclient.Client) {

	fmt.Println("\n=== 查询最新区块 ===")

	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	header, err := client.HeaderByNumber(reqCtx, nil)
	if err != nil {
		log.Fatalf("查询最新区块失败: %v", err)
	}

	fmt.Printf("区块哈希: %s\n", header.Hash().Hex())
	fmt.Printf("区块号: %d\n", header.Number.Uint64())
	fmt.Printf("时间戳: %d\n", header.Time)
	fmt.Printf("交易数量: %d\n", len(header.TxHash))
}

func sendTestTransaction(client *ethclient.Client, to common.Address) {

	// 获取发送方私钥（从环境变量）
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("请设置 PRIVATE_KEY 环境变量")
	}

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("无效的私钥: %v", err)
	}

	// 获取公钥和地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("无法转换为 ECDSA 公钥")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("发送方地址: %s\n", fromAddress.Hex())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 查询账户余额
	balance, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Fatalf("查询余额失败: %v", err)
	}
	fmt.Printf("账户余额: %s wei (%s ETH)\n", balance.String(), weiToEth(balance))

	// 接收方地址（请替换为您想要发送的地址）
	toAddress := to

	// 转账金额（单位：wei，这里是 0.001 ETH）
	value := big.NewInt(1000000000000000) // 0.001 ETH in wei

	// 获取链 ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}

	// 获取 nonce
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("获取 nonce 失败: %v", err)
	}

	// 获取建议的 gas 价格
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf("获取 gas 价格失败: %v", err)
	}

	// 估算 gas 限制
	msg := ethereum.CallMsg{
		From:  fromAddress,
		To:    &toAddress,
		Value: value,
	}
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		log.Fatalf("估算 gas 失败: %v", err)
	}

	// 创建交易
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败: %v", err)
	}

	// 发送交易
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	fmt.Printf("\n✅ 交易发送成功！\n")
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("发送方: %s\n", fromAddress.Hex())
	fmt.Printf("接收方: %s\n", toAddress.Hex())
	fmt.Printf("金额: %s ETH\n", weiToEth(value))
	fmt.Printf("Gas 限制: %d\n", gasLimit)
	fmt.Printf("Gas 价格: %s wei\n", gasPrice.String())
	fmt.Printf("\n请在区块浏览器查看: https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())
}

// weiToEth 将 wei 转换为 ETH
func weiToEth(wei *big.Int) string {
	eth := new(big.Float)
	eth.SetString(wei.String())
	ethQuotient := new(big.Float).Quo(eth, big.NewFloat(1e18))
	return ethQuotient.Text('f', 18)
}
