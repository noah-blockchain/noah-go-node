package types

// ChainID is ID of the network (1 - mainnet, 2 - testnet)
type ChainID byte

const (
	// ChainMainnet is mainnet chain ID of the network
	ChainMainnet ChainID = 0x01
	// ChainTestnet is mainnet chain ID of the network
	ChainTestnet ChainID = 0x02
)

// CurrentChainID is current ChainID of the network
var CurrentChainID = ChainMainnet

var (
	coinTestnet = StrToCoinSymbol("NOAH")
	coinMainnet = StrToCoinSymbol("NOAH")
)

// GetBaseCoin returns the coin symbol of the current ChainID
func GetBaseCoin() CoinSymbol {
	return getBaseCoin(CurrentChainID)
}

// GetBaseCoinID returns ID of base coin
func GetBaseCoinID() CoinID {
	return BasecoinID
}

func getBaseCoin(chainID ChainID) CoinSymbol {
	switch chainID {
	case ChainMainnet:
		return coinMainnet
	case ChainTestnet:
		return coinTestnet
	}

	panic("Unknown chain id")
}
