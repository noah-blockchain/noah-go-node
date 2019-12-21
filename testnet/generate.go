package testnet

import (
	"github.com/gobuffalo/packr"
	"github.com/tendermint/tendermint/libs/common"
)

func GenerateStatic(genesisFile string) error {
	box := packr.NewBox("./noah-testnet-1")
	bytes, err := box.Find("/genesis.json")
	if err != nil {
		panic(err)
	}

	if err := common.WriteFile(genesisFile, bytes, 0644); err != nil {
		return err
	}

	return nil
}
