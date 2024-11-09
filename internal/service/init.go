package service

import "github.com/ethereum/go-ethereum/ethclient"

const RPC_URL = "https://eth-mainnet.g.alchemy.com/v2/HRxbHqDu5RdRaVm5QGWGXZjAqVPBILKs"

func RetryDail() (*ethclient.Client, error) {
	var client *ethclient.Client
	for i := 0; i < 3; i++ {
		cli, err := ethclient.Dial(RPC_URL)
		if err == nil {
			client = cli
			break
		}
		if i == 2 {
			return nil, err
		}
	}
	return client, nil
}
