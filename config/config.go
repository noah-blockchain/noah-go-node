package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	tmConfig "github.com/tendermint/tendermint/config"
)

const (
	// LogFormatPlain is a format for colored text
	LogFormatPlain = "plain"
	// LogFormatJSON is a format for json output
	LogFormatJSON = "json"

	defaultConfigFileName  = "config.toml"
	defaultGenesisJSONName = "genesis.json"

	defaultPrivValName      = "priv_validator.json"
	defaultPrivValStateName = "priv_validator_state.json"
	defaultNodeKeyName      = "node_key.json"
)

var (
	NetworkId        string
	DefaultNetworkId = "noah-mainnet-1"

	ValidatorMode bool

	defaultConfigDir = "config"
	defaultDataDir   = "data"

	defaultConfigFilePath   = filepath.Join(defaultConfigDir, defaultConfigFileName)
	defaultGenesisJSONPath  = filepath.Join(defaultConfigDir, defaultGenesisJSONName)
	defaultPrivValKeyPath   = filepath.Join(defaultConfigDir, defaultPrivValName)
	defaultPrivValStatePath = filepath.Join(defaultConfigDir, defaultPrivValStateName)
	defaultNodeKeyPath      = filepath.Join(defaultConfigDir, defaultNodeKeyName)
)

func DefaultConfig() *Config {
	cfg := defaultConfig()

	cfg.P2P.Seeds = "eb7acbf988f2183b487c9a1ee69f85050d5aa3a8@1.testnet.noah-blockchain.com:26656," +
		"a4bbc9c38ec2cb73850109465579ed9e2c445a53@2.testnet.noah-blockchain.com:26656," +
		"672e70fcbf0284baff0082851826c8aa37a35fb3@3.testnet.noah-blockchain.com:26656," +
		"5a3eff103ade054d6b90b963c6a5990bed75336c@4.testnet.noah-blockchain.com:26656," +
		"49055a20a4ac0992bd492d485efe998f6a8869b1@5.testnet.noah-blockchain.com:26656," +
		"2c72b7408d44821de67d76daf83f62f5d65c0e7c@6.testnet.noah-blockchain.com:26656," +
		"a5b5f0296799d0a30909d1f0066355ff72808acf@7.testnet.noah-blockchain.com:26656," +
		"4a0c6ab31de82ee1988e339cc6efbf807d35d10e@8.testnet.noah-blockchain.com:26656"

	cfg.P2P.PersistentPeers = "eb7acbf988f2183b487c9a1ee69f85050d5aa3a8@1.testnet.noah-blockchain.com:26656," +
		"a4bbc9c38ec2cb73850109465579ed9e2c445a53@2.testnet.noah-blockchain.com:26656," +
		"672e70fcbf0284baff0082851826c8aa37a35fb3@3.testnet.noah-blockchain.com:26656," +
		"5a3eff103ade054d6b90b963c6a5990bed75336c@4.testnet.noah-blockchain.com:26656," +
		"49055a20a4ac0992bd492d485efe998f6a8869b1@5.testnet.noah-blockchain.com:26656," +
		"2c72b7408d44821de67d76daf83f62f5d65c0e7c@6.testnet.noah-blockchain.com:26656," +
		"a5b5f0296799d0a30909d1f0066355ff72808acf@7.testnet.noah-blockchain.com:26656," +
		"4a0c6ab31de82ee1988e339cc6efbf807d35d10e@8.testnet.noah-blockchain.com:26656"

	if NetworkId == "noah-mainnet-1" {
		cfg.P2P.Seeds = "a3beb81d555e181933b335b56b9967db353723ba@mainnet1.noah-blockchain.com:26656," +
			"970449c0c2a3218e8e1a94a5e43076249fa54c81@mainnet2.noah-blockchain.com:26656," +
			"82af4dcfab461fb2923d3c8df1640ac794765602@mainnet3.noah-blockchain.com:26656," +
			"373c7f34bde64a400efaf120d6329a12eee7f8f0@mainnet4.noah-blockchain.com:26656," +
			"6d9868e1f43988dbc192a5d652a58eae3d5fbafb@mainnet5.noah-blockchain.com:26656," +
			"bda470075f5bafbda3ba95a0343ea786a59dc2b3@mainnet6.noah-blockchain.com:26656," +
			"2eb583880f6d35e4c20558a18ab7d7ad5a3cf515@mainnet7.noah-blockchain.com:26656," +
			"4b77488f8010d5020ca1a26cb910c4e49b0bbbd2@mainnet8.noah-blockchain.com:26656," +
			"fe29718ffd0e9b3e6165c793cb920630af361b94@mainnet9.noah-blockchain.com:26656"

		cfg.P2P.PersistentPeers = "a3beb81d555e181933b335b56b9967db353723ba@mainnet1.noah-blockchain.com:26656," +
			"970449c0c2a3218e8e1a94a5e43076249fa54c81@mainnet2.noah-blockchain.com:26656," +
			"82af4dcfab461fb2923d3c8df1640ac794765602@mainnet3.noah-blockchain.com:26656," +
			"373c7f34bde64a400efaf120d6329a12eee7f8f0@mainnet4.noah-blockchain.com:26656," +
			"6d9868e1f43988dbc192a5d652a58eae3d5fbafb@mainnet5.noah-blockchain.com:26656," +
			"bda470075f5bafbda3ba95a0343ea786a59dc2b3@mainnet6.noah-blockchain.com:26656," +
			"2eb583880f6d35e4c20558a18ab7d7ad5a3cf515@mainnet7.noah-blockchain.com:26656," +
			"4b77488f8010d5020ca1a26cb910c4e49b0bbbd2@mainnet8.noah-blockchain.com:26656," +
			"fe29718ffd0e9b3e6165c793cb920630af361b94@mainnet9.noah-blockchain.com:26656"
	}

	cfg.TxIndex = &tmConfig.TxIndexConfig{
		Indexer:      "kv",
		IndexTags:    "",
		IndexAllTags: true,
	}

	cfg.DBPath = "tmdata"

	cfg.Mempool.CacheSize = 100000
	cfg.Mempool.Recheck = false
	cfg.Mempool.Size = 10000

	cfg.Consensus.WalPath = "tmdata/cs.wal/wal"
	cfg.Consensus.TimeoutPropose = 2 * time.Second
	cfg.Consensus.TimeoutProposeDelta = 500 * time.Millisecond
	cfg.Consensus.TimeoutPrevote = 1 * time.Second
	cfg.Consensus.TimeoutPrevoteDelta = 500 * time.Millisecond
	cfg.Consensus.TimeoutPrecommit = 1 * time.Second
	cfg.Consensus.TimeoutPrecommitDelta = 500 * time.Millisecond
	cfg.Consensus.TimeoutCommit = 4500 * time.Millisecond

	cfg.P2P.RecvRate = 15360000 // 15 mB/s
	cfg.P2P.SendRate = 15360000 // 15 mB/s
	cfg.P2P.FlushThrottleTimeout = 10 * time.Millisecond

	cfg.PrivValidatorKey = "config/priv_validator.json"
	cfg.PrivValidatorState = "config/priv_validator_state.json"
	cfg.NodeKey = "config/node_key.json"

	return cfg
}

func GetConfig() *Config {
	cfg := DefaultConfig()

	if cfg.ValidatorMode {
		cfg.TxIndex.IndexAllTags = false
		cfg.TxIndex.IndexTags = ""

		cfg.RPC.ListenAddress = ""
		cfg.RPC.GRPCListenAddress = ""
	}

	cfg.Mempool.Recheck = false

	cfg.P2P.AddrBook = fmt.Sprintf("config/addrbook-%s.json", NetworkId)

	cfg.SetRoot(utils.GetNoahHome())
	EnsureRoot(utils.GetNoahHome())

	return cfg
}

// Config defines the top level configuration for a Tendermint node
type Config struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	// Options for services
	RPC             *tmConfig.RPCConfig             `mapstructure:"rpc"`
	P2P             *tmConfig.P2PConfig             `mapstructure:"p2p"`
	Mempool         *tmConfig.MempoolConfig         `mapstructure:"mempool"`
	Consensus       *tmConfig.ConsensusConfig       `mapstructure:"consensus"`
	FastSyncSection *tmConfig.FastSyncConfig        `mapstructure:"fastsync"`
	TxIndex         *tmConfig.TxIndexConfig         `mapstructure:"tx_index"`
	Instrumentation *tmConfig.InstrumentationConfig `mapstructure:"instrumentation"`
}

// DefaultConfig returns a default configuration for a Tendermint node
func defaultConfig() *Config {
	return &Config{
		BaseConfig:      DefaultBaseConfig(),
		RPC:             tmConfig.DefaultRPCConfig(),
		P2P:             tmConfig.DefaultP2PConfig(),
		Mempool:         tmConfig.DefaultMempoolConfig(),
		Consensus:       tmConfig.DefaultConsensusConfig(),
		FastSyncSection: tmConfig.DefaultFastSyncConfig(),
		TxIndex:         tmConfig.DefaultTxIndexConfig(),
		Instrumentation: tmConfig.DefaultInstrumentationConfig(),
	}
}

// SetRoot sets the RootDir for all Config structs
func (cfg *Config) SetRoot(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.Mempool.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}

func GetTmConfig(cfg *Config) *tmConfig.Config {
	return &tmConfig.Config{
		BaseConfig: tmConfig.BaseConfig{
			RootDir:                 cfg.RootDir,
			Genesis:                 cfg.Genesis,
			PrivValidatorKey:        cfg.PrivValidatorKey,
			PrivValidatorState:      cfg.PrivValidatorState,
			NodeKey:                 cfg.NodeKey,
			Moniker:                 cfg.Moniker,
			PrivValidatorListenAddr: cfg.PrivValidatorListenAddr,
			ProxyApp:                cfg.ProxyApp,
			ABCI:                    cfg.ABCI,
			LogLevel:                cfg.LogLevel,
			LogFormat:               cfg.LogFormat,
			ProfListenAddress:       cfg.ProfListenAddress,
			FastSyncMode:            cfg.FastSync,
			FilterPeers:             cfg.FilterPeers,
			DBBackend:               cfg.DBBackend,
			DBPath:                  cfg.DBPath,
		},
		RPC:             cfg.RPC,
		P2P:             cfg.P2P,
		Mempool:         cfg.Mempool,
		Consensus:       cfg.Consensus,
		FastSync:        cfg.FastSyncSection,
		TxIndex:         cfg.TxIndex,
		Instrumentation: cfg.Instrumentation,
	}
}

//-----------------------------------------------------------------------------
// BaseConfig

// BaseConfig defines the base configuration for a Tendermint node
type BaseConfig struct {
	// chainID is unexposed and immutable but here for convenience
	chainID string

	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`

	// Path to the JSON file containing the initial validator set and other meta data
	Genesis string `mapstructure:"genesis_file"`

	// Path to the JSON file containing the private key to use as a validator in the consensus protocol
	PrivValidatorKey string `mapstructure:"priv_validator_key_file"`

	// Path to the JSON file containing the last sign state of a validator
	PrivValidatorState string `mapstructure:"priv_validator_state_file"`

	// TCP or UNIX socket address for Tendermint to listen on for
	// connections from an external PrivValidator process
	PrivValidatorListenAddr string `mapstructure:"priv_validator_laddr"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `mapstructure:"node_key_file"`

	// A custom human readable name for this node
	Moniker string `mapstructure:"moniker"`

	// TCP or UNIX socket address of the ABCI application,
	// or the name of an ABCI application compiled in with the Tendermint binary
	ProxyApp string `mapstructure:"proxy_app"`

	// Mechanism to connect to the ABCI application: socket | grpc
	ABCI string `mapstructure:"abci"`

	// Output level for logging
	LogLevel string `mapstructure:"log_level"`

	// Output format: 'plain' (colored text) or 'json'
	LogFormat string `mapstructure:"log_format"`

	// TCP or UNIX socket address for the profiling server to listen on
	ProfListenAddress string `mapstructure:"prof_laddr"`

	// If this node is many blocks behind the tip of the chain, FastSync
	// allows them to catchup quickly by downloading blocks in parallel
	// and verifying their commits
	FastSync bool `mapstructure:"fast_sync"`

	// If true, query the ABCI app on connecting to a new peer
	// so the app can decide if we should keep the connection or not
	FilterPeers bool `mapstructure:"filter_peers"` // false

	// Database backend: leveldb | memdb
	DBBackend string `mapstructure:"db_backend"`

	// Database directory
	DBPath string `mapstructure:"db_dir"`

	// Address to listen for GUI connections
	GUIListenAddress string `mapstructure:"gui_listen_addr"`

	// Address to listen for API connections
	APIListenAddress string `mapstructure:"api_listen_addr"`

	ValidatorMode bool `mapstructure:"validator_mode"`

	KeepStateHistory bool `mapstructure:"keep_state_history"`

	APISimultaneousRequests int `mapstructure:"api_simultaneous_requests"`

	LogPath string `mapstructure:"log_path"`
}

// DefaultBaseConfig returns a default base configuration for a Tendermint node
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		Genesis:                 defaultGenesisJSONPath,
		PrivValidatorKey:        defaultPrivValKeyPath,
		PrivValidatorState:      defaultPrivValStatePath,
		NodeKey:                 defaultNodeKeyPath,
		Moniker:                 defaultMoniker,
		LogLevel:                DefaultPackageLogLevels(),
		ProfListenAddress:       "",
		FastSync:                true,
		FilterPeers:             false,
		DBBackend:               "goleveldb",
		DBPath:                  "data",
		GUIListenAddress:        ":3000",
		APIListenAddress:        "tcp://0.0.0.0:8841",
		ValidatorMode:           true,
		KeepStateHistory:        false,
		APISimultaneousRequests: 100,
		LogPath:                 "stdout",
		LogFormat:               LogFormatPlain,
	}
}

func (cfg BaseConfig) ChainID() string {
	return cfg.chainID
}

// GenesisFile returns the full path to the genesis.json file
func (cfg BaseConfig) GenesisFile() string {
	return rootify(cfg.Genesis, cfg.RootDir)
}

// PrivValidatorFile returns the full path to the priv_validator.json file
func (cfg BaseConfig) PrivValidatorStateFile() string {
	return rootify(cfg.PrivValidatorState, cfg.RootDir)
}

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return rootify(cfg.NodeKey, cfg.RootDir)
}

func (cfg BaseConfig) PrivValidatorKeyFile() string {
	return rootify(cfg.PrivValidatorKey, cfg.RootDir)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	return rootify(cfg.DBPath, cfg.RootDir)
}

// DefaultLogLevel returns a default log level of "error"
func DefaultLogLevel() string {
	return "error"
}

// DefaultPackageLogLevels returns a default log level setting so all packages
// log at "error", while the `state` and `main` packages log at "info"
func DefaultPackageLogLevels() string {
	return fmt.Sprintf("consensus:info,main:info,blockchain:info,state:info,*:%s", DefaultLogLevel())
}

//-----------------------------------------------------------------------------
// Utils

// helper function to make config creation independent of root dir
func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

var defaultMoniker = getDefaultMoniker()

// getDefaultMoniker returns a default moniker, which is the host name. If runtime
// fails to get the host name, "anonymous" will be returned.
func getDefaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "anonymous"
	}
	return moniker
}
