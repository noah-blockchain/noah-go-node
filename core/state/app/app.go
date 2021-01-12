package app

import (
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/state/bus"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/rlp"
	"github.com/noah-blockchain/noah-go-node/tree"
	"math/big"
)

const mainPrefix = 'd'

type RApp interface {
	Export(state *types.AppState, height uint64)
	GetMaxGas() uint64
	GetTotalSlashed() *big.Int
	GetCoinsCount() uint32
	GetNextCoinID() types.CoinID
}

func (v *App) Tree() tree.ReadOnlyTree {
	return v.iavl
}

type App struct {
	model   *Model
	isDirty bool

	bus  *bus.Bus
	iavl tree.MTree
}

func NewApp(stateBus *bus.Bus, iavl tree.MTree) (*App, error) {
	app := &App{bus: stateBus, iavl: iavl}
	app.bus.SetApp(NewBus(app))

	return app, nil
}

func (v *App) Commit() error {
	if !v.isDirty {
		return nil
	}

	v.isDirty = false

	data, err := rlp.EncodeToBytes(v.model)
	if err != nil {
		return fmt.Errorf("can't encode legacyApp model: %s", err)
	}

	path := []byte{mainPrefix}
	v.iavl.Set(path, data)

	return nil
}

func (v *App) GetMaxGas() uint64 {
	model := v.getOrNew()

	return model.getMaxGas()
}

func (v *App) SetMaxGas(gas uint64) {
	model := v.getOrNew()
	model.setMaxGas(gas)
}

func (v *App) GetTotalSlashed() *big.Int {
	model := v.getOrNew()

	return model.getTotalSlashed()
}

func (v *App) AddTotalSlashed(amount *big.Int) {
	if amount.Cmp(big.NewInt(0)) == 0 {
		return
	}

	model := v.getOrNew()
	model.setTotalSlashed(big.NewInt(0).Add(model.getTotalSlashed(), amount))
	v.bus.Checker().AddCoin(types.GetBaseCoinID(), amount)
}

func (v *App) get() *Model {
	if v.model != nil {
		return v.model
	}

	path := []byte{mainPrefix}
	_, enc := v.iavl.Get(path)
	if len(enc) == 0 {
		return nil
	}

	model := &Model{}
	if err := rlp.DecodeBytes(enc, model); err != nil {
		panic(fmt.Sprintf("failed to decode legacyApp model at: %s", err))
	}

	v.model = model
	v.model.markDirty = v.markDirty
	return v.model
}

func (v *App) getOrNew() *Model {
	model := v.get()
	if model == nil {
		model = &Model{
			TotalSlashed: big.NewInt(0),
			CoinsCount:   0,
			MaxGas:       0,
			markDirty:    v.markDirty,
		}
		v.model = model
	}

	return model
}

func (v *App) markDirty() {
	v.isDirty = true
}

func (v *App) SetTotalSlashed(amount *big.Int) {
	v.getOrNew().setTotalSlashed(amount)
}

func (v *App) GetCoinsCount() uint32 {
	return v.getOrNew().getCoinsCount()
}

func (v *App) GetNextCoinID() types.CoinID {
	return types.CoinID(v.GetCoinsCount() + 1)
}

func (v *App) SetCoinsCount(count uint32) {
	v.getOrNew().setCoinsCount(count)
}

func (v *App) Export(state *types.AppState, height uint64) {
	state.MaxGas = v.GetMaxGas()
	state.TotalSlashed = v.GetTotalSlashed().String()
	state.StartHeight = height
}
