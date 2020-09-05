package mainnet

import (
	"github.com/gobuffalo/packr"
	"github.com/tendermint/tendermint/libs/os"
)

func GenerateStatic(genesisFile string) error {
	box := packr.NewBox("./noah-mainnet-1")
	bytes, err := box.Find("/genesis.json")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(genesisFile, bytes, 0644); err != nil {
		return err
	}

	return nil
}
