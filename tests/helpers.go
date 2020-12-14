package tests

import (
	"crypto/ecdsa"
	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	"github.com/noah-blockchain/noah-go-node/config"
	"github.com/noah-blockchain/noah-go-node/core/noah"
	"github.com/noah-blockchain/noah-go-node/core/transaction"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/crypto"
	"github.com/noah-blockchain/noah-go-node/rlp"
	"github.com/tendermint/go-amino"
	tmTypes "github.com/tendermint/tendermint/abci/types"
	"os"
	"path/filepath"
	"time"
)

// CreateApp creates and returns new Blockchain instance
// Recreates $HOME/.noah_test dir
func CreateApp(state types.AppState) *noah.Blockchain {
	utils.NoahHome = os.ExpandEnv(filepath.Join("$HOME", ".noah_test"))
	_ = os.RemoveAll(utils.NoahHome)

	jsonState, err := amino.MarshalJSON(state)
	if err != nil {
		panic(err)
	}

	cfg := config.GetConfig()
	app := noah.NewNoahBlockchain(cfg)
	app.InitChain(tmTypes.RequestInitChain{
		Time:    time.Now(),
		ChainId: "test",
		Validators: []tmTypes.ValidatorUpdate{
			{
				PubKey: tmTypes.PubKey{},
				Power:  1,
			},
		},
		AppStateBytes: jsonState,
	})

	return app
}

// SendCommit sends Commit message to given Blockchain instance
func SendCommit(app *noah.Blockchain) tmTypes.ResponseCommit {
	return app.Commit()
}

// SendBeginBlock sends BeginBlock message to given Blockchain instance
func SendBeginBlock(app *noah.Blockchain) tmTypes.ResponseBeginBlock {
	return app.BeginBlock(tmTypes.RequestBeginBlock{
		Hash: nil,
		Header: tmTypes.Header{
			Version:            tmTypes.Version{},
			ChainID:            "",
			Height:             1,
			Time:               time.Time{},
			LastBlockId:        tmTypes.BlockID{},
			LastCommitHash:     nil,
			DataHash:           nil,
			ValidatorsHash:     nil,
			NextValidatorsHash: nil,
			ConsensusHash:      nil,
			AppHash:            nil,
			LastResultsHash:    nil,
			EvidenceHash:       nil,
			ProposerAddress:    nil,
		},
		LastCommitInfo: tmTypes.LastCommitInfo{
			Round: 0,
			Votes: nil,
		},
		ByzantineValidators: nil,
	})
}

// SendEndBlock sends EndBlock message to given Blockchain instance
func SendEndBlock(app *noah.Blockchain) tmTypes.ResponseEndBlock {
	return app.EndBlock(tmTypes.RequestEndBlock{
		Height: 0,
	})
}

// CreateTx composes and returns Tx with given params.
// Nonce, chain id, gas price, gas coin and signature type fields are auto-filled.
func CreateTx(app *noah.Blockchain, address types.Address, txType transaction.TxType, data interface{}) transaction.Transaction {
	nonce := app.CurrentState().Accounts().GetNonce(address) + 1
	bData, err := rlp.EncodeToBytes(data)
	if err != nil {
		panic(err)
	}

	tx := transaction.Transaction{
		Nonce:         nonce,
		ChainID:       types.CurrentChainID,
		GasPrice:      1,
		GasCoin:       types.GetBaseCoinID(),
		Type:          txType,
		Data:          bData,
		SignatureType: transaction.SigTypeSingle,
	}

	return tx
}

// SendTx sends DeliverTx message to given Blockchain instance
func SendTx(app *noah.Blockchain, bytes []byte) tmTypes.ResponseDeliverTx {
	return app.DeliverTx(tmTypes.RequestDeliverTx{
		Tx: bytes,
	})
}

// SignTx returns bytes of signed with given pk transaction
func SignTx(pk *ecdsa.PrivateKey, tx transaction.Transaction) []byte {
	err := tx.Sign(pk)
	if err != nil {
		panic(err)
	}

	b, _ := rlp.EncodeToBytes(tx)

	return b
}

// CreateAddress returns random address and corresponding private key
func CreateAddress() (types.Address, *ecdsa.PrivateKey) {
	pk, _ := crypto.GenerateKey()

	return crypto.PubkeyToAddress(pk.PublicKey), pk
}

// DefaultAppState returns new AppState with some predefined values
func DefaultAppState() types.AppState {
	return types.AppState{
		Note:                "",
		StartHeight:         1,
		Validators:          nil,
		Candidates:          nil,
		BlockListCandidates: nil,
		Accounts:            nil,
		Coins:               nil,
		FrozenFunds:         nil,
		HaltBlocks:          nil,
		UsedChecks:          nil,
		MaxGas:              0,
		TotalSlashed:        "0",
	}
}
