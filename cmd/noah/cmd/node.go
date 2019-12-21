package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/noah-blockchain/noah-go-node/api"
	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	"github.com/noah-blockchain/noah-go-node/config"
	"github.com/noah-blockchain/noah-go-node/core/noah"
	"github.com/noah-blockchain/noah-go-node/eventsdb"
	"github.com/noah-blockchain/noah-go-node/gui"
	"github.com/noah-blockchain/noah-go-node/log"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/abci/types"
	tmCfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/common"
	tmNode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	rpc "github.com/tendermint/tendermint/rpc/client"
	bc "github.com/tendermint/tendermint/store"
	tmTypes "github.com/tendermint/tendermint/types"
)

var RunNode = &cobra.Command{
	Use:   "node",
	Short: "Run the Noah node",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNode()
	},
}

func runNode() error {
	now := time.Now()
	startTime := time.Date(2019, time.June, 5, 17, 0, 0, 0, time.UTC)
	if startTime.After(now) {
		fmt.Printf("Start time is in the future, sleeping until %s", startTime)
		time.Sleep(startTime.Sub(now))
	}

	tmConfig := config.GetTmConfig(cfg)

	if err := common.EnsureDir(fmt.Sprintf("%s/config-%s", utils.GetNoahHome(), config.NetworkId), 0777); err != nil {
		return err
	}

	if err := common.EnsureDir(fmt.Sprintf("%s/tmdata-%s", utils.GetNoahHome(), config.NetworkId), 0777); err != nil {
		return err
	}

	eventsdb.InitDB(cfg)

	app := noah.NewNoahBlockchain(cfg)
	// update BlocksTimeDelta in case it was corrupted
	updateBlocksTimeDelta(app, tmConfig)

	// start TM node
	node := startTendermintNode(app, tmConfig)

	client := rpc.NewLocal(node)
	status, _ := client.Status()
	if status.NodeInfo.Network != config.NetworkId {
		log.Fatal("Different networks", "expected", config.NetworkId, "got", status.NodeInfo.Network)
	}

	app.SetTmNode(node)

	if !cfg.ValidatorMode {
		go api.RunAPI(app, client, cfg)
		go gui.Run(cfg.GUIListenAddress)
	}

	fmt.Println("Noah node successful started.")

	// Recheck mempool. Currently kind a hack.
	go recheckMempool(node, cfg)

	common.TrapSignal(log.With("module", "trap"), func() {
		// Cleanup
		err := node.Stop()
		app.Stop()
		if err != nil {
			panic(err)
		}
	})

	// Run forever
	select {}
}

func recheckMempool(node *tmNode.Node, config *config.Config) {
	ticker := time.NewTicker(time.Minute)
	mempool := node.Mempool()
	for {
		select {
		case <-ticker.C:
			txs := mempool.ReapMaxTxs(config.Mempool.Size)
			mempool.Flush()

			for _, tx := range txs {
				_ = mempool.CheckTx(tx, func(res *types.Response) {})
			}
		}
	}
}

func updateBlocksTimeDelta(app *noah.Blockchain, config *tmCfg.Config) {
	blockStoreDB, err := tmNode.DefaultDBProvider(&tmNode.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		panic(err)
	}

	blockStore := bc.NewBlockStore(blockStoreDB)
	height := uint64(blockStore.Height())
	count := uint64(3)
	if _, err := app.GetBlocksTimeDelta(height, count); height >= 20 && err != nil {
		blockA := blockStore.LoadBlockMeta(int64(height - count - 1))
		blockB := blockStore.LoadBlockMeta(int64(height - 1))

		delta := int(blockB.Header.Time.Sub(blockA.Header.Time).Seconds())
		app.SetBlocksTimeDelta(height, delta)
	}
	blockStoreDB.Close()
}

func getNodeKey() (*p2p.NodeKey, error) {
	nodeKeyJSON := config.GetEnv("NODE_KEY", "")
	if len(nodeKeyJSON) > 0 {
		if err := ioutil.WriteFile(cfg.NodeKeyFile(), []byte(nodeKeyJSON), 0600); err != nil {
			return nil, err
		}
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	return nodeKey, nil
}

func getValidatorKey() (*privval.FilePV, error) {
	validatorKeyJSON := config.GetEnv("VALIDATOR_KEY", "")
	if len(validatorKeyJSON) > 0 {
		if err := ioutil.WriteFile(cfg.PrivValidatorKeyFile(), []byte(validatorKeyJSON), 0600); err != nil {
			return nil, err
		}

		if err := ioutil.WriteFile(cfg.PrivValidatorStateFile(), []byte("{}"), 0600); err != nil {
			return nil, err
		}
	}

	var pv *privval.FilePV
	if common.FileExists(cfg.PrivValidatorKeyFile()) {
		pv = privval.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	} else {
		pv = privval.GenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
		pv.Save()
	}
	return pv, nil
}

func startTendermintNode(app types.Application, cfg *tmCfg.Config) *tmNode.Node {
	nodeKey, err := getNodeKey()
	if err != nil {
		panic(err)
	}

	validatorKey, err := getValidatorKey()
	if err != nil {
		panic(err)
	}

	node, err := tmNode.NewNode(
		cfg,
		validatorKey,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		getGenesis,
		tmNode.DefaultDBProvider,
		tmNode.DefaultMetricsProvider(cfg.Instrumentation),
		log.With("module", "tendermint"),
	)

	if err != nil {
		log.Fatal("failed to create a node", "err", err)
	}

	if err = node.Start(); err != nil {
		log.Fatal("failed to start node", "err", err)
	}

	log.Info("Started node", "nodeInfo", node.Switch().NodeInfo())

	return node
}

func getGenesis() (doc *tmTypes.GenesisDoc, e error) {
	genesisFile := fmt.Sprintf("%s/config-%s/genesis.json", utils.GetNoahHome(), config.NetworkId)

	if !common.FileExists(genesisFile) {
		rootDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		input, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s/genesis.json", rootDir, config.ChainId, config.NetworkId))
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(genesisFile, input, 0644)
		if err != nil {
			fmt.Println("Error creating", genesisFile)
			panic(err)
		}
	}

	return tmTypes.GenesisDocFromFile(genesisFile)
}
