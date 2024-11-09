package service

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log/slog"
	"math/big"
)

const ADDRESS1 = "0x0604cc2a4d90d0854d4551133c31d6c55232c749"
const ADDRESS2 = "0xc1e4400506b6178ff92ed8a353e996a3227ed877"

func (serv *Serv) EstimateGas() {
	client, err := RetryDail()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 估算转账 gas
	address1 := common.HexToAddress(ADDRESS1)
	address2 := common.HexToAddress(ADDRESS2)
	callMsg := newCallMsg(address1, &address2)

	//gasFee := gas * (callMsg.GasFeeCap + callMsg.GasTipCap)
	baseGas, err := getCurrentBaseFee(client)
	if err != nil {
		panic(err)
	}
	if callMsg.GasFeeCap.Cmp(baseGas) == -1 {
		callMsg.GasFeeCap = baseGas
	}
	gas, err := client.EstimateGas(context.Background(), callMsg)
	if err != nil {
		panic(err)
	}
	gasFee1 := new(big.Int).Mul(new(big.Int).SetUint64(gas), new(big.Int).Add(callMsg.GasFeeCap, callMsg.GasTipCap))
	fmt.Println("gas:", gas)
	fmt.Println("gasFee:", gasFee1)
	slog.Info(fmt.Sprintf("Log_gasFee: %i", gasFee1), "tesdResult", true)
}

func newCallMsg(from common.Address, to *common.Address) ethereum.CallMsg {
	return ethereum.CallMsg{
		From:      from,
		To:        to,
		Gas:       20000,                                                         // 标准转账 gas 限制
		GasFeeCap: new(big.Int).SetUint64(7000000000),                            // 30 Gwei, EIP-1559 最大总费用
		GasTipCap: new(big.Int).SetUint64(1500000000),                            // 1.5 Gwei, EIP-1559 小费
		Value:     new(big.Int).Mul(big.NewInt(2000000000000000), big.NewInt(1)), // 0.02 ETH
		Data:      []byte{},                                                      // 简单转账不需要数据
	}
}

func getCurrentBaseFee(client *ethclient.Client) (*big.Int, error) {
	ctx := context.Background()

	// 获取最新的区块头
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %v", err)
	}

	// 检查 BaseFee 是否存在（在某些非 EIP-1559 兼容的网络上可能不存在）
	if header.BaseFee == nil {
		return nil, fmt.Errorf("base fee not available (non EIP-1559 network?)")
	}

	return header.BaseFee, nil
}
