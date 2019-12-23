package types

import "github.com/noah-blockchain/noah-go-node/config"

type ChainID byte

const (
	ChainTestnet = 0x02
	ChainMainnet = 0x01
)

func GetCurrentChainID() ChainID {
	if config.NetworkId == "noah-mainnet-1" {
		return ChainMainnet
	}
	return ChainTestnet
}

func GetBaseCoin() CoinSymbol {
	return getBaseCoin(GetCurrentChainID())
}

func getBaseCoin(chainID ChainID) CoinSymbol {
	var coin CoinSymbol

	switch chainID {
	case ChainMainnet:
		copy(coin[:], []byte("NOAH"))
	case ChainTestnet:
		copy(coin[:], []byte("NOAH"))
	}

	coin[4] = byte(0)

	return coin
}
