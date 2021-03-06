package config

import (
	"fmt"
	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	tmConfig "github.com/tendermint/tendermint/config"
	"os"
	"path/filepath"
	"time"
)

const (
	// LogFormatPlain is a format for colored text
	LogFormatPlain = "plain"
	// LogFormatJSON is a format for json output
	LogFormatJSON = "json"

	defaultConfigDir = "config"
	defaultDataDir   = "data"

	defaultConfigFileName  = "config.toml"
	defaultGenesisJSONName = "genesis.json"

	defaultPrivValName      = "priv_validator.json"
	defaultPrivValStateName = "priv_validator_state.json"
	defaultNodeKeyName      = "node_key.json"
)

var (
	ChainId                 string
	NetworkId               string
	defaultConfigFilePath   = filepath.Join(defaultConfigDir, defaultConfigFileName)
	defaultGenesisJSONPath  = filepath.Join(defaultConfigDir, defaultGenesisJSONName)
	defaultPrivValKeyPath   = filepath.Join(defaultConfigDir, defaultPrivValName)
	defaultPrivValStatePath = filepath.Join(defaultConfigDir, defaultPrivValStateName)
	defaultNodeKeyPath      = filepath.Join(defaultConfigDir, defaultNodeKeyName)
)

func DefaultConfig() *Config {
	cfg := defaultConfig()

	if ChainId == "testnet" {
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
	} else {
		cfg.P2P.Seeds = "708a9732b65562e16030cb176ac739ddc7cb2160@mainnet1.noah-blockchain.com:26656," +
			"778c4af8ebbc907dc7ff34dd1aad514fd28fe188@mainnet2.noah-blockchain.com:26656," +
			"eb5c0cc7259ebde8fbaef3028ea65b5c84c53098@mainnet3.noah-blockchain.com:26656," +
			"f95f52174e6f23ee45e5806287b6e9a07cfdd6a3@mainnet4.noah-blockchain.com:26656"

		cfg.P2P.PersistentPeers = "708a9732b65562e16030cb176ac739ddc7cb2160@mainnet1.noah-blockchain.com:26656," +
			"778c4af8ebbc907dc7ff34dd1aad514fd28fe188@mainnet2.noah-blockchain.com:26656," +
			"eb5c0cc7259ebde8fbaef3028ea65b5c84c53098@mainnet3.noah-blockchain.com:26656," +
			"f95f52174e6f23ee45e5806287b6e9a07cfdd6a3@mainnet4.noah-blockchain.com:26656"
	}

	cfg.TxIndex = &tmConfig.TxIndexConfig{
		Indexer:      "kv",
		IndexKeys:    "",
		IndexAllKeys: true,
	}

	cfg.DBPath = "tmdata"

	cfg.Mempool.CacheSize = 100000
	cfg.Mempool.Recheck = false
	cfg.Mempool.Size = 10000

	cfg.Consensus.WalPath = "tmdata/cs.wal/wal"
	cfg.Consensus.TimeoutPropose = 2 * time.Second               // timeout_propose = how long we wait for a proposal block before prevoting nil
	cfg.Consensus.TimeoutProposeDelta = 500 * time.Millisecond   // timeout_propose_delta = how much timeout_propose increases with each round
	cfg.Consensus.TimeoutPrevote = 1 * time.Second               // timeout_prevote = how long we wait after receiving +2/3 prevotes for "anything" (ie. not a single block or nil)
	cfg.Consensus.TimeoutPrevoteDelta = 500 * time.Millisecond   // timeout_prevote_delta = how much the timeout_prevote increases with each round
	cfg.Consensus.TimeoutPrecommit = 1 * time.Second             // timeout_precommit = how long we wait after receiving +2/3 precommits for "anything" (ie. not a single block or nil)
	cfg.Consensus.TimeoutPrecommitDelta = 500 * time.Millisecond // timeout_precommit_delta = how much the timeout_precommit increases with each round
	cfg.Consensus.TimeoutCommit = 4200 * time.Millisecond        // timeout_commit = how long we wait after committing a block, before starting on the new height (this gives us a chance to receive some more precommits, even though we already have +2/3)

	cfg.P2P.RecvRate = 15360000 // 15 mB/s
	cfg.P2P.SendRate = 15360000 // 15 mB/s
	cfg.P2P.FlushThrottleTimeout = 10 * time.Millisecond

	cfg.PrivValidatorKey = "config/priv_validator.json"
	cfg.PrivValidatorState = "config/priv_validator_state.json"
	cfg.NodeKey = "config/node_key.json"

	return cfg
}

// GetConfig returns DefaultConfig with some changes
func GetConfig() *Config {
	cfg := DefaultConfig()

	if cfg.ValidatorMode {
		cfg.TxIndex.IndexAllKeys = false
		cfg.TxIndex.IndexKeys = ""

		cfg.RPC.ListenAddress = ""
		cfg.RPC.GRPCListenAddress = ""
	}

	cfg.Mempool.Recheck = false

	cfg.P2P.AddrBook = "config/addrbook.json"

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

// GetTmConfig composes and returns config for Tendermint engine based on given Noah config
func GetTmConfig(cfg *Config) *tmConfig.Config {
	return &tmConfig.Config{
		BaseConfig: tmConfig.BaseConfig{
			RootDir:                 cfg.RootDir,
			ProxyApp:                cfg.ProxyApp,
			Moniker:                 cfg.Moniker,
			FastSyncMode:            cfg.FastSync,
			DBBackend:               cfg.DBBackend,
			DBPath:                  cfg.DBPath,
			LogLevel:                cfg.LogLevel,
			LogFormat:               cfg.LogFormat,
			Genesis:                 cfg.Genesis,
			PrivValidatorKey:        cfg.PrivValidatorKey,
			PrivValidatorState:      cfg.PrivValidatorState,
			PrivValidatorListenAddr: cfg.PrivValidatorListenAddr,
			NodeKey:                 cfg.NodeKey,
			ABCI:                    cfg.ABCI,
			ProfListenAddress:       cfg.ProfListenAddress,
			FilterPeers:             cfg.FilterPeers,
		},
		RPC:             cfg.RPC,
		P2P:             cfg.P2P,
		Mempool:         cfg.Mempool,
		FastSync:        &tmConfig.FastSyncConfig{Version: "v0"},
		Consensus:       cfg.Consensus,
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

	// Address to listen for API connections
	APIListenAddress string `mapstructure:"api_listen_addr"`

	// Address to listen for gRPC connections
	GRPCListenAddress string `mapstructure:"grpc_listen_addr"`

	// Address to listen for API v2 connections
	APIv2ListenAddress string `mapstructure:"api_v2_listen_addr"`

	// API v2 Timeout
	APIv2TimeoutDuration time.Duration `mapstructure:"api_v2_timeout_duration"`

	ValidatorMode bool `mapstructure:"validator_mode"`

	KeepLastStates int64 `mapstructure:"keep_last_states"`

	APISimultaneousRequests int `mapstructure:"api_simultaneous_requests"`

	LogPath string `mapstructure:"log_path"`

	StateCacheSize int `mapstructure:"state_cache_size"`

	StateMemAvailable int `mapstructure:"state_mem_available"`

	HaltHeight int `mapstructure:"halt_height"`
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
		APIListenAddress:        "tcp://0.0.0.0:8841",
		GRPCListenAddress:       "tcp://0.0.0.0:8842",
		APIv2ListenAddress:      "tcp://0.0.0.0:8843",
		APIv2TimeoutDuration:    10 * time.Second,
		ValidatorMode:           false,
		KeepLastStates:          120,
		StateCacheSize:          1000000,
		StateMemAvailable:       1024,
		APISimultaneousRequests: 100,
		LogPath:                 "stdout",
		LogFormat:               LogFormatPlain,
	}
}

// ChainID returns the id of a chain
func (cfg BaseConfig) ChainID() string {
	return cfg.chainID
}

// GenesisFile returns the full path to the genesis.json file
func (cfg BaseConfig) GenesisFile() string {
	return rootify(cfg.Genesis, cfg.RootDir)
}

// PrivValidatorStateFile returns the full path to the priv_validator_state.json file
func (cfg BaseConfig) PrivValidatorStateFile() string {
	return rootify(cfg.PrivValidatorState, cfg.RootDir)
}

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return rootify(cfg.NodeKey, cfg.RootDir)
}

// PrivValidatorKeyFile returns the full path to the priv_validator.json file
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
	return fmt.Sprintf("consensus:info,main:info,state:info,*:%s", DefaultLogLevel())
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
