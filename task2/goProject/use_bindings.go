package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	counter "go-homework5/task2/foundryProject/bindings"
)

func main() {
	// ==================== 1. 环境准备 ====================

	// 获取 Infura API Key
	apiKey := os.Getenv("INFURA_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 INFURA_API_KEY 环境变量")
	}

	// 获取私钥
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("请设置 PRIVATE_KEY 环境变量")
	}

	// 连接到 Sepolia 测试网络
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, "https://sepolia.infura.io/v3/"+apiKey)
	if err != nil {
		log.Fatalf("无法连接到以太坊节点: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到 Sepolia 测试网络")

	// ==================== 2. 准备交易授权对象 ====================

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

	fmt.Printf("📝 发送方地址: %s\n", fromAddress.Hex())

	// 查询账户余额
	balance, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Fatalf("查询余额失败: %v", err)
	}
	fmt.Printf("💰 账户余额: %s ETH\n", weiToEth(balance))

	// 获取链 ID（Sepolia = 11155111）
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}
	fmt.Printf("🔗 链 ID: %s\n", chainID.String())

	// 创建交易授权对象
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("创建交易授权失败: %v", err)
	}

	// 设置 Gas 参数（可选，不设置会自动估算）
	// auth.GasLimit = uint64(300000)
	// auth.GasPrice = big.NewInt(20000000000) // 20 Gwei

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("开始部署智能合约...")
	fmt.Println(strings.Repeat("=", 50))

	// ==================== 3. 部署合约 ====================

	address, tx, _, err := counter.DeployCounter(auth, client)
	if err != nil {
		log.Fatalf("部署合约失败: %v", err)
	}

	fmt.Printf("📤 合约部署交易已发送\n")
	fmt.Printf("   交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("   合约地址: %s\n", address.Hex())

	// 等待交易确认
	fmt.Println("\n⏳ 等待交易确认...")
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		log.Fatalf("等待交易确认失败: %v", err)
	}

	fmt.Printf("✅ 合约部署成功！\n")
	fmt.Printf("   区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("   Gas 使用量: %d\n", receipt.GasUsed)
	fmt.Printf("   合约地址: %s\n", address.Hex())
	fmt.Printf("   区块浏览器: https://sepolia.etherscan.io/address/%s\n", address.Hex())

	// ==================== 4. 使用合约 - 读取数据 ====================

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("调用合约方法 - 读取数据")
	fmt.Println(strings.Repeat("=", 50))

	// 创建合约实例（用于后续交互）
	counterInstance, err := counter.NewCounter(address, client)
	if err != nil {
		log.Fatalf("创建合约实例失败: %v", err)
	}

	// 调用 number() 方法读取当前计数值
	count, err := counterInstance.Number(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("调用 number() 失败: %v", err)
	}
	fmt.Printf("📊 初始计数值: %s\n", count.String())

	// ==================== 5. 使用合约 - 写入数据 ====================

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("调用合约方法 - 写入数据")
	fmt.Println(strings.Repeat("=", 50))

	// 方法 1: 调用 increment() 增加计数
	fmt.Println("\n🔺 调用 increment() 方法...")
	txIncrement, err := counterInstance.Increment(auth)
	if err != nil {
		log.Fatalf("调用 increment() 失败: %v", err)
	}
	fmt.Printf("   交易哈希: %s\n", txIncrement.Hash().Hex())

	// 等待交易确认
	receiptIncrement, err := bind.WaitMined(ctx, client, txIncrement)
	if err != nil {
		log.Fatalf("等待 increment 交易确认失败: %v", err)
	}
	fmt.Printf("   ✅ increment 交易已确认 (区块: %d)\n", receiptIncrement.BlockNumber.Uint64())

	// 再次读取计数值
	countAfterIncrement, err := counterInstance.Number(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("调用 number() 失败: %v", err)
	}
	fmt.Printf("   📊 增加后的计数值: %s\n", countAfterIncrement.String())

	// 方法 2: 调用 setNumber() 设置特定值
	fmt.Println("\n🔧 调用 setNumber(100) 方法...")
	newNumber := big.NewInt(100)
	txSetNumber, err := counterInstance.SetNumber(auth, newNumber)
	if err != nil {
		log.Fatalf("调用 setNumber() 失败: %v", err)
	}
	fmt.Printf("   交易哈希: %s\n", txSetNumber.Hash().Hex())

	// 等待交易确认
	receiptSetNumber, err := bind.WaitMined(ctx, client, txSetNumber)
	if err != nil {
		log.Fatalf("等待 setNumber 交易确认失败: %v", err)
	}
	fmt.Printf("   ✅ setNumber 交易已确认 (区块: %d)\n", receiptSetNumber.BlockNumber.Uint64())

	// 再次读取计数值
	countAfterSet, err := counterInstance.Number(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("调用 number() 失败: %v", err)
	}
	fmt.Printf("   📊 设置后的计数值: %s\n", countAfterSet.String())

	// ==================== 6. 验证结果 ====================

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("验证结果")
	fmt.Println(strings.Repeat("=", 50))

	// 验证最终值是否符合预期
	expectedValue := big.NewInt(100)
	if countAfterSet.Cmp(expectedValue) == 0 {
		fmt.Println("✅ 验证通过：合约状态符合预期")
	} else {
		fmt.Printf("❌ 验证失败：期望值 %s，实际值 %s\n", expectedValue.String(), countAfterSet.String())
	}

	// 显示所有交易记录
	fmt.Println("\n📋 交易记录汇总:")
	fmt.Printf("   1. 部署合约: %s\n", tx.Hash().Hex())
	fmt.Printf("   2. Increment: %s\n", txIncrement.Hash().Hex())
	fmt.Printf("   3. SetNumber: %s\n", txSetNumber.Hash().Hex())

	fmt.Println("\n🎉 合约交互完成！")
	fmt.Printf("   在区块浏览器查看合约: https://sepolia.etherscan.io/address/%s\n", address.Hex())
}

// weiToEth 将 wei 转换为 ETH
func weiToEth(wei *big.Int) string {
	eth := new(big.Float)
	eth.SetString(wei.String())
	ethQuotient := new(big.Float).Quo(eth, big.NewFloat(1e18))
	return ethQuotient.Text('f', 18)
}
